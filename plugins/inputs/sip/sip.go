//go:generate ../../../tools/readme_config_includer/generator
package sip

import (
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	net_url "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	commontls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type serverInfo struct {
	Host      string
	Port      int
	Transport string
	Secure    bool
}

type SIP struct {
	Server       string          `toml:"server"`
	Transport    string          `toml:"transport"`
	Method       string          `toml:"method"`
	Timeout      config.Duration `toml:"timeout"`
	FromUser     string          `toml:"from_user"`
	FromDomain   string          `toml:"from_domain"`
	ToUser       string          `toml:"to_user"`
	LocalAddress string          `toml:"local_address"`
	Username     config.Secret   `toml:"username"`
	Password     config.Secret   `toml:"password"`
	commontls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	ua         *sipgo.UserAgent
	client     *sipgo.Client
	tlsConfig  *tls.Config
	serverInfo *serverInfo
}

func (*SIP) SampleConfig() string {
	return sampleConfig
}

func (s *SIP) Init() error {
	// Set defaults
	if s.FromUser == "" {
		s.FromUser = "telegraf"
	}
	if s.ToUser == "" {
		s.ToUser = s.FromUser
	}
	if s.Transport == "" {
		s.Transport = "udp"
	}

	// Validate server
	if s.Server == "" {
		return errors.New("server must be specified")
	}

	// Validate method
	switch strings.ToUpper(s.Method) {
	case "":
		s.Method = "OPTIONS"
	case "OPTIONS", "INVITE", "MESSAGE":
		s.Method = strings.ToUpper(s.Method)
	default:
		return fmt.Errorf("invalid SIP method %q", s.Method)
	}

	if s.Timeout <= 0 {
		return errors.New("timeout has to be greater than zero")
	}

	// Validate server URL scheme
	u, err := net_url.Parse(s.Server)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	if u.Scheme != "sip" && u.Scheme != "sips" {
		return fmt.Errorf("server URL must use sip:// or sips:// scheme, got %q", u.Scheme)
	}

	// Validate transport
	// Note: "tls" transport is deprecated per RFC 3261. Use sips:// scheme instead.
	switch strings.ToLower(s.Transport) {
	case "udp", "tcp", "ws", "wss":
		s.Transport = strings.ToLower(s.Transport)
	default:
		return fmt.Errorf("invalid transport %q", s.Transport)
	}

	// Parse server info
	info, err := parseServer(u, s.Transport)
	if err != nil {
		return fmt.Errorf("failed to parse server: %w", err)
	}
	s.serverInfo = info

	// Set FromDomain default after serverInfo is available
	if s.FromDomain == "" {
		s.FromDomain = s.serverInfo.Host
	}

	// Setup TLS configuration if transport requires it
	if info.Secure || info.Transport == "wss" {
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
	var uaOpts []sipgo.UserAgentOption

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

	s.ua = ua
	s.client = client

	return nil
}

func (s *SIP) Stop() {
	if s.ua != nil {
		s.ua.Close()
		s.ua = nil
	}
	s.client = nil
}

func (s *SIP) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]any)
	tags := map[string]string{
		"server":    s.Server,
		"method":    s.Method,
		"transport": s.serverInfo.Transport,
	}

	// Build SIP URI for the request
	scheme := "sip"
	if s.serverInfo.Secure {
		scheme = "sips"
	}

	requestURI := &sip.Uri{
		Scheme: scheme,
		User:   s.ToUser,
		Host:   s.serverInfo.Host,
		Port:   s.serverInfo.Port,
	}

	// Add transport parameter for non-default transports
	// Note: sips:// always uses TLS via the scheme, not the transport parameter
	if scheme == "sip" && s.serverInfo.Transport != "udp" {
		requestURI.UriParams = sip.NewParams()
		requestURI.UriParams.Add("transport", s.serverInfo.Transport)
	}

	// Create SIP request
	req := sip.NewRequest(sip.RequestMethod(s.Method), *requestURI)

	// Add From header
	fromURI := sip.Uri{
		Scheme: scheme,
		User:   s.FromUser,
		Host:   s.FromDomain,
	}
	from := &sip.FromHeader{
		Address: fromURI,
		Params:  sip.NewParams(),
	}
	from.Params.Add("tag", sip.GenerateTagN(16))
	req.AppendHeader(from)

	// Add To header
	to := &sip.ToHeader{
		Address: *requestURI,
	}
	req.AppendHeader(to)

	// Add User-Agent header
	req.AppendHeader(sip.NewHeader("User-Agent", internal.ProductToken()))

	// Add Contact header if local address specified
	if s.LocalAddress != "" {
		contactURI := sip.Uri{
			Scheme: scheme,
			User:   s.FromUser,
			Host:   s.LocalAddress,
		}
		contact := &sip.ContactHeader{
			Address: contactURI,
		}
		req.AppendHeader(contact)
	}

	// Send request and measure response time
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout))
	defer cancel()

	// Send request
	res, err := s.client.Do(ctx, req)
	if err != nil {
		// Check if it's a timeout
		if errors.Is(err, context.DeadlineExceeded) {
			tags["result"] = "timeout"
			fields["reason"] = "Timeout"
			acc.AddFields("sip", fields, tags)
			return nil
		}
		s.handleGatherError(err, fields, tags, acc)
		return nil
	}

	// If we got auth challenge and have credentials, retry with authentication
	if res.StatusCode == 401 || res.StatusCode == 407 {
		// Get credentials
		username, err := s.Username.Get()
		if err == nil {
			defer username.Destroy()
			password, err := s.Password.Get()
			if err == nil {
				defer password.Destroy()

				auth := sipgo.DigestAuth{
					Username: username.String(),
					Password: password.String(),
				}
				res, err = s.client.DoDigestAuth(ctx, req, res, auth)
				if err != nil {
					s.handleGatherError(err, fields, tags, acc)
					return nil
				}
			}
		}
	}

	// Record response time
	fields["response_time"] = time.Since(start).Seconds()

	// Process response
	if res != nil {
		tags["status_code"] = strconv.Itoa(res.StatusCode)
		if res.Reason != "" {
			fields["reason"] = res.Reason
		}
		if serverAgent := res.GetHeader("Server"); serverAgent != nil {
			tags["server_agent"] = serverAgent.Value()
		}

		// Determine result based on status code
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			tags["result"] = "success"
		} else if res.StatusCode == 401 || res.StatusCode == 407 {
			// Check if we have credentials configured
			u, uErr := s.Username.Get()
			p, pErr := s.Password.Get()
			hasCredentials := uErr == nil && pErr == nil && u.String() != "" && p.String() != ""
			if uErr == nil {
				u.Destroy()
			}
			if pErr == nil {
				p.Destroy()
			}

			if hasCredentials {
				tags["result"] = "auth_failed"
			} else {
				tags["result"] = "auth_required"
			}
		} else {
			tags["result"] = "error_response"
		}
	} else {
		tags["result"] = "no_response"
	}

	acc.AddFields("sip", fields, tags)
	return nil
}

