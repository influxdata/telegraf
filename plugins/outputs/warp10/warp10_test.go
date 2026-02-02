package warp10

import (
	"math"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
)

type ErrorTest struct {
	Message           string
	Expected          string
	ExpectedRetryable bool
}

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
	tests := [...]*ErrorTest{
		{
			Message: `
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
			Expected:          "invalid token",
			ExpectedRetryable: false, // Authentication error
		},
		{
			Message: `
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
			Expected:          "token expired",
			ExpectedRetryable: false, // Authentication error
		},
		{
			Message: `
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
			Expected:          "token revoked",
			ExpectedRetryable: false, // Authentication error
		},
		{
			Message: `
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
			Expected:          "write token missing",
			ExpectedRetryable: false, // Authentication error
		},
		{
			Message:           `<title>Error 503: server unavailable</title>`,
			Expected:          "<title>Error 503: server unavailable</title>",
			ExpectedRetryable: true, // Temporary server error, retryable
		},
	}

	for _, handledError := range tests {
		werr := HandleError(handledError.Message, 511)
		require.IsType(t, &internal.HTTPError{}, werr)
		require.Equal(t, handledError.Expected, werr.Error())
		require.Equal(t, handledError.ExpectedRetryable, werr.Retryable, "retryable mismatch for: %s", handledError.Expected)
	}
}

func TestTokenChangeDetection_SameTokenBlocksWrites(t *testing.T) {
	// Server always returns auth error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Invalid token"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	w := &Warp10{
		Prefix:             "test.",
		WarpURL:            server.URL,
		Token:              config.NewSecret([]byte("static-token")),
		MaxStringErrorSize: 511,
		Log:                testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())

	metrics := testutil.MockMetrics()

	// First write - should return error (auth failure), metrics stay in buffer
	err := w.Write(metrics)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "max retries exceeded")
	require.Equal(t, "static-token", w.lastFailedToken)
	require.Equal(t, 1, w.authFailureCount)

	// Second write with same token - should return error, metrics stay in buffer
	err = w.Write(metrics)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pending token refresh")
	require.Equal(t, 2, w.authFailureCount)
}

func TestTokenChangeDetection_TokenChangeResumesWrites(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		token := r.Header.Get("X-Warp10-Token")

		// First request with old token fails, second with new token succeeds
		if count == 1 && token == "old-token" {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("Invalid token"))
			assert.NoError(t, err)
			return
		}
		if token == "new-token" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Invalid token"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	// Start with old token
	w := &Warp10{
		Prefix:             "test.",
		WarpURL:            server.URL,
		Token:              config.NewSecret([]byte("old-token")),
		MaxStringErrorSize: 511,
		Log:                testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())

	metrics := testutil.MockMetrics()

	// First write fails with auth error
	err := w.Write(metrics)
	require.Error(t, err)
	require.Equal(t, "old-token", w.lastFailedToken)
	require.Equal(t, 1, w.authFailureCount)

	// Simulate token refresh by secret-store
	w.Token = config.NewSecret([]byte("new-token"))

	// Second write with new token should succeed
	err = w.Write(metrics)
	require.NoError(t, err)
	require.Empty(t, w.lastFailedToken)
	require.Equal(t, 0, w.authFailureCount)
}

func TestTokenChangeDetection_MaxRetriesDropsMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Invalid token"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	w := &Warp10{
		Prefix:             "test.",
		WarpURL:            server.URL,
		Token:              config.NewSecret([]byte("static-token")),
		MaxStringErrorSize: 511,
		Log:                testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())

	metrics := testutil.MockMetrics()

	// First write - auth failure, count = 1
	err := w.Write(metrics)
	require.Error(t, err)
	require.Equal(t, 1, w.authFailureCount)

	// Second write - same token, count = 2
	err = w.Write(metrics)
	require.Error(t, err)
	require.Equal(t, 2, w.authFailureCount)

	// Third write - max retries reached, should return PartialWriteError
	err = w.Write(metrics)
	require.Error(t, err)
	var partialErr *internal.PartialWriteError
	require.ErrorAs(t, err, &partialErr)
	require.Contains(t, partialErr.Error(), "max retries exceeded")
	// State should be cleared after dropping metrics
	require.Empty(t, w.lastFailedToken)
	require.Equal(t, 0, w.authFailureCount)
}

func TestTokenChangeDetection_SuccessClearsFailureState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	w := &Warp10{
		Prefix:             "test.",
		WarpURL:            server.URL,
		Token:              config.NewSecret([]byte("valid-token")),
		MaxStringErrorSize: 511,
		Log:                testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())

	// Simulate previous auth failure state
	w.lastFailedToken = "old-failed-token"
	w.authFailureCount = 2

	metrics := testutil.MockMetrics()

	// Write succeeds (token changed)
	err := w.Write(metrics)
	require.NoError(t, err)
	require.Empty(t, w.lastFailedToken)
	require.Equal(t, 0, w.authFailureCount)
}

func TestTokenChangeDetection_RetryableErrorClearsAuthState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("exceed your Monthly Active Data Streams limit"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	w := &Warp10{
		Prefix:             "test.",
		WarpURL:            server.URL,
		Token:              config.NewSecret([]byte("valid-token")),
		MaxStringErrorSize: 511,
		Log:                testutil.Logger{},
	}
	require.NoError(t, w.Init())
	require.NoError(t, w.Connect())

	// Simulate previous auth failure state
	w.lastFailedToken = "old-failed-token"
	w.authFailureCount = 2

	metrics := testutil.MockMetrics()

	// Write returns retryable error - should clear auth failure state
	err := w.Write(metrics)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Monthly Active Data Streams limit")
	// Auth failure state should be cleared since this is a different error type
	require.Empty(t, w.lastFailedToken)
	require.Equal(t, 0, w.authFailureCount)
}
