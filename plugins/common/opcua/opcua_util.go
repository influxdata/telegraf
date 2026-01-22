package opcua

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf/config"
)

// SELF SIGNED CERT FUNCTIONS

func newTempDir() (string, error) {
	dir, err := os.MkdirTemp("", "ssc")
	return dir, err
}

func generateCert(host string, rsaBits int, certFile, keyFile string, dur time.Duration) (cert, key string, err error) {
	if len(host) == 0 {
		return "", "", &CertificateError{
			Operation: "validation",
			Err:       fmt.Errorf("%w: missing required host parameter", ErrCertificateGeneration),
		}
	}
	if rsaBits == 0 {
		rsaBits = 2048
	}

	// If both paths are empty, use temporary directory (backward compatible behavior)
	if certFile == "" && keyFile == "" {
		dir, err := newTempDir()
		if err != nil {
			return "", "", &CertificateError{
				Operation: "directory creation",
				Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
			}
		}
		certFile = dir + "/cert.pem"
		keyFile = dir + "/key.pem"
	} else {
		// If paths are provided, create parent directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(certFile), 0750); err != nil {
			return "", "", &CertificateError{
				Operation: "directory creation",
				Path:      certFile,
				Err:       fmt.Errorf("%w: failed to create parent directory: %w", ErrCertificateGeneration, err),
			}
		}

		if err := os.MkdirAll(filepath.Dir(keyFile), 0750); err != nil {
			return "", "", &CertificateError{
				Operation: "directory creation",
				Path:      keyFile,
				Err:       fmt.Errorf("%w: failed to create parent directory: %w", ErrCertificateGeneration, err),
			}
		}
	}

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return "", "", &CertificateError{
			Operation: "private key generation",
			Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
		}
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(dur)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", &CertificateError{
			Operation: "serial number generation",
			Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
		}
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Telegraf OPC UA Client"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage: x509.KeyUsageContentCommitment | x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
		if uri, err := url.Parse(h); err == nil {
			template.URIs = append(template.URIs, uri)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return "", "", &CertificateError{
			Operation: "certificate creation",
			Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
		}
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return "", "", &CertificateError{
			Operation: "file creation",
			Path:      certFile,
			Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
		}
	}
	defer func() {
		if closeErr := certOut.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			_ = closeErr // Acknowledge that we're intentionally ignoring this error
		}
	}()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", &CertificateError{
			Operation: "encoding",
			Path:      certFile,
			Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
		}
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", &CertificateError{
			Operation: "file creation",
			Path:      keyFile,
			Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
		}
	}
	defer func() {
		if closeErr := keyOut.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			_ = closeErr // Acknowledge that we're intentionally ignoring this error
		}
	}()

	keyBlock, err := pemBlockForKey(priv)
	if err != nil {
		return "", "", &CertificateError{
			Operation: "key block generation",
			Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
		}
	}
	if err := pem.Encode(keyOut, keyBlock); err != nil {
		return "", "", &CertificateError{
			Operation: "encoding",
			Path:      keyFile,
			Err:       fmt.Errorf("%w: %w", ErrCertificateGeneration, err),
		}
	}

	return certFile, keyFile, nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) (*pem.Block, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal ECDSA private key: %w", err)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, nil
	}
}

