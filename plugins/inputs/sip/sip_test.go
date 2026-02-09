package sip

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

type testSIPServer struct {
	ua     *sipgo.UserAgent
	server *sipgo.Server
	addr   string
}

func startTestSIPServerForMethod(
	t *testing.T,
	method sip.RequestMethod,
	handler func(req *sip.Request, tx sip.ServerTransaction),
) *testSIPServer {
	t.Helper()

	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent("Test SIP Server"),
	)
	require.NoError(t, err)

	server, err := sipgo.NewServer(ua)
	require.NoError(t, err)

	// Register handler for the specified method
	server.OnRequest(method, handler)

	// Use sipgo's test context key to signal when server is ready
	serverReady := make(chan struct{})
	//nolint:staticcheck // SA1029: sipgo.ListenReadyCtxKey is a string constant defined by sipgo library
	ctx := context.WithValue(context.Background(), sipgo.ListenReadyCtxKey, sipgo.ListenReadyCtxValue(serverReady))

	go func() {
		//nolint:errcheck // Background server for testing, errors are not critical
		server.ListenAndServe(ctx, "udp", "127.0.0.1:0")
	}()

	// Wait for server to be ready
	<-serverReady

	// Get the actual port the server is listening on
	port := server.TransportLayer().GetListenPort("udp")
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	return &testSIPServer{
		ua:     ua,
		server: server,
		addr:   addr,
	}
}

func (s *testSIPServer) close() {
	_ = s.server.Close()
	_ = s.ua.Close()
}

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
	validMethods := []string{"OPTIONS", "INVITE", "MESSAGE", "options", "invite", "message"}

	for _, method := range validMethods {
		t.Run(method, func(t *testing.T) {
			plugin := &SIP{
				Server:  "sip://sip.example.com:5060",
				Method:  method,
				Timeout: config.Duration(5 * time.Second),
			}

			require.NoError(t, plugin.Init())
			require.Equal(t, strings.ToUpper(method), plugin.Method)
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
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
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
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "200",
				"result":      "OK",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}

func TestSIPServerErrorResponse(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 404, "Not Found", nil)
		require.NoError(t, tx.Respond(res))
	})
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
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "404",
				"result":      "Not Found",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}

func TestSIPServerTimeout(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(_ *sip.Request, _ sip.ServerTransaction) {
		// Intentionally no response to trigger timeout
	})
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
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":    "sip://" + server.addr,
				"method":    "options",
				"transport": "udp",
				"result":    "timeout",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))

	// Additionally verify response_time_s equals timeout value
	require.Len(t, acc.Metrics, 1)
	rt, ok := acc.Metrics[0].Fields["response_time_s"].(float64)
	require.True(t, ok)
	require.InDelta(t, 0.1, rt, 0.01, "response_time_s should equal timeout value")
}

func TestSIPServerDelayedResponse(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		time.Sleep(300 * time.Millisecond)
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server.close()

	plugin := &SIP{
		Server:   "sip://" + server.addr,
		Method:   "OPTIONS",
		Timeout:  config.Duration(1 * time.Second),
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
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "200",
				"result":      "OK",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))

	// Additionally verify response time is within expected range
	require.Len(t, acc.Metrics, 1)
	rt := acc.Metrics[0].Fields["response_time_s"].(float64)
	require.Greater(t, rt, 0.3, "response time should be at least 300ms")
	require.Less(t, rt, 1.0, "response time should be less than timeout")
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
			server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
				res := sip.NewResponseFromRequest(req, tt.statusCode, tt.reason, nil)
				require.NoError(t, tx.Respond(res))
			})
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
				testutil.MustMetric(
					"sip",
					map[string]string{
						"source":      "sip://" + server.addr,
						"method":      "options",
						"transport":   "udp",
						"status_code": strconv.Itoa(tt.statusCode),
						"result":      tt.reason,
					},
					map[string]interface{}{
						"response_time_s": float64(0),
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
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		// Respond with 401 Unauthorized to require authentication
		res := sip.NewResponseFromRequest(req, 401, "Unauthorized", nil)
		// Add WWW-Authenticate header (required for digest auth)
		res.AppendHeader(sip.NewHeader("WWW-Authenticate", `Digest realm="test", nonce="abc123"`))
		require.NoError(t, tx.Respond(res))
	})
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
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "401",
				"result":      "Unauthorized",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
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
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
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
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "200",
				"result":      "OK",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))

	// SECURITY: Verify credentials never appear in tags
	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]
	for k, v := range m.Tags {
		require.NotContains(t, v, validUsername, "tag %s must not contain username", k)
		require.NotContains(t, v, validPassword, "tag %s must not contain password", k)
	}
}

func TestSIPCredentialsNotInTags(t *testing.T) {
	// This test verifies that username/password never appear in tags
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server.close()

	// Create plugin with credentials
	username := config.NewSecret([]byte("testuser"))
	password := config.NewSecret([]byte("testpass"))

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

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "options",
				"transport":   "udp",
				"status_code": "200",
				"result":      "OK",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))

	// SECURITY CHECK: Verify all tags don't contain credentials
	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]
	for k, v := range m.Tags {
		require.NotContains(t, v, "testuser", "tag %s must not contain username", k)
		require.NotContains(t, v, "testpass", "tag %s must not contain password", k)
	}
}

func TestSIPMethodINVITE(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.INVITE, func(req *sip.Request, tx sip.ServerTransaction) {
		// Verify we received an INVITE request
		require.Equal(t, "INVITE", req.Method.String())

		// INVITE typically gets a 200 OK or 180 Ringing
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
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
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "invite",
				"transport":   "udp",
				"status_code": "200",
				"result":      "OK",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}

func TestSIPMethodMESSAGE(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.MESSAGE, func(req *sip.Request, tx sip.ServerTransaction) {
		// Verify we received a MESSAGE request
		require.Equal(t, "MESSAGE", req.Method.String())

		// MESSAGE typically gets a 200 OK or 202 Accepted
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
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
		testutil.MustMetric(
			"sip",
			map[string]string{
				"source":      "sip://" + server.addr,
				"method":      "message",
				"transport":   "udp",
				"status_code": "200",
				"result":      "OK",
			},
			map[string]interface{}{
				"response_time_s": float64(0),
			},
			time.Now(),
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("response_time_s"))
}
