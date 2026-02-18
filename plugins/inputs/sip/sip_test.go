package sip

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestSampleConfig(t *testing.T) {
	plugin := &SIP{}
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestInitDefaults(t *testing.T) {
	plugin := &SIP{
		Server:  "sip://sip.example.com:5060",
		Timeout: config.Duration(5 * time.Second),
	}

	require.NoError(t, plugin.Init())

	// Check defaults
	require.Equal(t, "OPTIONS", plugin.Method)
	require.Equal(t, "telegraf", plugin.FromUser)
	require.Equal(t, "telegraf", plugin.ToUser)
	require.Equal(t, "udp", plugin.Transport)
}

func TestInitErrors(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *SIP
		expected string
	}{
		{
			name:     "no_server",
			plugin:   &SIP{},
			expected: "server must be specified",
		},
		{
			name: "invalid_method",
			plugin: &SIP{
				Server:  "sip://sip.example.com:5060",
				Method:  "INVALID",
				Timeout: config.Duration(5 * time.Second),
			},
			expected: "invalid SIP method",
		},
		{
			name: "invalid_transport",
			plugin: &SIP{
				Server:    "sip://sip.example.com:5060",
				Transport: "invalid",
				Timeout:   config.Duration(5 * time.Second),
			},
			expected: "invalid transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestInitValidMethods(t *testing.T) {
	validMethods := []string{"OPTIONS", "INVITE", "MESSAGE"}

	for _, method := range validMethods {
		t.Run(method, func(t *testing.T) {
			plugin := &SIP{
				Server:  "sip://sip.example.com:5060",
				Method:  method,
				Timeout: config.Duration(5 * time.Second),
			}

			require.NoError(t, plugin.Init())
			require.Equal(t, method, plugin.Method)
		})
	}
}

func TestInitInvalidMethodCase(t *testing.T) {
	invalidCaseMethods := []string{"options", "invite", "message", "OpTiOnS"}

	for _, method := range invalidCaseMethods {
		t.Run(method, func(t *testing.T) {
			plugin := &SIP{
				Server:  "sip://sip.example.com:5060",
				Method:  method,
				Timeout: config.Duration(5 * time.Second),
			}

			require.ErrorContains(t, plugin.Init(), "invalid SIP method")
		})
	}
}

func TestInitValidTransports(t *testing.T) {
	validTransports := []struct {
		name      string
		server    string
		transport string
	}{
		{"udp", "sip://sip.example.com:5060", "udp"},
		{"tcp", "sip://sip.example.com:5060", "tcp"},
		{"ws", "sip://sip.example.com:5060", "ws"},
		{"wss", "sips://sip.example.com:5061", "wss"},
	}

	for _, tt := range validTransports {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				Server:    tt.server,
				Transport: tt.transport,
				Timeout:   config.Duration(5 * time.Second),
			}

			require.NoError(t, plugin.Init())
			require.Equal(t, tt.transport, plugin.Transport)
		})
	}
}

func TestParseServer(t *testing.T) {
	tests := []struct {
		name      string
		server    string
		transport string
		want      serverInfo
	}{
		{
			name:      "sip:// URL with UDP transport",
			server:    "sip://sip.example.com:5060",
			transport: "udp",
			want:      serverInfo{host: "sip.example.com", port: 5060, transport: "udp", secure: false},
		},
		{
			name:      "sips:// URL with TCP transport",
			server:    "sips://sip.example.com:5061",
			transport: "tcp",
			want:      serverInfo{host: "sip.example.com", port: 5061, transport: "tcp", secure: true},
		},
		{
			name:      "sip:// with TCP transport",
			server:    "sip://sip.example.com:5060",
			transport: "tcp",
			want:      serverInfo{host: "sip.example.com", port: 5060, transport: "tcp", secure: false},
		},
		{
			name:      "sip:// without port defaults to 5060",
			server:    "sip://sip.example.com",
			transport: "udp",
			want:      serverInfo{host: "sip.example.com", port: 5060, transport: "udp", secure: false},
		},
		{
			name:      "sips:// without port defaults to 5061",
			server:    "sips://secure.example.com",
			transport: "tcp",
			want:      serverInfo{host: "secure.example.com", port: 5061, transport: "tcp", secure: true},
		},
		{
			name:      "IP address with port",
			server:    "sip://192.168.1.100:5070",
			transport: "udp",
			want:      serverInfo{host: "192.168.1.100", port: 5070, transport: "udp", secure: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				Server:    tt.server,
				Transport: tt.transport,
				Timeout:   config.Duration(5 * time.Second),
			}

			require.NoError(t, plugin.Init())
			require.Equal(t, tt.want, *plugin.serverInfo)
		})
	}
}