func (o *OpcUAClient) generateClientOpts(endpoints []*ua.EndpointDescription) ([]opcua.Option, error) {
	appuri := "urn:telegraf:gopcua:client"
	appname := "Telegraf"

	// ApplicationURI is automatically read from the cert so is not required if a cert if provided
	opts := []opcua.Option{
		opcua.ApplicationURI(appuri),
		opcua.ApplicationName(appname),
		opcua.RequestTimeout(time.Duration(o.Config.RequestTimeout)),
	}

	if o.Config.SessionTimeout != 0 {
		opts = append(opts, opcua.SessionTimeout(time.Duration(o.Config.SessionTimeout)))
	}

	certFile := o.Config.Certificate
	keyFile := o.Config.PrivateKey
	policy := o.Config.SecurityPolicy
	mode := o.Config.SecurityMode
	var err error
	if certFile == "" && keyFile == "" {
		if policy != "None" || mode != "None" {
			certFile, keyFile, err = generateCert(appuri, 2048, certFile, keyFile, 365*24*time.Hour)
			if err != nil {
				return nil, err
			}
		}
	}

	var cert []byte
	var pk *rsa.PrivateKey
	if certFile != "" && keyFile != "" {
		debug.Printf("Loading cert/key from %s/%s", certFile, keyFile)
		c, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			o.Log.Warnf("Failed to load certificate: %s", err)
		} else {
			pkTemp, ok := c.PrivateKey.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("invalid private key")
			}
			pk = pkTemp
			cert = c.Certificate[0]
			opts = append(opts, opcua.PrivateKey(pk), opcua.Certificate(cert))
		}
	}

	var secPolicy string
	switch {
	case policy == "auto":
		// set it later
	case strings.HasPrefix(policy, ua.SecurityPolicyURIPrefix):
		secPolicy = policy
		policy = ""
	case policy == "None" || policy == "Basic128Rsa15" || policy == "Basic256" || policy == "Basic256Sha256" ||
		policy == "Aes128_Sha256_RsaOaep" || policy == "Aes256_Sha256_RsaPss":
		secPolicy = ua.SecurityPolicyURIPrefix + policy
		policy = ""
	default:
		return nil, &SecurityError{
			Policy: policy,
			Err:    fmt.Errorf("%w: %s", ErrInvalidSecurityPolicy, policy),
		}
	}

	o.Log.Debugf("security policy from configuration %s", secPolicy)

	// Select the most appropriate authentication mode from server capabilities and user input
	authMode, authOptions, err := o.generateAuth(o.Config.AuthMethod, cert, pk, o.Config.Username, o.Config.Password)
	if err != nil {
		return nil, err
	}

	opts = append(opts, authOptions...)

	var secMode ua.MessageSecurityMode
	switch strings.ToLower(mode) {
	case "auto":
	case "none":
		secMode = ua.MessageSecurityModeNone
		mode = ""
	case "sign":
		secMode = ua.MessageSecurityModeSign
		mode = ""
	case "signandencrypt":
		secMode = ua.MessageSecurityModeSignAndEncrypt
		mode = ""
	default:
		return nil, &SecurityError{
			Mode: mode,
			Err:  fmt.Errorf("%w: %s", ErrInvalidSecurityMode, mode),
		}
	}

	// Allow input of only one of sec-mode,sec-policy when choosing 'None'
	if secMode == ua.MessageSecurityModeNone || secPolicy == ua.SecurityPolicyURINone {
		secMode = ua.MessageSecurityModeNone
		secPolicy = ua.SecurityPolicyURINone
	}

	// Find the best endpoint based on our input and server recommendation (highest SecurityMode+SecurityLevel)
	var serverEndpoint *ua.EndpointDescription
	switch {
	case mode == "auto" && policy == "auto": // No user selection, choose best
		for _, e := range endpoints {
			if serverEndpoint == nil || (e.SecurityMode >= serverEndpoint.SecurityMode && e.SecurityLevel >= serverEndpoint.SecurityLevel) {
				serverEndpoint = e
			}
		}

	case mode != "auto" && policy == "auto": // User only cares about mode, select highest securitylevel with that mode
		for _, e := range endpoints {
			if e.SecurityMode == secMode && (serverEndpoint == nil || e.SecurityLevel >= serverEndpoint.SecurityLevel) {
				serverEndpoint = e
			}
		}

	case mode == "auto" && policy != "auto": // User only cares about policy, select highest securitylevel with that policy
		for _, e := range endpoints {
			if e.SecurityPolicyURI == secPolicy && (serverEndpoint == nil || e.SecurityLevel >= serverEndpoint.SecurityLevel) {
				serverEndpoint = e
			}
		}

	default: // User cares about both
		o.Log.Debugf("User cares about both the policy (%s) and security mode (%s)", secPolicy, secMode)
		o.Log.Debugf("Server has %d endpoints", len(endpoints))
		for _, e := range endpoints {
			o.Log.Debugf("Evaluating endpoint %s, policy %s, mode %s, level %d", e.EndpointURL, e.SecurityPolicyURI, e.SecurityMode, e.SecurityLevel)
			if e.SecurityPolicyURI == secPolicy && e.SecurityMode == secMode && (serverEndpoint == nil || e.SecurityLevel >= serverEndpoint.SecurityLevel) {
				serverEndpoint = e
				o.Log.Debugf(
					"Security policy and mode found. Using server endpoint %s for security. Policy %s",
					serverEndpoint.EndpointURL,
					serverEndpoint.SecurityPolicyURI,
				)
			}
		}
	}

	if serverEndpoint == nil { // Didn't find an endpoint with matching policy and mode.
		return nil, &SecurityError{
			Policy: secPolicy,
			Mode:   mode,
			Err:    fmt.Errorf("%w: no suitable server endpoint found", ErrEndpointNotFound),
		}
	}

	secPolicy = serverEndpoint.SecurityPolicyURI
	secMode = serverEndpoint.SecurityMode

	// Check that the selected endpoint is a valid combo
	err = validateEndpointConfig(endpoints, secPolicy, secMode, authMode)
	if err != nil {
		return nil, fmt.Errorf("endpoint validation failed: %w", err)
	}

	opts = append(opts, opcua.SecurityFromEndpoint(serverEndpoint, authMode))

	// If a remote certificate is explicitly configured, use it to override
	// the certificate from the endpoint. This allows trusting self-signed certificates.
	if o.Config.RemoteCertificate != "" {
		o.Log.Debugf("Using explicitly configured remote certificate from %s", o.Config.RemoteCertificate)
		opts = append(opts, opcua.RemoteCertificateFile(o.Config.RemoteCertificate))
	}

	return opts, nil
}

