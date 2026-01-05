//go:generate ../../../tools/readme_config_includer/generator
package sip

import (
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	commontls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultTimeout    = 5 * time.Second
	defaultMethod     = "OPTIONS"
	defaultTransport  = "udp"
	defaultPort       = 5060
	defaultTLSPort    = 5061
	defaultFromUser   = "telegraf"
	defaultUserAgent  = "Telegraf SIP Monitor"
	defaultExpectCode = 200
)

var resultCodes = map[string]int{
	"success":                0,
	"response_code_mismatch": 1,
	"timeout":                2,
	"connection_refused":     3,
	"connection_failed":      4,
	"no_response":            5,
	"parse_error":            6,
	"request_error":          7,
	"transaction_error":      8,
	"no_route":               9,
	"network_unreachable":    10,
	"error_response":         11,
	"auth_required":          12,
	"auth_failed":            13,
	"auth_error":             14,
}

type serverInfo struct {
	Host      string
	Port      int
	Transport string
	Secure    bool
}

type SIP struct {
	Servers      []string        `toml:"servers"`
	Method       string          `toml:"method"`
	Timeout      config.Duration `toml:"timeout"`
	FromUser     string          `toml:"from_user"`
	FromDomain   string          `toml:"from_domain"`
	ToUser       string          `toml:"to_user"`
	UserAgent    string          `toml:"user_agent"`
	ExpectCode   int             `toml:"expect_code"`
	LocalAddress string          `toml:"local_address"`
	Username     config.Secret   `toml:"username"`
	Password     config.Secret   `toml:"password"`
	commontls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	ua        *sipgo.UserAgent
	client    *sipgo.Client
	tlsConfig *tls.Config
	mu        sync.Mutex
}

func (*SIP) SampleConfig() string {
	return sampleConfig
}

func (s *SIP) Init() error {
	// Set defaults
	if s.Timeout == 0 {
		s.Timeout = config.Duration(defaultTimeout)
	}
	if s.Method == "" {
		s.Method = defaultMethod
	}
	if s.FromUser == "" {
		s.FromUser = defaultFromUser
	}
	if s.UserAgent == "" {
		s.UserAgent = defaultUserAgent
	}
	if s.ExpectCode == 0 {
		s.ExpectCode = defaultExpectCode
	}

	// Validate method
	validMethods := map[string]bool{
		"OPTIONS": true,
		"INVITE":  true,
		"MESSAGE": true,
	}
	if !validMethods[strings.ToUpper(s.Method)] {
		return fmt.Errorf("invalid SIP method %q: must be OPTIONS, INVITE, or MESSAGE", s.Method)
	}
	s.Method = strings.ToUpper(s.Method)

	// Validate servers
	if len(s.Servers) == 0 {
		return errors.New("at least one server must be specified")
	}

	// Parse all servers to check if any require TLS
	needsTLS := false
	for _, server := range s.Servers {
		info, err := parseServer(server)
		if err != nil {
			return fmt.Errorf("failed to parse server %q: %w", server, err)
		}
		if info.Secure || info.Transport == "tls" || info.Transport == "wss" {
			needsTLS = true
			break
		}
	}

	// Setup TLS configuration if any server requires it
	if needsTLS {
		tlsConfig, err := s.ClientConfig.TLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		s.tlsConfig = tlsConfig
	}

	return nil
}

func (s *SIP) Start(_ telegraf.Accumulator) error {
	// Create SIP user agent with optional TLS config
	uaOpts := []sipgo.UserAgentOption{
		sipgo.WithUserAgent(s.UserAgent),
	}

	// Add TLS config if transport requires it
	if s.tlsConfig != nil {
		uaOpts = append(uaOpts, sipgo.WithUserAgenTLSConfig(s.tlsConfig))
	}

	ua, err := sipgo.NewUA(uaOpts...)
	if err != nil {
		return fmt.Errorf("failed to create SIP user agent: %w", err)
	}

	// Create SIP client
	client, err := sipgo.NewClient(ua)
	if err != nil {
		ua.Close()
		return fmt.Errorf("failed to create SIP client: %w", err)
	}

	s.mu.Lock()
	s.ua = ua
	s.client = client
	s.mu.Unlock()

	return nil
}

