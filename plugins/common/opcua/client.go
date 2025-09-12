package opcua

import (
	"context"
	"fmt"
	"log" //nolint:depguard // just for debug
	"net/url"
	"strconv"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

type OpcUAWorkarounds struct {
	AdditionalValidStatusCodes []string `toml:"additional_valid_status_codes"`
}

type ConnectionState opcua.ConnState

const (
	Closed       ConnectionState = ConnectionState(opcua.Closed)
	Connected    ConnectionState = ConnectionState(opcua.Connected)
	Connecting   ConnectionState = ConnectionState(opcua.Connecting)
	Disconnected ConnectionState = ConnectionState(opcua.Disconnected)
	Reconnecting ConnectionState = ConnectionState(opcua.Reconnecting)
)

func (c ConnectionState) String() string {
	return opcua.ConnState(c).String()
}

type OpcUAClientConfig struct {
	Endpoint       string          `toml:"endpoint"`
	SecurityPolicy string          `toml:"security_policy"`
	SecurityMode   string          `toml:"security_mode"`
	Certificate    string          `toml:"certificate"`
	PrivateKey     string          `toml:"private_key"`
	Username       config.Secret   `toml:"username"`
	Password       config.Secret   `toml:"password"`
	AuthMethod     string          `toml:"auth_method"`
	ConnectTimeout config.Duration `toml:"connect_timeout"`
	RequestTimeout config.Duration `toml:"request_timeout"`
	ClientTrace    bool            `toml:"client_trace"`

	OptionalFields []string         `toml:"optional_fields"`
	Workarounds    OpcUAWorkarounds `toml:"workarounds"`
	SessionTimeout config.Duration  `toml:"session_timeout"`
}

func (o *OpcUAClientConfig) Validate() error {
	validationErrors := &MultiValidationError{}

	// Validate endpoint
	if err := ValidateEndpoint(o.Endpoint); err != nil {
		validationErrors.Add(err)
	}

	// Validate security policy
	validSecurityPolicies := []string{"None", "Basic128Rsa15", "Basic256", "Basic256Sha256", "auto", ""}
	if err := ValidateChoice("security_policy", o.SecurityPolicy, validSecurityPolicies); err != nil {
		validationErrors.Add(err)
	}

	// Validate security mode
	validSecurityModes := []string{"None", "Sign", "SignAndEncrypt", "auto", ""}
	if err := ValidateChoice("security_mode", o.SecurityMode, validSecurityModes); err != nil {
		validationErrors.Add(err)
	}

	// Validate authentication method
	validAuthMethods := []string{"Anonymous", "UserName", "Certificate", "auto", ""}
	if err := ValidateChoice("auth_method", o.AuthMethod, validAuthMethods); err != nil {
		validationErrors.Add(err)
	}

	// Validate certificate configuration
	if err := o.validateCertificateConfiguration(); err != nil {
		validationErrors.Add(err)
	}

	// Validate credentials based on auth method
	if err := o.validateCredentials(); err != nil {
		validationErrors.Add(err)
	}

	// Validate optional fields
	if err := ValidateOptionalFields(o.OptionalFields); err != nil {
		validationErrors.Add(err)
	}

	// Validate timeouts
	if err := o.validateTimeouts(); err != nil {
		validationErrors.Add(err)
	}

	return validationErrors.ErrorOrNil()
}

func (o *OpcUAClientConfig) validateCertificateConfiguration() error {
	// If using None/None security, certificates are optional
	if o.SecurityPolicy == "None" && o.SecurityMode == "None" {
		return nil
	}

	// Both empty is valid (will generate self-signed)
	if o.Certificate == "" && o.PrivateKey == "" {
		return nil
	}

	// Both must be provided if one is provided
	if o.Certificate == "" {
		return &ValidationError{
			Field:   "certificate",
			Message: "private key provided without certificate",
			Err:     ErrInvalidConfiguration,
		}
	}
	if o.PrivateKey == "" {
		return &ValidationError{
			Field:   "private_key",
			Message: "certificate provided without private key",
			Err:     ErrInvalidConfiguration,
		}
	}

	return nil
}

func (o *OpcUAClientConfig) validateCredentials() error {
	// If auth method is UserName, both username and password should be provided
	if o.AuthMethod == "UserName" {
		if err := o.validateSecret(o.Username, "username", "username required for UserName authentication"); err != nil {
			return err
		}

		if err := o.validateSecret(o.Password, "password", "password required for UserName authentication"); err != nil {
			return err
		}
	}

	return nil
}

func (*OpcUAClientConfig) validateSecret(secret config.Secret, fieldName, message string) error {
	secretStr, err := secret.Get()
	if secretStr != nil {
		defer secretStr.Destroy()
	}

	if err != nil || secretStr == nil || secretStr.Size() == 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: message,
			Err:     ErrInvalidConfiguration,
		}
	}

	return nil
}

func (o *OpcUAClientConfig) validateTimeouts() error {
	validationErrors := &MultiValidationError{}

	// Validate connect timeout
	if err := o.validateSingleTimeout("connect_timeout", o.ConnectTimeout, 100*time.Millisecond); err != nil {
		validationErrors.Add(err)
	}

	// Validate request timeout
	if err := o.validateSingleTimeout("request_timeout", o.RequestTimeout, 100*time.Millisecond); err != nil {
		validationErrors.Add(err)
	}

	// Validate session timeout
	if err := o.validateSingleTimeout("session_timeout", o.SessionTimeout, 1*time.Second); err != nil {
		validationErrors.Add(err)
	}

	return validationErrors.ErrorOrNil()
}

