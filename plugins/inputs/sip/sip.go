//go:generate ../../../tools/readme_config_includer/generator
package sip

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
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
	tls.ClientConfig

	ua         *sipgo.UserAgent
	client     *sipgo.Client
	serverInfo *serverInfo
	uaOpts     []sipgo.UserAgentOption

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
	switch s.Method {
	case "":
		s.Method = "OPTIONS"
	default:
		if err := choice.Check(s.Method, []string{"OPTIONS", "INVITE", "MESSAGE"}); err != nil {
			return fmt.Errorf("invalid SIP method: %w", err)
		}
	}

	if s.Timeout < 0 {
		return errors.New("timeout has to be greater than or equal to zero")
	}

	// Validate server URL scheme and transport combination
	// Note: "tls" transport is deprecated per RFC 3261. Use sips:// scheme instead.
	u, err := url.Parse(s.Server)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	switch u.Scheme {
	case "sip":
		// sip:// requires non-secure transport (udp, tcp, ws)
		switch s.Transport {
		case "":
			s.Transport = "udp"
		default:
			if err := choice.Check(s.Transport, []string{"udp", "tcp", "ws"}); err != nil {
				return fmt.Errorf("invalid transport %q; must be one of udp, tcp, ws", s.Transport)
			}
		}
	case "sips":
		// sips:// requires secure transport (tcp or wss for TLS)
		switch s.Transport {
		case "":
			s.Transport = "tcp"
		default:
			if err := choice.Check(s.Transport, []string{"tcp", "wss"}); err != nil {
				return fmt.Errorf("invalid transport %q for sips:// scheme; must be tcp or wss", s.Transport)
			}
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
		if s.ClientConfig.Enable == nil {
			s.ClientConfig.Enable = &s.serverInfo.secure
		}
		tlsConfig, err := s.ClientConfig.TLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		s.uaOpts = append(s.uaOpts, sipgo.WithUserAgenTLSConfig(tlsConfig))
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
	ua, err := sipgo.NewUA(s.uaOpts...)
	if err != nil {
		return fmt.Errorf("failed to create SIP user agent: %w", err)
	}
	s.ua = ua

	// Create SIP client
	client, err := sipgo.NewClient(ua)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to create SIP client: %w", err)
	}
	s.client = client

	return nil
}

func (s *SIP) Stop() {
	if s.ua != nil {
		s.ua.Close()
	}
	s.ua = nil
	s.client = nil
}

func (s *SIP) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]any)
	tags := map[string]string{
		"source":    s.Server,
		"method":    strings.ToLower(s.Method),
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
			fields["result"] = "Timeout"
			fields["response_time_s"] = time.Duration(s.Timeout).Seconds()
			acc.AddFields("sip", fields, tags)
			return nil
		}
		// Handle other errors inline
		s.Log.Debugf("unauthenticated request to %q failed with: %v", s.Server, err)
		fields["result"] = "Error"
		fields["response_time_s"] = time.Since(start).Seconds()
		acc.AddFields("sip", fields, tags)
		return nil
	}

	// Handle digest authentication challenge (RFC 8760)
	// SIP digest auth requires the server's challenge first to obtain the nonce,
	// so we cannot pre-authenticate on the first request.
	if (res.StatusCode == 401 || res.StatusCode == 407) && !s.Username.Empty() {
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
			s.Log.Debugf("authenticated request to %q failed with: %v", s.Server, err)
			fields["result"] = "Error"
			fields["response_time_s"] = time.Since(start).Seconds()
			acc.AddFields("sip", fields, tags)
			return nil
		}
	}

	// Record response time
	fields["response_time_s"] = time.Since(start).Seconds()

	// Process response
	if res != nil {
		tags["status_code"] = strconv.Itoa(res.StatusCode)
		fields["result"] = res.Reason
		if serverAgent := res.GetHeader("Server"); serverAgent != nil {
			fields["server_agent"] = serverAgent.Value()
		}
	} else {
		fields["result"] = "No Response"
		s.Log.Debugf("no response from %q", s.Server)
	}

	acc.AddFields("sip", fields, tags)
	return nil
}

func init() {
	inputs.Add("sip", func() telegraf.Input {
		return &SIP{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