func (s *SIP) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		s.client = nil
	}
	if s.ua != nil {
		s.ua.Close()
		s.ua = nil
	}
}

// Helper functions

func getScheme(secure bool) string {
	if secure {
		return "sips"
	}
	return "sip"
}

func (*SIP) addResponseMetadata(res *sip.Response, tags map[string]string) {
	if res == nil {
		return
	}
	tags["status_code"] = strconv.Itoa(res.StatusCode)
	if reason := res.Reason; reason != "" {
		tags["reason"] = reason
	}
	if serverAgent := res.GetHeader("Server"); serverAgent != nil {
		tags["server_agent"] = serverAgent.Value()
	}
}

func (s *SIP) getCredentials() (username, password string, err error) {
	u, err := s.Username.Get()
	if err != nil {
		return "", "", fmt.Errorf("failed to get username: %w", err)
	}
	defer u.Destroy()

	p, err := s.Password.Get()
	if err != nil {
		return "", "", fmt.Errorf("failed to get password: %w", err)
	}
	defer p.Destroy()

	return u.String(), p.String(), nil
}

func setResult(resultString string, fields map[string]any, tags map[string]string) {
	tags["result"] = resultString
	fields["result_type"] = resultString
	if code, ok := resultCodes[resultString]; ok {
		fields["result_code"] = code
	} else {
		fields["result_code"] = 99
	}
}

func (*SIP) recordMetric(result string, fields map[string]any, tags map[string]string, acc telegraf.Accumulator) {
	setResult(result, fields, tags)
	acc.AddFields("sip", fields, tags)
}

func (s *SIP) Gather(acc telegraf.Accumulator) error {
	s.mu.Lock()
	client := s.client
	s.mu.Unlock()

	if client == nil {
		return errors.New("sip client not initialized")
	}

	var wg sync.WaitGroup
	for _, server := range s.Servers {
		wg.Add(1)
		go func(srv string) {
			defer wg.Done()
			s.gatherServer(srv, client, acc)
		}(server)
	}
	wg.Wait()

	return nil
}

