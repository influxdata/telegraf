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
	host      string
	port      int
	transport string
	secure    bool
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
	Log          telegraf.Logger `toml:"-"`
	commontls.ClientConfig

	ua         *sipgo.UserAgent
	client     *sipgo.Client
	tlsConfig  *tls.Config
	serverInfo *serverInfo

	// Cached request components
	requestURI sip.Uri
	headers    []sip.Header
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

	// Validate server URL scheme and transport combination
	// Note: "tls" transport is deprecated per RFC 3261. Use sips:// scheme instead.
	u, err := net_url.Parse(s.Server)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	switch u.Scheme {
	case "sip":
		// sip:// requires non-secure transport (udp, tcp, ws)
		switch s.Transport {
		case "":
			s.Transport = "udp"
		case "udp", "tcp", "ws":
			// valid transports
		default:
			return fmt.Errorf("invalid transport %q; must be one of udp, tcp, ws", s.Transport)
		}
	case "sips":
		// sips:// requires secure transport (tcp or wss for TLS)
		switch s.Transport {
		case "":
			s.Transport = "tcp"
		case "tcp", "wss":
			// valid transports
		default:
			return fmt.Errorf("invalid transport %q for sips:// scheme; must be tcp or wss", s.Transport)
		}
	default:
		return fmt.Errorf("server URL must use sip:// or sips:// scheme, got %q", u.Scheme)
	}

	// Parse server info
	s.serverInfo = &serverInfo{
		secure:    u.Scheme == "sips",
		transport: s.Transport,
		host:      u.Hostname(),
	}

	if s.serverInfo.host == "" {
		return errors.New("server URL must specify a host")
	}

	portStr := u.Port()
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid port %q: %w", portStr, err)
		}
		s.serverInfo.port = port
	} else if s.serverInfo.secure {
		s.serverInfo.port = 5061
	} else {
		s.serverInfo.port = 5060
	}

	// Set FromDomain default after serverInfo is available
	if s.FromDomain == "" {
		s.FromDomain = s.serverInfo.host
	}

	// Setup TLS configuration if secure (sips:// scheme)
	if s.serverInfo.secure {
		// Force TLS connection even though no TLS properties are given. This will
		// use the system's TLS configuration (CA etc) if properties are empty.
		s.ClientConfig = &s.serverInfo.secure
		tlsConfig, err := s.ClientConfig.TLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		s.tlsConfig = tlsConfig
	}

	// Build cached request components
	s.requestURI = sip.Uri{
		Scheme: u.Scheme,
		User:   s.ToUser,
		Host:   s.serverInfo.host,
		Port:   s.serverInfo.port,
	}

	s.requestURI.UriParams = sip.NewParams()
	s.requestURI.UriParams.Add("transport", s.serverInfo.transport)

	// Build cached headers (To, User-Agent, Contact)
	// Note: From header has a dynamic tag per request, so it cannot be cached
	s.headers = []sip.Header{
		&sip.ToHeader{Address: s.requestURI},
		sip.NewHeader("User-Agent", internal.ProductToken()),
	}

	if s.LocalAddress != "" {
		s.headers = append(s.headers, &sip.ContactHeader{
			Address: sip.Uri{
				Scheme: u.Scheme,
				User:   s.FromUser,
				Host:   s.LocalAddress,
			},
		})
	}

	return nil
}

func (s *SIP) Start(telegraf.Accumulator) error {
	// Create SIP user agent with optional TLS config
	var opts []sipgo.UserAgentOption

	// Add TLS config if transport requires it
	if s.serverInfo.secure {
		opts = append(opts, sipgo.WithUserAgenTLSConfig(s.tlsConfig))
	}

	ua, err := sipgo.NewUA(opts...)
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
	}
	s.ua = nil
}

func (s *SIP) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]any)
	tags := map[string]string{
		"server":    s.Server,
		"method":    s.Method,
		"transport": s.serverInfo.transport,
	}

	// Create SIP request using cached requestURI
	req := sip.NewRequest(sip.RequestMethod(s.Method), s.requestURI)

	// Add From header (has dynamic tag per request, so cannot be cached)
	from := &sip.FromHeader{
		Address: sip.Uri{
			Scheme: s.requestURI.Scheme,
			User:   s.FromUser,
			Host:   s.FromDomain,
		},
		Params: sip.NewParams(),
	}
	from.Params.Add("tag", sip.GenerateTagN(16))
	req.AppendHeader(from)

	// Add cached headers (To, User-Agent, Contact)
	for _, header := range s.headers {
		req.AppendHeader(header)
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
			fields["up"] = 0
			acc.AddFields("sip", fields, tags)
			return nil
		}
		s.handleGatherError(err, fields, tags, acc)
		return nil
	}

	// Handle digest authentication challenge (RFC 8760)
	// SIP digest auth requires the server's challenge first to obtain the nonce,
	// so we cannot pre-authenticate on the first request.
	if res.StatusCode == 401 || res.StatusCode == 407 {
		// Get credentials
		usernameRaw, err := s.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		username := usernameRaw.String()
		usernameRaw.Destroy()

		passwordRaw, err := s.Password.Get()
		if err != nil {
			return fmt.Errorf("getting password failed: %w", err)
		}
		password := passwordRaw.String()
		passwordRaw.Destroy()

		res, err = s.client.DoDigestAuth(ctx, req, res, sipgo.DigestAuth{
			Username: username,
			Password: password,
		})
		if err != nil {
			s.handleGatherError(err, fields, tags, acc)
			return nil
		}
	}

	// Record response time
	fields["response_time"] = time.Since(start).Seconds()

	// Process response
	if res != nil {
		tags["status_code"] = strconv.Itoa(res.StatusCode)
		if serverAgent := res.GetHeader("Server"); serverAgent != nil {
			tags["server_agent"] = serverAgent.Value()
		}

		// Determine up based on status code (2xx = success)
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			fields["up"] = 1
		} else {
			fields["up"] = 0
		}
	} else {
		fields["up"] = 0
	}

	acc.AddFields("sip", fields, tags)
	return nil
}

func (s *SIP) handleGatherError(err error, fields map[string]any, tags map[string]string, acc telegraf.Accumulator) {
	s.Log.Debugf("SIP gather error: %v", err)
	// Mark as down for all connection failures
	fields["up"] = 0
	acc.AddFields("sip", fields, tags)
}

func init() {
	inputs.Add("sip", func() telegraf.Input {
		return &SIP{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
