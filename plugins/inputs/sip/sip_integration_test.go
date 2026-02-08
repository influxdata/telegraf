//go:build integration

package sip

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

type testSIPServer struct {
	ua     *sipgo.UserAgent
	server *sipgo.Server
	addr   string
}

func pickFreeUDPAddr(t *testing.T) string {
	t.Helper()

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := conn.LocalAddr().String()
	require.NoError(t, conn.Close())

	return addr
}

func startTestSIPServerForMethod(
	t *testing.T,
	method sip.RequestMethod,
	handler func(req *sip.Request, tx sip.ServerTransaction),
) *testSIPServer {
	t.Helper()

	addr := pickFreeUDPAddr(t)

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
		server.ListenAndServe(ctx, "udp", addr)
	}()

	// Wait for server to be ready
	<-serverReady

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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// Check tags
	require.Equal(t, "sip://"+server.addr, m.Tags["server"])
	require.Equal(t, "udp", m.Tags["transport"])
	require.Equal(t, "OPTIONS", m.Tags["method"])
	require.Equal(t, "200", m.Tags["status_code"])

	// Check fields
	require.Equal(t, 1, m.Fields["up"])
	rt, ok := m.Fields["response_time"].(float64)
	require.True(t, ok)
	require.GreaterOrEqual(t, rt, 0.0)
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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// Check tags
	require.Equal(t, "sip://"+server.addr, m.Tags["server"])
	require.Equal(t, "udp", m.Tags["transport"])
	require.Equal(t, "404", m.Tags["status_code"])

	// Check fields
	require.Equal(t, 0, m.Fields["up"])
	rt, ok := m.Fields["response_time"].(float64)
	require.True(t, ok)
	require.GreaterOrEqual(t, rt, 0.0)
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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// Check tags
	require.Equal(t, "sip://"+server.addr, m.Tags["server"])
	require.Equal(t, "udp", m.Tags["transport"])

	// Check fields - timeout should have up field but no response_time
	require.Equal(t, 0, m.Fields["up"])
	_, hasResponseTime := m.Fields["response_time"]
	require.False(t, hasResponseTime)
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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	// Verify response time is within expected range
	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	rt := m.Fields["response_time"].(float64)
	require.Greater(t, rt, 0.3, "response time should be at least 300ms")
	require.Less(t, rt, 1.0, "response time should be less than timeout")

	// Check remaining fields and tags
	require.Equal(t, "sip://"+server.addr, m.Tags["server"])
	require.Equal(t, "udp", m.Tags["transport"])
	require.Equal(t, "200", m.Tags["status_code"])
	require.Equal(t, 1, m.Fields["up"])
}

func TestSIPDifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		reason     string
		wantUp     int
	}{
		{
			name:       "200_ok",
			statusCode: 200,
			reason:     "OK",
			wantUp:     1,
		},
		{
			name:       "404_not_found",
			statusCode: 404,
			reason:     "Not Found",
			wantUp:     0,
		},
		{
			name:       "503_service_unavailable",
			statusCode: 503,
			reason:     "Service Unavailable",
			wantUp:     0,
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

			acc := &testutil.Accumulator{}
			require.NoError(t, plugin.Start(acc))
			defer plugin.Stop()

			require.NoError(t, plugin.Gather(acc))

			require.Len(t, acc.Metrics, 1)
			m := acc.Metrics[0]
			require.Equal(t, tt.wantUp, m.Fields["up"])
			require.Equal(t, strconv.Itoa(tt.statusCode), m.Tags["status_code"])
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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// Check tags
	require.Equal(t, "sip://"+server.addr, m.Tags["server"])
	require.Equal(t, "udp", m.Tags["transport"])
	require.Equal(t, "401", m.Tags["status_code"])

	// Check fields
	require.Equal(t, 0, m.Fields["up"])
	rt, ok := m.Fields["response_time"].(float64)
	require.True(t, ok)
	require.GreaterOrEqual(t, rt, 0.0)
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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	// Verify server was called twice (initial request + auth retry)
	require.Equal(t, 2, attemptCount, "server should be called twice: initial + auth retry")

	// Verify successful authentication
	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// Check tags
	require.Equal(t, "sip://"+server.addr, m.Tags["server"])
	require.Equal(t, "udp", m.Tags["transport"])
	require.Equal(t, "200", m.Tags["status_code"])

	// Check fields
	require.Equal(t, 1, m.Fields["up"])
	rt, ok := m.Fields["response_time"].(float64)
	require.True(t, ok)
	require.GreaterOrEqual(t, rt, 0.0)

	// SECURITY: Verify credentials never appear in tags
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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// SECURITY CHECK: Verify all tags don't contain credentials
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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// Check tags
	require.Equal(t, "sip://"+server.addr, m.Tags["server"])
	require.Equal(t, "udp", m.Tags["transport"])
	require.Equal(t, "INVITE", m.Tags["method"])
	require.Equal(t, "200", m.Tags["status_code"])

	// Check fields
	require.Equal(t, 1, m.Fields["up"])
	rt, ok := m.Fields["response_time"].(float64)
	require.True(t, ok)
	require.GreaterOrEqual(t, rt, 0.0)
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

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// Check tags
	require.Equal(t, "sip://"+server.addr, m.Tags["server"])
	require.Equal(t, "udp", m.Tags["transport"])
	require.Equal(t, "MESSAGE", m.Tags["method"])
	require.Equal(t, "200", m.Tags["status_code"])

	// Check fields
	require.Equal(t, 1, m.Fields["up"])
	rt, ok := m.Fields["response_time"].(float64)
	require.True(t, ok)
	require.GreaterOrEqual(t, rt, 0.0)
}