func (s *SIP) gatherServer(server string, client *sipgo.Client, acc telegraf.Accumulator) {
	fields := make(map[string]any)

	// Parse server address and extract transport
	info, err := parseServer(server)
	if err != nil {
		s.Log.Errorf("Failed to parse server %q: %s", server, err)
		tags := map[string]string{
			"server": server,
			"method": s.Method,
		}
		setResult("parse_error", fields, tags)
		acc.AddFields("sip", fields, tags)
		return
	}

	tags := map[string]string{
		"server":    server,
		"method":    s.Method,
		"transport": info.Transport,
	}

	// Build SIP URI
	sipURI := s.buildSIPURI(info.Host, info.Port, info.Transport, info.Secure)
	tags["sip_uri"] = sipURI.String()

	// Create SIP request
	req := s.createRequest(sipURI, info.Host, info.Secure)

	// Send request and measure response time
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout))
	defer cancel()

	tx, err := client.TransactionRequest(ctx, req)
	if err != nil {
		s.Log.Debugf("Failed to send SIP request to %q: %s", server, err)
		handleError(err, fields, tags)
		acc.AddFields("sip", fields, tags)
		return
	}
	defer tx.Terminate()

	// Wait for response
	var res *sip.Response
	select {
	case res = <-tx.Responses():
		// Response received, will be processed below

	case <-ctx.Done():
		s.Log.Debugf("SIP request to %q timed out", server)
		setResult("timeout", fields, tags)
		acc.AddFields("sip", fields, tags)
		return

	case <-tx.Done():
		if tx.Err() != nil {
			s.Log.Debugf("SIP transaction error for %q: %s", server, tx.Err())
			handleError(tx.Err(), fields, tags)
		} else {
			setResult("transaction_error", fields, tags)
		}
		acc.AddFields("sip", fields, tags)
		return
	}

	// Handle authentication challenge (401 Unauthorized or 407 Proxy Authentication Required)
	if res != nil && (res.StatusCode == 401 || res.StatusCode == 407) {
		username, password, err := s.getCredentials()
		if err != nil {
			s.Log.Errorf("Failed to get credentials for %q: %s", server, err)
			fields["response_time"] = time.Since(start).Seconds()
			s.recordMetric("auth_error", fields, tags, acc)
			return
		}

		if username == "" || password == "" {
			// No credentials configured but authentication is required
			s.Log.Debugf("SIP server %q requires authentication but no credentials provided", server)
			fields["response_time"] = time.Since(start).Seconds()
			s.addResponseMetadata(res, tags)
			s.recordMetric("auth_required", fields, tags, acc)
			return
		}

		// Credentials provided, attempt digest authentication
		s.Log.Debugf("SIP server %q requires authentication, attempting digest auth", server)

		auth := sipgo.DigestAuth{
			Username: username,
			Password: password,
		}

		// Perform digest authentication (this handles the retry internally)
		authRes, err := client.DoDigestAuth(ctx, req, res, auth)
		if err != nil {
			s.Log.Debugf("Digest authentication failed for %q: %s", server, err)
			fields["response_time"] = time.Since(start).Seconds()
			s.recordMetric("auth_failed", fields, tags, acc)
			return
		}
		// Use the authenticated response
		res = authRes
	}

	// Process the final response
	fields["response_time"] = time.Since(start).Seconds()

	if res != nil {
		statusCode := res.StatusCode
		s.addResponseMetadata(res, tags)

		// Check if response matches expected code
		if s.ExpectCode > 0 {
			if statusCode == s.ExpectCode {
				fields["response_code_match"] = 1
				s.recordMetric("success", fields, tags, acc)
			} else {
				fields["response_code_match"] = 0
				s.recordMetric("response_code_mismatch", fields, tags, acc)
			}
		} else {
			// If no expected code, any 2xx response is success
			if statusCode >= 200 && statusCode < 300 {
				s.recordMetric("success", fields, tags, acc)
			} else {
				s.recordMetric("error_response", fields, tags, acc)
			}
		}
	} else {
		s.recordMetric("no_response", fields, tags, acc)
	}
}

// parseServer parses a server URL and returns parsed server information
// Supports RFC 3261 compliant SIP URI formats:
//   - sip://host:port (defaults to UDP transport)
//   - sips://host:port (defaults to TLS transport, uses secure scheme)
//   - sip://host:port;transport=tcp
//   - sip://host:port;transport=udp
//   - sip://host:port;transport=ws
//   - sips://host:port;transport=wss
func parseServer(server string) (*serverInfo, error) {
	// Server must have sip:// or sips:// scheme
	if !strings.HasPrefix(server, "sip://") && !strings.HasPrefix(server, "sips://") {
		return nil, fmt.Errorf("server URL must start with sip:// or sips:// scheme, got: %q", server)
	}

	info := &serverInfo{}

	// Parse as SIP URI
	info.Secure = strings.HasPrefix(server, "sips://")

	// Remove scheme
	uriStr := strings.TrimPrefix(server, "sip://")
	uriStr = strings.TrimPrefix(uriStr, "sips://")

	// Check for transport parameter (e.g., ;transport=tcp)
	if idx := strings.Index(uriStr, ";transport="); idx != -1 {
		info.Transport = strings.ToLower(uriStr[idx+11:])
		if endIdx := strings.Index(info.Transport, ";"); endIdx != -1 {
			info.Transport = info.Transport[:endIdx]
		}
		uriStr = uriStr[:idx]
	} else {
		// No explicit transport, infer from scheme
		if info.Secure {
			info.Transport = "tls"
		} else {
			info.Transport = "udp"
		}
	}

	// Validate transport
	validTransports := map[string]bool{
		"udp": true,
		"tcp": true,
		"tls": true,
		"ws":  true,
		"wss": true,
	}
	if !validTransports[info.Transport] {
		return nil, fmt.Errorf("invalid transport %q in URI: must be udp, tcp, tls, ws, or wss", info.Transport)
	}

	// Parse host:port
	if strings.Contains(uriStr, ":") {
		h, p, err := net.SplitHostPort(uriStr)
		if err != nil {
			return nil, fmt.Errorf("invalid server address: %w", err)
		}
		info.Host = h
		portNum, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q: %w", p, err)
		}
		info.Port = portNum
	} else {
		// No port specified, use default based on secure flag
		info.Host = uriStr
		if info.Secure {
			info.Port = defaultTLSPort
		} else {
			info.Port = defaultPort
		}
	}

	return info, nil
}