func TestStartStop(t *testing.T) {
	plugin := &SIP{
		Server:  "sip://sip.example.com:5060",
		Timeout: config.Duration(5 * time.Second),
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator

	// Test Start
	require.NoError(t, plugin.Start(&acc))

	// Ensure user agent is created
	require.NotNil(t, plugin.ua)

	// Test Stop
	plugin.Stop()

	// Ensure user agent is cleaned up
	require.Nil(t, plugin.ua)
}

func TestTLSConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		server             string
		transport          string
		insecureSkipVerify bool
		shouldInitTLS      bool
	}{
		{
			name:          "TLS via sips:// scheme",
			server:        "sips://sip.example.com:5061",
			transport:     "tcp",
			shouldInitTLS: true,
		},
		{
			name:               "TLS with skip verify",
			server:             "sips://sip.example.com:5061",
			transport:          "tcp",
			insecureSkipVerify: true,
			shouldInitTLS:      true,
		},
		{
			name:          "WSS transport",
			server:        "sips://sip.example.com:5061",
			transport:     "wss",
			shouldInitTLS: true,
		},
		{
			name:          "UDP transport (no TLS)",
			server:        "sip://sip.example.com:5060",
			transport:     "udp",
			shouldInitTLS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				Server:    tt.server,
				Transport: tt.transport,
				Method:    "OPTIONS",
				Timeout:   config.Duration(5 * time.Second),
				FromUser:  "telegraf",
				Log:       testutil.Logger{},
			}

			plugin.ClientConfig.InsecureSkipVerify = tt.insecureSkipVerify

			require.NoError(t, plugin.Init())

			if tt.shouldInitTLS {
				require.NotEmpty(t, plugin.uaOpts, "uaOpts should contain TLS options")
			} else {
				require.Empty(t, plugin.uaOpts, "uaOpts should be empty for non-TLS")
			}
		})
	}
}

func TestTLSServerName(t *testing.T) {
	plugin := &SIP{
		Server:    "sips://192.168.1.100:5061",
		Transport: "tcp",
		Method:    "OPTIONS",
		Timeout:   config.Duration(5 * time.Second),
		FromUser:  "telegraf",
		Log:       testutil.Logger{},
	}

	plugin.ClientConfig.ServerName = "sip.example.com"

	require.NoError(t, plugin.Init())
	require.NotEmpty(t, plugin.uaOpts, "uaOpts should contain TLS options for sips://")
}

func TestInitRejectsDeprecatedTLSTransport(t *testing.T) {
	plugin := &SIP{
		Server:    "sip://sip.example.com:5060",
		Transport: "tls",
		Timeout:   config.Duration(5 * time.Second),
	}

	require.ErrorContains(t, plugin.Init(), "invalid transport")
}

func TestSecureProtocolWithoutTLSConfig(t *testing.T) {
	// Test that sips:// works with default TLS config
	plugin := &SIP{
		Server:   "sips://sip.example.com:5061",
		Method:   "OPTIONS",
		Timeout:  config.Duration(5 * time.Second),
		FromUser: "telegraf",
		Log:      testutil.Logger{},
	}

	require.NoError(t, plugin.Init())
	require.NotEmpty(t, plugin.uaOpts, "uaOpts should contain TLS options for sips://")
}

func TestSIPServerSuccess(t *testing.T) {
	server, err := newMockServer(sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	require.NoError(t, err)
	defer server.close()

	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "OPTIONS",
		Timeout:  config.Duration(2 * time.Second),
		FromUser: "telegraf",
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "200",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
				"result":          "OK",
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}

func TestSIPServerErrorResponse(t *testing.T) {
	server, err := newMockServer(sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 404, "Not Found", nil)
		require.NoError(t, tx.Respond(res))
	})
	require.NoError(t, err)
	defer server.close()

	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "OPTIONS",
		Timeout:  config.Duration(2 * time.Second),
		FromUser: "telegraf",
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "404",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
				"result":          "Not Found",
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}

