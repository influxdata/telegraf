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

	go func() {
		//nolint:errcheck // Background server for testing, errors are not critical
		server.ListenAndServe(context.Background(), "udp", addr)
	}()

	// Give sipgo time to bind
	time.Sleep(50 * time.Millisecond)

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

func assertSingleSIPMetric(
	t *testing.T,
	acc *testutil.Accumulator,
	expectedTags map[string]string,
	expectedFields map[string]any,
) {
	t.Helper()

	require.Len(t, acc.Metrics, 1)

	m := acc.Metrics[0]
	require.Equal(t, "sip", m.Measurement)

	// Check all expected tags
	for k, v := range expectedTags {
		require.Equal(t, v, m.Tags[k], "tag %s mismatch", k)
	}

	// Check all expected fields
	for k, v := range expectedFields {
		if v == nil {
			// nil means validate existence and type only (used for response_time)
			rt, ok := m.Fields[k].(float64)
			require.True(t, ok, "%s must be float64", k)
			require.Greater(t, rt, 0.0, "%s must be positive", k)
		} else {
			require.Equal(t, v, m.Fields[k], "field %s mismatch", k)
		}
	}
}

func newTestPlugin(serverAddr string, timeout time.Duration) *SIP {
	return &SIP{
		Servers:    []string{"sip://" + serverAddr},
		Method:     "OPTIONS",
		Timeout:    config.Duration(timeout),
		FromUser:   "telegraf",
		UserAgent:  "Telegraf SIP Monitor",
		ExpectCode: 200,
		Log:        testutil.Logger{},
	}
}

func TestSIPServerSuccess(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server.close()

	plugin := newTestPlugin(server.addr, 2*time.Second)
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	assertSingleSIPMetric(t, acc,
		map[string]string{
			"server":      "sip://" + server.addr,
			"transport":   "udp",
			"status_code": "200",
			"reason":      "OK",
			"result":      "success",
			"sip_uri":     "sip:telegraf@" + server.addr,
		},
		map[string]any{
			"result_type":         "success",
			"result_code":         int(0),
			"response_code_match": int(1),
			"response_time":       nil, // nil means validate existence and positive value
		},
	)
}

func TestSIPServerResponseCodeMismatch(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 404, "Not Found", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server.close()

	plugin := newTestPlugin(server.addr, 2*time.Second)
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	assertSingleSIPMetric(t, acc,
		map[string]string{
			"server":      "sip://" + server.addr,
			"transport":   "udp",
			"status_code": "404",
			"reason":      "Not Found",
			"result":      "response_code_mismatch",
			"sip_uri":     "sip:telegraf@" + server.addr,
		},
		map[string]any{
			"result_type":         "response_code_mismatch",
			"result_code":         int(1),
			"response_code_match": int(0),
			"response_time":       nil, // nil means validate existence and positive value
		},
	)
}

func TestSIPServerTimeout(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(_ *sip.Request, _ sip.ServerTransaction) {
		// Intentionally no response to trigger timeout
	})
	defer server.close()

	plugin := newTestPlugin(server.addr, 100*time.Millisecond)
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	assertSingleSIPMetric(t, acc,
		map[string]string{
			"server":    "sip://" + server.addr,
			"transport": "udp",
			"result":    "timeout",
			"sip_uri":   "sip:telegraf@" + server.addr,
		},
		map[string]any{
			"result_type": "timeout",
			"result_code": int(2),
		},
	)
}

func TestSIPServerDelayedResponse(t *testing.T) {
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		time.Sleep(300 * time.Millisecond)
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server.close()

	plugin := newTestPlugin(server.addr, 1*time.Second)
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	// Verify response time is within expected range
	require.Len(t, acc.Metrics, 1)
	rt := acc.Metrics[0].Fields["response_time"].(float64)
	require.Greater(t, rt, 0.3, "response time should be at least 300ms")
	require.Less(t, rt, 1.0, "response time should be less than timeout")

	assertSingleSIPMetric(t, acc,
		map[string]string{
			"server":      "sip://" + server.addr,
			"transport":   "udp",
			"status_code": "200",
			"reason":      "OK",
			"result":      "success",
			"sip_uri":     "sip:telegraf@" + server.addr,
		},
		map[string]any{
			"result_type":         "success",
			"result_code":         int(0),
			"response_code_match": int(1),
			"response_time":       nil, // Already validated above with specific bounds
		},
	)
}

