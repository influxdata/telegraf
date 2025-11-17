package arc

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinylib/msgp/msgp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/testutil"
)

func TestSampleConfig(t *testing.T) {
	plugin := &Arc{}
	cfg := plugin.SampleConfig()
	require.NotEmpty(t, cfg)
	require.Contains(t, cfg, "url")
	require.Contains(t, cfg, "api_key")
}

func TestInit(t *testing.T) {
	tests := []struct {
		name        string
		arc         *Arc
		expectError bool
	}{
		{
			name: "valid config",
			arc: &Arc{
				URL:              "http://localhost:8000/api/v1/write/msgpack",
				HTTPClientConfig: common_http.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
			},
			expectError: false,
		},
		{
			name:        "missing url",
			arc:         &Arc{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.arc.Init()
			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "url is required")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name            string
		metrics         []telegraf.Metric
		expectedRecords int
		contentEncoding string
	}{
		{
			name: "single metric",
			metrics: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{
						"host": "server01",
						"cpu":  "cpu0",
					},
					map[string]interface{}{
						"usage_idle":   float64(95.5),
						"usage_system": float64(2.5),
						"usage_user":   float64(2.0),
					},
					time.Unix(1633024800, 0),
				),
			},
			expectedRecords: 1,
			contentEncoding: "identity",
		},
		{
			name: "multiple metrics with gzip",
			metrics: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{"host": "server01"},
					map[string]interface{}{"usage_idle": float64(95.5)},
					time.Unix(1633024800, 0),
				),
				metric.New(
					"cpu",
					map[string]string{"host": "server02"},
					map[string]interface{}{"usage_idle": float64(85.0)},
					time.Unix(1633024801, 0),
				),
			},
			expectedRecords: 2,
			contentEncoding: "gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedMethod string
			var receivedContentType string
			var receivedUserAgent string
			var receivedContentEncoding string
			var receivedBody []byte
			var receivedMu sync.Mutex
			var done atomic.Bool

			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMu.Lock()
				defer receivedMu.Unlock()

				receivedMethod = r.Method
				receivedContentType = r.Header.Get("Content-Type")
				receivedUserAgent = r.Header.Get("User-Agent")
				receivedContentEncoding = r.Header.Get("Content-Encoding")

				// Read body
				var body io.Reader = r.Body
				if receivedContentEncoding == "gzip" {
					gz, err := gzip.NewReader(r.Body)
					if err != nil {
						t.Fail()
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					defer gz.Close()
					body = gz
				}

				data, err := io.ReadAll(body)
				if err != nil {
					t.Fail()
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				receivedBody = data
				done.Store(true)

				w.WriteHeader(http.StatusNoContent)
			}))
			defer ts.Close()

			// Configure Arc plugin
			plugin := &Arc{
				URL:              ts.URL,
				HTTPClientConfig: common_http.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
				APIKey:           config.NewSecret([]byte("test-api-key")),
				Headers:          make(map[string]string),
				ContentEncoding:  tt.contentEncoding,
				Log:              testutil.Logger{},
			}

			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())

			// Write metrics
			require.NoError(t, plugin.Write(tt.metrics))

			// Wait for the data to arrive
			require.Eventually(t, func() bool {
				return done.Load()
			}, 1*time.Second, 100*time.Millisecond)

			// Verify HTTP request
			receivedMu.Lock()
			defer receivedMu.Unlock()
			require.Equal(t, "POST", receivedMethod)
			require.Equal(t, "application/msgpack", receivedContentType)
			require.Contains(t, receivedUserAgent, "Telegraf")
			if tt.contentEncoding == "gzip" {
				require.Equal(t, "gzip", receivedContentEncoding)
			}

			// Decode MessagePack payload and verify it's valid
			reader := msgp.NewReader(bytes.NewReader(receivedBody))
			decoded, err := reader.ReadIntf()
			require.NoError(t, err)
			require.NotNil(t, decoded)

			// Verify the structure is a map with measurement and columns
			data, ok := decoded.(map[string]interface{})
			require.True(t, ok, "decoded data should be a map")
			require.Contains(t, data, "m", "should have measurement name")
			require.Contains(t, data, "columns", "should have columns")
		})
	}
}

func TestWriteWithAPIKey(t *testing.T) {
	expectedAPIKey := "test-secret-key-12345"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey := r.Header.Get("x-api-key"); apiKey != expectedAPIKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	plugin := &Arc{
		URL:              ts.URL,
		HTTPClientConfig: common_http.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
		APIKey:           config.NewSecret([]byte(expectedAPIKey)),
		Log:              testutil.Logger{},
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	metrics := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{"host": "server01"},
			map[string]interface{}{"usage_idle": float64(95.5)},
			time.Now(),
		),
	}

	require.NoError(t, plugin.Write(metrics))
}

func TestWriteServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("Internal Server Error")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	plugin := &Arc{
		URL:              ts.URL,
		HTTPClientConfig: common_http.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
		Log:              testutil.Logger{},
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	metrics := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{"host": "server01"},
			map[string]interface{}{"usage_idle": float64(95.5)},
			time.Now(),
		),
	}

	require.ErrorContains(t, plugin.Write(metrics), "returned status 500")
}

func TestMessagePackEncoding(t *testing.T) {
	var receivedBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fail()
			return
		}
		receivedBody = data
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	plugin := &Arc{
		URL:              ts.URL,
		HTTPClientConfig: common_http.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
		ContentEncoding:  "identity",
		Log:              testutil.Logger{},
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	metrics := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{
				"host":        "server01",
				"cpu":         "cpu0",
				"environment": "production",
			},
			map[string]interface{}{
				"usage_idle":   float64(95.5),
				"usage_system": float64(2.5),
				"usage_user":   float64(2.0),
			},
			time.Unix(1633024800, 0),
		),
		metric.New(
			"cpu",
			map[string]string{
				"host":        "server02",
				"cpu":         "cpu1",
				"environment": "production",
			},
			map[string]interface{}{
				"usage_idle":   float64(85.0),
				"usage_system": float64(4.5),
				"usage_user":   float64(10.5),
			},
			time.Unix(1633024801, 0),
		),
	}

	require.NoError(t, plugin.Write(metrics))

	// Verify MessagePack encoding
	reader := msgp.NewReader(bytes.NewReader(receivedBody))
	decodedIntf, err := reader.ReadIntf()
	require.NoError(t, err)

	decoded := decodedIntf.(map[string]interface{})
	require.Equal(t, "cpu", decoded["m"])
	require.NotEmpty(t, decoded["columns"])

	// Verify columns structure
	columnsMap := decoded["columns"].(map[string]interface{})

	// Verify time column
	timeCol, ok := columnsMap["time"]
	require.True(t, ok)
	timeArray, ok := timeCol.([]interface{})
	require.True(t, ok)
	require.Len(t, timeArray, 2)

	// Verify tag columns
	hostCol := columnsMap["host"].([]interface{})
	require.Equal(t, "server01", hostCol[0])
	require.Equal(t, "server02", hostCol[1])

	// Verify field columns
	usageIdleCol := columnsMap["usage_idle"].([]interface{})
	require.InDelta(t, 95.5, usageIdleCol[0], 0.01)
	require.InDelta(t, 85.0, usageIdleCol[1], 0.01)
}

func TestMultipleMeasurements(t *testing.T) {
	var receivedBody []byte
	var receivedContentEncoding string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentEncoding = r.Header.Get("Content-Encoding")

		var body io.Reader = r.Body
		if receivedContentEncoding == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Fail()
				return
			}
			defer gz.Close()
			body = gz
		}

		data, err := io.ReadAll(body)
		if err != nil {
			t.Fail()
			return
		}
		receivedBody = data

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	plugin := &Arc{
		URL:              ts.URL,
		HTTPClientConfig: common_http.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
		ContentEncoding:  "identity",
		Log:              testutil.Logger{},
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Write metrics from different measurements
	metrics := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{"host": "server01"},
			map[string]interface{}{"usage_idle": float64(95.5)},
			time.Unix(1633024800, 0),
		),
		metric.New(
			"mem",
			map[string]string{"host": "server01"},
			map[string]interface{}{"usage_percent": float64(75.0)},
			time.Unix(1633024800, 0),
		),
	}

	require.NoError(t, plugin.Write(metrics))

	// Verify received data outside the handler
	reader := msgp.NewReader(bytes.NewReader(receivedBody))
	decoded, err := reader.ReadIntf()
	require.NoError(t, err)

	columnarDataArray := decoded.([]interface{})
	require.Len(t, columnarDataArray, 2, "should have 2 measurements")

	// Verify we have both cpu and mem measurements
	measurementNames := make(map[string]bool)
	for _, item := range columnarDataArray {
		colData := item.(map[string]interface{})
		measurementNames[colData["m"].(string)] = true
		require.NotEmpty(t, colData["columns"])
	}

	require.True(t, measurementNames["cpu"], "should have cpu measurement")
	require.True(t, measurementNames["mem"], "should have mem measurement")
}