func (s *SIP) buildSIPURI(host string, port int, transport string, secure bool) *sip.Uri {
	scheme := getScheme(secure)

	// Use ToUser if specified, otherwise use the same as FromUser
	toUser := s.ToUser
	if toUser == "" {
		toUser = s.FromUser
	}

	uri := &sip.Uri{
		Scheme: scheme,
		User:   toUser,
		Host:   host,
		Port:   port,
	}

	// Add transport parameter for non-default transports
	// UDP is default for sip:, TLS is default for sips:, so only add explicit parameter for others
	if (scheme == "sip" && transport != "udp") || (scheme == "sips" && transport != "tls") {
		uri.UriParams = sip.NewParams()
		uri.UriParams.Add("transport", transport)
	}

	return uri
}

func (s *SIP) createRequest(requestURI *sip.Uri, host string, secure bool) *sip.Request {
	// Determine From domain
	fromDomain := s.FromDomain
	if fromDomain == "" {
		fromDomain = host
	}

	// Build From URI with correct scheme
	fromScheme := getScheme(secure)
	fromURI := sip.Uri{
		Scheme: fromScheme,
		User:   s.FromUser,
		Host:   fromDomain,
	}

	// Create request (method is already validated and uppercased in Init)
	req := sip.NewRequest(sip.RequestMethod(s.Method), *requestURI)

	// Add required headers
	from := &sip.FromHeader{
		Address: fromURI,
		Params:  sip.NewParams(),
	}
	from.Params.Add("tag", sip.GenerateTagN(16))

	to := &sip.ToHeader{
		Address: *requestURI,
	}

	req.AppendHeader(from)
	req.AppendHeader(to)

	// Add User-Agent header
	req.AppendHeader(sip.NewHeader("User-Agent", s.UserAgent))

	// Add Contact header with local address if specified
	if s.LocalAddress != "" {
		contactURI := sip.Uri{
			Scheme: fromScheme,
			User:   s.FromUser,
			Host:   s.LocalAddress,
		}
		contact := &sip.ContactHeader{
			Address: contactURI,
		}
		req.AppendHeader(contact)
	}

	return req
}

func handleError(err error, fields map[string]any, tags map[string]string) {
	errStr := err.Error()

	// Categorize errors
	if errors.Is(err, context.DeadlineExceeded) {
		setResult("timeout", fields, tags)
	} else if strings.Contains(errStr, "connection refused") {
		setResult("connection_refused", fields, tags)
	} else if strings.Contains(errStr, "no route to host") {
		setResult("no_route", fields, tags)
	} else if strings.Contains(errStr, "network is unreachable") {
		setResult("network_unreachable", fields, tags)
	} else {
		setResult("connection_failed", fields, tags)
	}
}

func init() {
	inputs.Add("sip", func() telegraf.Input {
		return &SIP{}
	})
}