func (*SIP) handleGatherError(err error, fields map[string]any, tags map[string]string, acc telegraf.Accumulator) {
	errStr := err.Error()

	// Categorize errors
	if errors.Is(err, context.DeadlineExceeded) {
		tags["result"] = "timeout"
	} else if strings.Contains(errStr, "connection refused") {
		tags["result"] = "connection_refused"
	} else if strings.Contains(errStr, "no route to host") {
		tags["result"] = "no_route"
	} else if strings.Contains(errStr, "network is unreachable") {
		tags["result"] = "network_unreachable"
	} else {
		tags["result"] = "connection_failed"
	}

	acc.AddFields("sip", fields, tags)
}

// parseServer parses a server URL and returns parsed server information
// Supports RFC 3261 compliant SIP URI formats:
//   - sip://host:port (defaults to UDP transport)
//   - sips://host:port (defaults to TLS transport, uses secure scheme)
//   - sip://host:port;transport=tcp
//   - sip://host:port;transport=udp
//   - sip://host:port;transport=ws
//   - sips://host:port;transport=wss
func parseServer(u *net_url.URL, transport string) (*serverInfo, error) {
	info := &serverInfo{
		Secure:    u.Scheme == "sips",
		Transport: transport,
		Host:      u.Hostname(),
	}

	if info.Host == "" {
		return nil, errors.New("server URL must specify a host")
	}

	// Parse port
	portStr := u.Port()
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q: %w", portStr, err)
		}
		info.Port = port
	} else {
		// Use default port based on scheme
		if info.Secure {
			info.Port = 5061
		} else {
			info.Port = 5060
		}
	}

	return info, nil
}

func init() {
	inputs.Add("sip", func() telegraf.Input {
		return &SIP{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
