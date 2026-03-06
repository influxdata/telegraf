package warp10

import (
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestWriteWarp10(t *testing.T) {
	w := Warp10{
		Prefix:  "unit.test",
		WarpURL: "http://localhost:8090",
		Token:   config.NewSecret([]byte("WRITE")),
	}

	payload := w.GenWarp10Payload(testutil.MockMetrics())
	require.Exactly(t, "1257894000000000// unit.testtest1.value{source=telegraf,tag1=value1} 1.000000\n", payload)
}

func TestWriteWarp10ValueNaN(t *testing.T) {
	w := Warp10{
		Prefix:  "unit.test",
		WarpURL: "http://localhost:8090",
		Token:   config.NewSecret([]byte("WRITE")),
	}

	payload := w.GenWarp10Payload(testutil.MockMetricsWithValue(math.NaN()))
	require.Exactly(t, "1257894000000000// unit.testtest1.value{source=telegraf,tag1=value1} NaN\n", payload)
}

func TestWriteWarp10ValueInfinity(t *testing.T) {
	w := Warp10{
		Prefix:  "unit.test",
		WarpURL: "http://localhost:8090",
		Token:   config.NewSecret([]byte("WRITE")),
	}

	payload := w.GenWarp10Payload(testutil.MockMetricsWithValue(math.Inf(1)))
	require.Exactly(t, "1257894000000000// unit.testtest1.value{source=telegraf,tag1=value1} Infinity\n", payload)
}

func TestWriteWarp10ValueMinusInfinity(t *testing.T) {
	w := Warp10{
		Prefix:  "unit.test",
		WarpURL: "http://localhost:8090",
		Token:   config.NewSecret([]byte("WRITE")),
	}

	payload := w.GenWarp10Payload(testutil.MockMetricsWithValue(math.Inf(-1)))
	require.Exactly(t, "1257894000000000// unit.testtest1.value{source=telegraf,tag1=value1} -Infinity\n", payload)
}

func TestWriteWarp10EncodedTags(t *testing.T) {
	w := Warp10{
		Prefix:  "unit.test",
		WarpURL: "http://localhost:8090",
		Token:   config.NewSecret([]byte("WRITE")),
	}

	metrics := testutil.MockMetrics()
	for _, metric := range metrics {
		metric.AddTag("encoded{tag", "value1,value2")
	}

	payload := w.GenWarp10Payload(metrics)
	require.Exactly(t, "1257894000000000// unit.testtest1.value{encoded%7Btag=value1%2Cvalue2,source=telegraf,tag1=value1} 1.000000\n", payload)
}

func TestHandleWarp10Error(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		expected  string
		isAuthErr bool
	}{
		{
			name: "invalid token",
			message: `
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html;charset=utf-8"/>
			<title>Error 500 io.warp10.script.WarpScriptException: Invalid token.</title>
			</head>
			<body><h2>HTTP ERROR 500</h2>
			<p>Problem accessing /api/v0/update. Reason:
			<pre>    io.warp10.script.WarpScriptException: Invalid token.</pre></p>
			</body>
			</html>
			`,
			expected:  "Invalid token",
			isAuthErr: true,
		},
		{
			name: "token expired",
			message: `
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html;charset=utf-8"/>
			<title>Error 500 io.warp10.script.WarpScriptException: Token Expired.</title>
			</head>
			<body><h2>HTTP ERROR 500</h2>
			<p>Problem accessing /api/v0/update. Reason:
			<pre>    io.warp10.script.WarpScriptException: Token Expired.</pre></p>
			</body>
			</html>
			`,
			expected:  "Token Expired",
			isAuthErr: true,
		},
		{
			name: "token revoked",
			message: `
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html;charset=utf-8"/>
			<title>Error 500 io.warp10.script.WarpScriptException: Token revoked.</title>
			</head>
			<body><h2>HTTP ERROR 500</h2>
			<p>Problem accessing /api/v0/update. Reason:
			<pre>    io.warp10.script.WarpScriptException: Token revoked.</pre></p>
			</body>
			</html>
			`,
			expected:  "Token revoked",
			isAuthErr: true,
		},
		{
			name: "write token missing",
			message: `
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html;charset=utf-8"/>
			<title>Error 500 io.warp10.script.WarpScriptException: Write token missing.</title>
			</head>
			<body><h2>HTTP ERROR 500</h2>
			<p>Problem accessing /api/v0/update. Reason:
			<pre>    io.warp10.script.WarpScriptException: Write token missing.</pre></p>
			</body>
			</html>
			`,
			expected:  "Write token missing",
			isAuthErr: true,
		},
		{
			name:      "server unavailable (retryable)",
			message:   `<title>Error 503: server unavailable</title>`,
			expected:  "<title>Error 503: server unavailable</title>",
			isAuthErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleError(tt.message, 511)
			require.EqualError(t, err, tt.expected)
			var authErr *tokenAuthError
			if tt.isAuthErr {
				require.ErrorAs(t, err, &authErr)
			} else {
				require.False(t, errors.As(err, &authErr))
			}
		})
	}
}