func TestSIPServerTimeout(t *testing.T) {
	server, err := newMockServer(sip.OPTIONS, func(_ *sip.Request, _ sip.ServerTransaction) {
		// Intentionally no response to trigger timeout
	})
	require.NoError(t, err)
	defer server.close()

	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "OPTIONS",
		Timeout:  config.Duration(100 * time.Millisecond),
		FromUser: "telegraf",
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	// Use RequireMetricsEqual for tags (no status_code on timeout)
	expected := []telegraf.Metric{
		metric.New(
			"sip",
			map[string]string{
				"source":    "sip://" + server.addr,
				"method":    "options",
				"transport": "udp",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
				"result":          "Timeout",
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))

	// Additionally verify response_time_s equals timeout value
	require.Len(t, acc.Metrics, 1)
	rt, ok := acc.FloatField("sip", "response_time_s")
	require.True(t, ok)
	require.InDelta(t, 0.1, rt, 0.01, "response_time_s should equal timeout value")
}

func TestSIPServerDelayedResponse(t *testing.T) {
	server, err := newMockServer(sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		time.Sleep(50 * time.Millisecond)
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	require.NoError(t, err)
	defer server.close()

	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "OPTIONS",
		Timeout:  config.Duration(200 * time.Millisecond),
		FromUser: "telegraf",
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	// Use RequireMetricsEqual for tags and fields structure
	expected := []telegraf.Metric{
		metric.New(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "200",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
				"result":          "OK",
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))

	// Additionally verify response time is within expected range
	require.Len(t, acc.Metrics, 1)
	rt := acc.Metrics[0].Fields["response_time_s"].(float64)
	require.Greater(t, rt, 0.05, "response time should be at least 50ms")
	require.Less(t, rt, 0.2, "response time should be less than timeout")
}

func TestSIPDifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		reason     string
	}{
		{
			name:       "200_ok",
			statusCode: 200,
			reason:     "OK",
		},
		{
			name:       "404_not_found",
			statusCode: 404,
			reason:     "Not Found",
		},
		{
			name:       "503_service_unavailable",
			statusCode: 503,
			reason:     "Service Unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := newMockServer(sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
				res := sip.NewResponseFromRequest(req, tt.statusCode, tt.reason, nil)
				require.NoError(t, tx.Respond(res))
			})
			require.NoError(t, err)
			defer server.close()

			plugin := &SIP{
				Server:   "sip://" + server.addr,
				Method:   "OPTIONS",
				Timeout:  config.Duration(2 * time.Second),
				FromUser: "telegraf",
				Log:      testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			require.NoError(t, plugin.Gather(&acc))

			expected := []telegraf.Metric{
				metric.New(
					"sip",
					map[string]string{
						"source":      "sip://" + server.addr,
						"method":      "options",
						"transport":   "udp",
						"status_code": strconv.Itoa(tt.statusCode),
					},
					map[string]interface{}{
						"response_time_s": float64(0),
						"result":          tt.reason,
					},
					time.Now(),
				),
			}
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
		})
	}
}

func TestSIPAuthenticationRequired(t *testing.T) {
	server, err := newMockServer(sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		// Respond with 401 Unauthorized to require authentication
		res := sip.NewResponseFromRequest(req, 401, "Unauthorized", nil)
		// Add WWW-Authenticate header (required for digest auth)
		res.AppendHeader(sip.NewHeader("WWW-Authenticate", `Digest realm="test", nonce="abc123"`))
		require.NoError(t, tx.Respond(res))
	})
	require.NoError(t, err)
	defer server.close()

	// Test without credentials - should get auth_required
	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "OPTIONS",
		Timeout:  config.Duration(2 * time.Second),
		FromUser: "telegraf",
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "401",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
				"result":          "Unauthorized",
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}