func TestSIPMultipleServers(t *testing.T) {
	server1 := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server1.close()

	server2 := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 503, "Service Unavailable", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server2.close()

	plugin := &SIP{
		Servers:    []string{"sip://" + server1.addr, "sip://" + server2.addr},
		Method:     "OPTIONS",
		Timeout:    config.Duration(2 * time.Second),
		ExpectCode: 200,
		Log:        testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	// Should have metrics for both servers
	require.Len(t, acc.Metrics, 2)

	// Verify first server succeeded
	var found1, found2 bool
	for _, m := range acc.Metrics {
		if m.Tags["server"] == "sip://"+server1.addr {
			found1 = true
			require.Equal(t, "success", m.Tags["result"])
			require.Equal(t, "200", m.Tags["status_code"])
		}
		if m.Tags["server"] == "sip://"+server2.addr {
			found2 = true
			require.Equal(t, "response_code_mismatch", m.Tags["result"])
			require.Equal(t, "503", m.Tags["status_code"])
		}
	}
	require.True(t, found1, "server1 metric not found")
	require.True(t, found2, "server2 metric not found")
}

func TestSIPDifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		reason     string
		expectCode int
		wantResult string
	}{
		{
			name:       "100_trying",
			statusCode: 100,
			reason:     "Trying",
			expectCode: 200,
			wantResult: "response_code_mismatch",
		},
		{
			name:       "180_ringing",
			statusCode: 180,
			reason:     "Ringing",
			expectCode: 200,
			wantResult: "response_code_mismatch",
		},
		{
			name:       "503_service_unavailable",
			statusCode: 503,
			reason:     "Service Unavailable",
			expectCode: 200,
			wantResult: "response_code_mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
				res := sip.NewResponseFromRequest(req, tt.statusCode, tt.reason, nil)
				require.NoError(t, tx.Respond(res))
			})
			defer server.close()

			plugin := newTestPlugin(server.addr, 2*time.Second)
			require.NoError(t, plugin.Init())

			acc := &testutil.Accumulator{}
			require.NoError(t, plugin.Start(acc))
			defer plugin.Stop()

			require.NoError(t, plugin.Gather(acc))

			require.Len(t, acc.Metrics, 1)
			m := acc.Metrics[0]
			require.Equal(t, tt.wantResult, m.Tags["result"])
			require.Equal(t, strconv.Itoa(tt.statusCode), m.Tags["status_code"])
			require.Equal(t, tt.reason, m.Tags["reason"])
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
	plugin := newTestPlugin(server.addr, 2*time.Second)
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	assertSingleSIPMetric(t, acc,
		map[string]string{
			"server":      "sip://" + server.addr,
			"transport":   "udp",
			"result":      "auth_required",
			"sip_uri":     "sip:telegraf@" + server.addr,
			"status_code": "401",
			"reason":      "Unauthorized",
		},
		map[string]any{
			"result_type":   "auth_required",
			"result_code":   int(12),
			"response_time": nil,
		},
	)
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
		// Note: We don't fully validate the digest here as that would require
		// implementing full digest validation in the test, but the sipgo client
		// will have computed the correct digest
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server.close()

	// Create plugin with valid credentials
	username := config.NewSecret([]byte(validUsername))
	password := config.NewSecret([]byte(validPassword))

	plugin := &SIP{
		Servers:    []string{"sip://" + server.addr},
		Method:     "OPTIONS",
		Timeout:    config.Duration(2 * time.Second),
		FromUser:   "telegraf",
		UserAgent:  "Telegraf SIP Monitor",
		ExpectCode: 200,
		Username:   username,
		Password:   password,
		Log:        testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	// Verify server was called twice (initial request + auth retry)
	require.Equal(t, 2, attemptCount, "server should be called twice: initial + auth retry")

	// Verify successful authentication
	assertSingleSIPMetric(t, acc,
		map[string]string{
			"server":      "sip://" + server.addr,
			"transport":   "udp",
			"status_code": "200",
			"reason":      "OK",
			"result":      "success",
			"sip_uri":     "sip:telegraf@" + server.addr,
		},
		map[string]any{
			"result_type":         "success",
			"result_code":         int(0),
			"response_code_match": int(1),
			"response_time":       nil,
		},
	)

	// SECURITY: Verify credentials never appear in the sip_uri
	m := acc.Metrics[0]
	sipURI := m.Tags["sip_uri"]
	require.NotContains(t, sipURI, validUsername, "sip_uri must not contain username")
	require.NotContains(t, sipURI, validPassword, "sip_uri must not contain password")
}