func newTestServer(t *testing.T, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if _, err := w.Write([]byte(body)); err != nil {
			t.Log(err)
			t.Fail()
			return
		}
	}))
}

func TestAuthFailure_SameTokenBlocksWrites(t *testing.T) {
	ts := newTestServer(t, http.StatusInternalServerError, "Invalid token")
	defer ts.Close()

	w := &Warp10{
		WarpURL:          ts.URL,
		Token:            config.NewSecret([]byte("BAD_TOKEN")),
		AuthErrorRetries: 3,
		PrintErrorBody:   true,
		Log:              testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())
	defer w.Close()

	metrics := testutil.MockMetrics()

	// First write: triggers auth error, sets failure state
	require.ErrorContains(t, w.Write(metrics), "Invalid token")
	require.Equal(t, "BAD_TOKEN", w.failedToken)
	require.Equal(t, uint(3), w.authRetriesLeft)

	// Next 3 writes: silently dropped (countdown 2, 1, 0)
	require.NoError(t, w.Write(metrics))
	require.Equal(t, uint(2), w.authRetriesLeft)
	require.NoError(t, w.Write(metrics))
	require.Equal(t, uint(1), w.authRetriesLeft)
	require.NoError(t, w.Write(metrics))
	require.Equal(t, uint(0), w.authRetriesLeft)

	// 4th write after initial: retries (countdown reached 0)
	require.ErrorContains(t, w.Write(metrics), "Invalid token")
	require.Equal(t, uint(3), w.authRetriesLeft)
}

func TestAuthFailure_TokenChangeResumesWrites(t *testing.T) {
	ts := newTestServer(t, http.StatusOK, "")
	defer ts.Close()

	w := &Warp10{
		WarpURL:          ts.URL,
		Token:            config.NewSecret([]byte("NEW_TOKEN")),
		AuthErrorRetries: 10,
		Log:              testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())
	defer w.Close()

	// Simulate prior failure state with a different token
	w.failedToken = "OLD_TOKEN"
	w.authRetriesLeft = 5

	// Write should succeed because token changed
	require.NoError(t, w.Write(testutil.MockMetrics()))
	require.Empty(t, w.failedToken)
	require.Equal(t, uint(0), w.authRetriesLeft)
}

func TestAuthFailure_DefaultZeroRetriesEveryInterval(t *testing.T) {
	ts := newTestServer(t, http.StatusInternalServerError, "Invalid token")
	defer ts.Close()

	w := &Warp10{
		WarpURL:          ts.URL,
		Token:            config.NewSecret([]byte("BAD_TOKEN")),
		AuthErrorRetries: 0, // default
		PrintErrorBody:   true,
		Log:              testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())
	defer w.Close()

	metrics := testutil.MockMetrics()

	// Every write should attempt the request (retry every flush)
	require.ErrorContains(t, w.Write(metrics), "Invalid token")
	require.ErrorContains(t, w.Write(metrics), "Invalid token")
	require.ErrorContains(t, w.Write(metrics), "Invalid token")
}

func TestAuthFailure_SuccessClearsFailureState(t *testing.T) {
	ts := newTestServer(t, http.StatusOK, "")
	defer ts.Close()

	w := &Warp10{
		WarpURL:          ts.URL,
		Token:            config.NewSecret([]byte("GOOD_TOKEN")),
		AuthErrorRetries: 5,
		Log:              testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())
	defer w.Close()

	// Simulate prior failure state with same token
	w.failedToken = "GOOD_TOKEN"
	w.authRetriesLeft = 0 // countdown exhausted, will retry

	// Write succeeds
	require.NoError(t, w.Write(testutil.MockMetrics()))
	require.Empty(t, w.failedToken)
	require.Equal(t, uint(0), w.authRetriesLeft)
}

func TestAuthFailure_RetryableErrorDoesNotTrackAuthFailure(t *testing.T) {
	ts := newTestServer(t, http.StatusInternalServerError, "broken pipe")
	defer ts.Close()

	w := &Warp10{
		WarpURL:          ts.URL,
		Token:            config.NewSecret([]byte("GOOD_TOKEN")),
		AuthErrorRetries: 3,
		PrintErrorBody:   true,
		Log:              testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())
	defer w.Close()

	require.ErrorContains(t, w.Write(testutil.MockMetrics()), "broken pipe")
	require.Empty(t, w.failedToken)
	require.Equal(t, uint(0), w.authRetriesLeft)
}

func TestAuthFailure_PrintErrorBodyStillTracksAuthFailure(t *testing.T) {
	ts := newTestServer(t, http.StatusInternalServerError, "Token Expired")
	defer ts.Close()

	w := &Warp10{
		WarpURL:          ts.URL,
		Token:            config.NewSecret([]byte("EXPIRED_TOKEN")),
		AuthErrorRetries: 2,
		PrintErrorBody:   true,
		Log:              testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())
	defer w.Close()

	require.ErrorContains(t, w.Write(testutil.MockMetrics()), "Token Expired")
	require.Equal(t, "EXPIRED_TOKEN", w.failedToken)
	require.Equal(t, uint(2), w.authRetriesLeft)
}