func TestSIPAuthenticationSuccess(t *testing.T) {
	const (
		validUsername = "alice"
		validPassword = "secret123"
	)

	attemptCount := 0
	server, err := newMockServer(sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		attemptCount++

		// Check if Authorization header is present
		authHeader := req.GetHeader("Authorization")

		if authHeader == nil {
			// First attempt without auth - send 401 challenge
			res := sip.NewResponseFromRequest(req, 401, "Unauthorized", nil)
			res.AppendHeader(sip.NewHeader("WWW-Authenticate", `Digest realm="test", nonce="abc123", algorithm=MD5`))
			require.NoError(t, tx.Respond(res))
			return
		}

		// Second attempt with auth - validate it exists and respond with 200
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	require.NoError(t, err)
	defer server.close()

	// Create plugin with valid credentials
	username := config.NewSecret([]byte(validUsername))
	password := config.NewSecret([]byte(validPassword))

	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "OPTIONS",
		Timeout:  config.Duration(2 * time.Second),
		FromUser: "telegraf",
		Username: username,
		Password: password,
		Log:      testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	// Verify server was called twice (initial request + auth retry)
	require.Equal(t, 2, attemptCount, "server should be called twice: initial + auth retry")

	// Verify successful authentication
	expected := []telegraf.Metric{
		metric.New(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "200",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
				"result":          "OK",
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))

	// SECURITY: Verify credentials never appear in tags or fields
	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]
	for k, v := range m.Tags {
		require.NotContains(t, v, validUsername, "tag %s must not contain username", k)
		require.NotContains(t, v, validPassword, "tag %s must not contain password", k)
	}
	for k, v := range m.Fields {
		s := fmt.Sprintf("%v", v)
		require.NotContains(t, s, validUsername, "field %s must not contain username", k)
		require.NotContains(t, s, validPassword, "field %s must not contain password", k)
	}
}

func TestSIPMethodINVITE(t *testing.T) {
	server, err := newMockServer(sip.INVITE, func(req *sip.Request, tx sip.ServerTransaction) {
		// Verify we received an INVITE request
		require.Equal(t, "INVITE", req.Method.String())

		// INVITE typically gets a 200 OK or 180 Ringing
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	require.NoError(t, err)
	defer server.close()

	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "INVITE",
		Timeout:  config.Duration(2 * time.Second),
		FromUser: "telegraf",
		Log:      testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "invite",
				"transport":   "udp",
				"status_code": "200",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
				"result":          "OK",
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}

func TestSIPMethodMESSAGE(t *testing.T) {
	server, err := newMockServer(sip.MESSAGE, func(req *sip.Request, tx sip.ServerTransaction) {
		// Verify we received a MESSAGE request
		require.Equal(t, "MESSAGE", req.Method.String())

		// MESSAGE typically gets a 200 OK or 202 Accepted
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	require.NoError(t, err)
	defer server.close()

	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "MESSAGE",
		Timeout:  config.Duration(2 * time.Second),
		FromUser: "telegraf",
		Log:      testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "message",
				"transport":   "udp",
				"status_code": "200",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
				"result":          "OK",
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}

// Mock server utilities

type mockServer struct {
	ua     *sipgo.UserAgent
	server *sipgo.Server
	addr   string
}

func newMockServer(method sip.RequestMethod, handler func(*sip.Request, sip.ServerTransaction)) (*mockServer, error) {
	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent("Test SIP Server"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating user agent: %w", err)
	}

	server, err := sipgo.NewServer(ua)
	if err != nil {
		_ = ua.Close()
		return nil, fmt.Errorf("creating server: %w", err)
	}

	// Register handler for the specified method
	server.OnRequest(method, handler)

	// Create UDP listener ourselves to know the address before serving.
	// This avoids a data race in sipgo where GetListenPort reads the
	// transport layer map without holding the lock.
	udpConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		_ = server.Close()
		_ = ua.Close()
		return nil, fmt.Errorf("listening udp: %w", err)
	}
	addr := udpConn.LocalAddr().String()

	go func() {
		//nolint:errcheck // Background server for testing, errors are not critical
		server.ServeUDP(udpConn)
	}()

	return &mockServer{
		ua:     ua,
		server: server,
		addr:   addr,
	}, nil
}

func (s *mockServer) close() {
	_ = s.server.Close()
	_ = s.ua.Close()
}