func TestSIPCredentialsNotInURI(t *testing.T) {
	// This test verifies that username/password never appear in the sip_uri tag
	server := startTestSIPServerForMethod(t, sip.OPTIONS, func(req *sip.Request, tx sip.ServerTransaction) {
		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		require.NoError(t, tx.Respond(res))
	})
	defer server.close()

	// Create plugin with credentials
	username := config.NewSecret([]byte("testuser"))
	password := config.NewSecret([]byte("testpass"))

	plugin := &SIP{
		Servers:    []string{"sip://" + server.addr},
		Method:     "OPTIONS",
		Timeout:    config.Duration(2 * time.Second),
		FromUser:   "telegraf",
		UserAgent:  "Telegraf SIP Monitor",
		ExpectCode: 200,
		Username:   username,
		Password:   password,
		Log:        testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]

	// SECURITY CHECK: Verify sip_uri does NOT contain credentials
	sipURI := m.Tags["sip_uri"]
	require.NotContains(t, sipURI, "testuser", "sip_uri must not contain username")
	require.NotContains(t, sipURI, "testpass", "sip_uri must not contain password")
	require.Equal(t, "sip:telegraf@"+server.addr, sipURI, "sip_uri should only contain FromUser, not authentication credentials")

	// Verify all tags don't contain credentials
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
		Servers:    []string{"sip://" + server.addr},
		Method:     "INVITE",
		Timeout:    config.Duration(2 * time.Second),
		FromUser:   "telegraf",
		UserAgent:  "Telegraf SIP Monitor",
		ExpectCode: 200,
		Log:        testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	assertSingleSIPMetric(t, acc,
		map[string]string{
			"server":      "sip://" + server.addr,
			"transport":   "udp",
			"method":      "INVITE",
			"status_code": "200",
			"reason":      "OK",
			"result":      "success",
			"sip_uri":     "sip:telegraf@" + server.addr,
		},
		map[string]any{
			"result_type":         "success",
			"result_code":         int(0),
			"response_code_match": int(1),
			"response_time":       nil,
		},
	)
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
		Servers:    []string{"sip://" + server.addr},
		Method:     "MESSAGE",
		Timeout:    config.Duration(2 * time.Second),
		FromUser:   "telegraf",
		UserAgent:  "Telegraf SIP Monitor",
		ExpectCode: 200,
		Log:        testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(acc))

	assertSingleSIPMetric(t, acc,
		map[string]string{
			"server":      "sip://" + server.addr,
			"transport":   "udp",
			"method":      "MESSAGE",
			"status_code": "200",
			"reason":      "OK",
			"result":      "success",
			"sip_uri":     "sip:telegraf@" + server.addr,
		},
		map[string]any{
			"result_type":         "success",
			"result_code":         int(0),
			"response_code_match": int(1),
			"response_time":       nil,
		},
	)
}