func (o *OpcUAClient) generateAuth(a string, cert []byte, pk *rsa.PrivateKey, user, passwd config.Secret) (ua.UserTokenType, []opcua.Option, error) {
	var authMode ua.UserTokenType
	var authOptions []opcua.Option
	switch strings.ToLower(a) {
	case "anonymous":
		authMode = ua.UserTokenTypeAnonymous
		authOptions = []opcua.Option{opcua.AuthAnonymous()}
	case "username":
		authMode = ua.UserTokenTypeUserName

		var username, password []byte
		if !user.Empty() {
			usecret, err := user.Get()
			if err != nil {
				return 0, nil, &AuthenticationError{
					Method: a,
					Err:    fmt.Errorf("error reading username: %w", err),
				}
			}
			defer usecret.Destroy()
			username = usecret.Bytes()
		}

		if !passwd.Empty() {
			psecret, err := passwd.Get()
			if err != nil {
				return 0, nil, &AuthenticationError{
					Method: a,
					Err:    fmt.Errorf("error reading password: %w", err),
				}
			}
			defer psecret.Destroy()
			password = psecret.Bytes()
		}
		authOptions = []opcua.Option{opcua.AuthUsername(string(username), string(password))}
	case "certificate":
		authMode = ua.UserTokenTypeCertificate
		authOptions = []opcua.Option{opcua.AuthCertificate(cert)}
		if pk != nil {
			o.Log.Debug("Setting private key for certificate-based user authentication")
			authOptions = append(authOptions, opcua.AuthPrivateKey(pk))
		}
	case "issuedtoken":
		// TODO: this is unsupported, should we fail here or let the opcua package handle it?
		authMode = ua.UserTokenTypeIssuedToken
		authOptions = []opcua.Option{opcua.AuthIssuedToken([]byte(nil))}
	case "":
		// Default to anonymous when auth method is not specified
		authMode = ua.UserTokenTypeAnonymous
		authOptions = []opcua.Option{opcua.AuthAnonymous()}
	default:
		o.Log.Warnf("unknown auth method %q, defaulting to Anonymous", a)
		authMode = ua.UserTokenTypeAnonymous
		authOptions = []opcua.Option{opcua.AuthAnonymous()}
	}

	return authMode, authOptions, nil
}

func validateEndpointConfig(endpoints []*ua.EndpointDescription, secPolicy string, secMode ua.MessageSecurityMode, authMode ua.UserTokenType) error {
	for _, e := range endpoints {
		if e.SecurityMode == secMode && e.SecurityPolicyURI == secPolicy {
			for _, t := range e.UserIdentityTokens {
				if t.TokenType == authMode {
					return nil
				}
			}
		}
	}

	return &SecurityError{
		Policy: secPolicy,
		Mode:   secMode.String(),
		Err:    fmt.Errorf("%w: server does not support the specified security configuration", ErrEndpointNotFound),
	}
}