func (*OpcUAClientConfig) validateSingleTimeout(field string, timeout config.Duration, minValue time.Duration) error {
	if timeout < 0 {
		return &ValidationError{
			Field:   field,
			Value:   timeout,
			Message: fmt.Sprintf("must be non-negative, got %v", timeout),
			Err:     ErrInvalidConfiguration,
		}
	}

	if timeout != 0 && time.Duration(timeout) < minValue {
		return &ValidationError{
			Field:   field,
			Value:   timeout,
			Message: fmt.Sprintf("too short (%v), minimum recommended is %v", timeout, minValue),
			Err:     ErrInvalidConfiguration,
		}
	}

	return nil
}

func (o *OpcUAClientConfig) CreateClient(telegrafLogger telegraf.Logger) (*OpcUAClient, error) {
	err := o.Validate()
	if err != nil {
		return nil, err
	}

	if o.ClientTrace {
		debug.Enable = true
		debug.Logger = log.New(&DebugLogger{Log: telegrafLogger}, "", 0)
	}

	c := &OpcUAClient{
		Config: o,
		Log:    telegrafLogger,
	}
	c.Log.Debug("Initialising OpcUAClient")

	err = c.setupWorkarounds()
	return c, err
}

type OpcUAClient struct {
	Config *OpcUAClientConfig
	Log    telegraf.Logger

	Client *opcua.Client

	opts  []opcua.Option
	codes []ua.StatusCode
}

// SetupOptions reads the endpoints from the specified server and sets up all authentication
func (o *OpcUAClient) SetupOptions() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Config.ConnectTimeout))
	defer cancel()
	// Get a list of the endpoints for our target server
	endpoints, err := opcua.GetEndpoints(ctx, o.Config.Endpoint)
	if err != nil {
		return &EndpointError{
			Endpoint: o.Config.Endpoint,
			Err:      fmt.Errorf("failed to get endpoints: %w", err),
		}
	}

	if o.Config.Certificate == "" && o.Config.PrivateKey == "" {
		if o.Config.SecurityPolicy != "None" || o.Config.SecurityMode != "None" {
			o.Log.Debug("Generating self-signed certificate")
			cert, privateKey, err := generateCert("urn:telegraf:gopcua:client", 2048,
				o.Config.Certificate, o.Config.PrivateKey, 365*24*time.Hour)
			if err != nil {
				return err
			}

			o.Config.Certificate = cert
			o.Config.PrivateKey = privateKey
		}
	}

	o.Log.Debug("Configuring OPC UA connection options")
	o.opts, err = o.generateClientOpts(endpoints)

	return err
}

func (o *OpcUAClient) setupWorkarounds() error {
	o.codes = []ua.StatusCode{ua.StatusOK}
	for i, c := range o.Config.Workarounds.AdditionalValidStatusCodes {
		val, err := strconv.ParseUint(c, 0, 32) // setting 32 bits to allow for safe conversion
		if err != nil {
			return fmt.Errorf("%w: invalid status code %q at index %d: %w", ErrStatusCodeParsing, c, i, err)
		}
		o.codes = append(o.codes, ua.StatusCode(val))
	}

	return nil
}

func (o *OpcUAClient) StatusCodeOK(code ua.StatusCode) bool {
	for _, val := range o.codes {
		if val == code {
			return true
		}
	}
	return false
}

// Connect to an OPC UA device
func (o *OpcUAClient) Connect(ctx context.Context) error {
	o.Log.Debug("Connecting OPC UA Client to server")
	u, err := url.Parse(o.Config.Endpoint)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "opc.tcp":
		if err := o.SetupOptions(); err != nil {
			return err
		}

		if o.Client != nil {
			o.Log.Warnf("Closing connection to %q as already connected", u)
			if err := o.Client.Close(ctx); err != nil {
				// Only log the error but do not bail-out here as this prevents
				// reconnections for multiple parties (see e.g. #9523).
				o.Log.Errorf("Closing connection to %s failed: %v", o.Config.Endpoint, err)
			}
		}

		o.Client, err = opcua.NewClient(o.Config.Endpoint, o.opts...)
		if err != nil {
			return &EndpointError{
				Endpoint: o.Config.Endpoint,
				Err:      fmt.Errorf("%w: failed to create client: %w", ErrConnectionFailed, err),
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Config.ConnectTimeout))
		defer cancel()
		if err := o.Client.Connect(ctx); err != nil {
			return &EndpointError{
				Endpoint: o.Config.Endpoint,
				Err:      fmt.Errorf("%w: %w", ErrConnectionFailed, err),
			}
		}
		o.Log.Debug("Connected to OPC UA Server")

	default:
		return &EndpointError{
			Endpoint: o.Config.Endpoint,
			Err:      fmt.Errorf("%w: unsupported scheme %q, expected opc.tcp", ErrInvalidEndpoint, u.Scheme),
		}
	}
	return nil
}

func (o *OpcUAClient) Disconnect(ctx context.Context) error {
	o.Log.Debug("Disconnecting from OPC UA Server")
	u, err := url.Parse(o.Config.Endpoint)
	if err != nil {
		return &EndpointError{
			Endpoint: o.Config.Endpoint,
			Err:      fmt.Errorf("%w: %w", ErrInvalidEndpoint, err),
		}
	}

	switch u.Scheme {
	case "opc.tcp":
		// We can't do anything about failing to close a connection
		err := o.Client.Close(ctx)
		o.Client = nil
		if err != nil {
			return fmt.Errorf("failed to close connection to %s: %w", o.Config.Endpoint, err)
		}
		return nil
	default:
		return &EndpointError{
			Endpoint: o.Config.Endpoint,
			Err:      fmt.Errorf("%w: unsupported scheme %q", ErrInvalidEndpoint, u.Scheme),
		}
	}
}

func (o *OpcUAClient) State() ConnectionState {
	if o.Client == nil {
		return Disconnected
	}
	return ConnectionState(o.Client.State())
}
