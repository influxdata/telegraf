package arc

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tinylib/msgp/msgp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnect(t *testing.T) {
	a := &Arc{
		URL:              "http://localhost:8000/api/v1/write/msgpack",
		HTTPClientConfig: httpconfig.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
		Log:              testutil.Logger{},
	}

	err := a.Init()
	require.NoError(t, err)

	err = a.Connect()
	require.NoError(t, err)
	require.NotNil(t, a.client)
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name            string
		metrics         []telegraf.Metric
		expectedRecords int
		gzipEnabled     bool
	}{
		{
			name: "single metric",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
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
			gzipEnabled:     false,
		},
		{
			name: "multiple metrics with gzip",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{"host": "server01"},
					map[string]interface{}{"usage_idle": float64(95.5)},
					time.Unix(1633024800, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{"host": "server02"},
					map[string]interface{}{"usage_idle": float64(85.0)},
					time.Unix(1633024801, 0),
				),
			},
			expectedRecords: 2,
			gzipEnabled:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify HTTP method
				assert.Equal(t, "POST", r.Method)

				// Verify Content-Type
				assert.Equal(t, "application/msgpack", r.Header.Get("Content-Type"))

				// Verify User-Agent
				assert.Contains(t, r.Header.Get("User-Agent"), "Telegraf")

				// Read body
				var body io.Reader = r.Body
				if r.Header.Get("Content-Encoding") == "gzip" {
					gz, err := gzip.NewReader(r.Body)
					assert.NoError(t, err)
					defer gz.Close()
					body = gz
				}

				data, err := io.ReadAll(body)
				assert.NoError(t, err)

				// Decode MessagePack - handle both single and array format
				reader := msgp.NewReader(bytes.NewReader(data))
				decoded, err := reader.ReadIntf()
				assert.NoError(t, err)

				var columnarData arcColumnarData
				// Check if it's a map (single measurement) or array (multiple measurements)
				if decodedMap, ok := decoded.(map[string]interface{}); ok {
					// Single measurement
					columnarData.Measurement = decodedMap["m"].(string)
					columnarData.Columns = decodedMap["columns"].(map[string]interface{})
				} else if decodedArray, ok := decoded.([]interface{}); ok {
					// Multiple measurements - take the first one
					firstItem := decodedArray[0].(map[string]interface{})
					columnarData.Measurement = firstItem["m"].(string)
					columnarData.Columns = firstItem["columns"].(map[string]interface{})
				}

				// Verify columnar structure
				assert.NotEmpty(t, columnarData.Measurement)
				assert.NotEmpty(t, columnarData.Columns)

				// Verify time column exists
				timeCol, ok := columnarData.Columns["time"]
				assert.True(t, ok, "time column should exist")

				// Verify time column length matches expected records
				timeArray, ok := timeCol.([]interface{})
				assert.True(t, ok, "time should be an array")
				assert.Len(t, timeArray, tt.expectedRecords)

				// Return success
				w.WriteHeader(http.StatusNoContent)
			}))
			defer ts.Close()

			// Configure Arc plugin
			a := &Arc{
				URL:              ts.URL,
				HTTPClientConfig: httpconfig.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
				APIKey:           config.NewSecret([]byte("test-api-key")),
				Headers:          make(map[string]string),
				Log:              testutil.Logger{},
			}

			if tt.gzipEnabled {
				a.ContentEncoding = "gzip"
			} else {
				a.ContentEncoding = "identity"
			}

			err := a.Init()
			require.NoError(t, err)

			err = a.Connect()
			require.NoError(t, err)

			// Write metrics
			err = a.Write(tt.metrics)
			require.NoError(t, err)
		})
	}
}

func TestWriteWithAPIKey(t *testing.T) {
	expectedAPIKey := "test-secret-key-12345"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key header
		apiKey := r.Header.Get("x-api-key")
		assert.Equal(t, expectedAPIKey, apiKey)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	a := &Arc{
		URL:              ts.URL,
		HTTPClientConfig: httpconfig.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
		APIKey:           config.NewSecret([]byte(expectedAPIKey)),
		Log:              testutil.Logger{},
	}

	err := a.Init()
	require.NoError(t, err)

	err = a.Connect()
	require.NoError(t, err)

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{"host": "server01"},
			map[string]interface{}{"usage_idle": float64(95.5)},
			time.Now(),
		),
	}

	err = a.Write(metrics)
	require.NoError(t, err)
}

func TestWriteServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("Internal Server Error")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	a := &Arc{
		URL:              ts.URL,
		HTTPClientConfig: httpconfig.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
		Log:              testutil.Logger{},
	}

	err := a.Init()
	require.NoError(t, err)

	err = a.Connect()
	require.NoError(t, err)

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{"host": "server01"},
			map[string]interface{}{"usage_idle": float64(95.5)},
			time.Now(),
		),
	}

	err = a.Write(metrics)
	require.Error(t, err)
	require.Contains(t, err.Error(), "returned status 500")
}

func TestMessagePackEncoding(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.MustMetric(
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
		testutil.MustMetric(
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

	// Build columnar data
	columns := make(map[string]interface{})

	// Time column
	columns["time"] = []int64{
		metrics[0].Time().UnixMilli(),
		metrics[1].Time().UnixMilli(),
	}

	// Tag columns
	columns["host"] = []string{"server01", "server02"}
	columns["cpu"] = []string{"cpu0", "cpu1"}
	columns["environment"] = []string{"production", "production"}

	// Field columns
	columns["usage_idle"] = []interface{}{float64(95.5), float64(85.0)}
	columns["usage_system"] = []interface{}{float64(2.5), float64(4.5)}
	columns["usage_user"] = []interface{}{float64(2.0), float64(10.5)}

	// Create data as map[string]interface{} for msgp
	dataMap := map[string]interface{}{
		"m":       "cpu",
		"columns": columns,
	}

	// Marshal to MessagePack using tinylib/msgp
	var buf bytes.Buffer
	writer := msgp.NewWriter(&buf)
	err := writer.WriteIntf(dataMap)
	require.NoError(t, err)
	err = writer.Flush()
	require.NoError(t, err)

	data := buf.Bytes()
	require.NotEmpty(t, data)

	// Verify it can be unmarshaled
	reader := msgp.NewReader(bytes.NewReader(data))
	decodedIntf, err := reader.ReadIntf()
	require.NoError(t, err)

	decoded := decodedIntf.(map[string]interface{})
	require.Equal(t, "cpu", decoded["m"])
	require.NotEmpty(t, decoded["columns"])

	// Verify time column
	columnsMap := decoded["columns"].(map[string]interface{})
	timeCol, ok := columnsMap["time"]
	require.True(t, ok)
	timeArray, ok := timeCol.([]interface{})
	require.True(t, ok)
	require.Len(t, timeArray, 2)

	// Verify field columns
	usageIdleCol, ok := columnsMap["usage_idle"]
	require.True(t, ok)
	usageIdleArray, ok := usageIdleCol.([]interface{})
	require.True(t, ok)
	require.InDelta(t, 95.5, usageIdleArray[0], 0.01)
	require.InDelta(t, 85.0, usageIdleArray[1], 0.01)
}

func TestMultipleMeasurements(t *testing.T) {
	// Test with multiple different measurements
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read body
		var body io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			assert.NoError(t, err)
			defer gz.Close()
			body = gz
		}

		data, err := io.ReadAll(body)
		assert.NoError(t, err)

		// Try to decode as array of columnar data (multiple measurements)
		reader := msgp.NewReader(bytes.NewReader(data))
		decoded, err := reader.ReadIntf()
		assert.NoError(t, err)

		columnarDataArray := decoded.([]interface{})
		assert.Len(t, columnarDataArray, 2, "should have 2 measurements")

		// Verify we have both cpu and mem measurements
		measurementNames := make(map[string]bool)
		for _, item := range columnarDataArray {
			colData := item.(map[string]interface{})
			measurementNames[colData["m"].(string)] = true
			assert.NotEmpty(t, colData["columns"])
		}

		assert.True(t, measurementNames["cpu"], "should have cpu measurement")
		assert.True(t, measurementNames["mem"], "should have mem measurement")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	a := &Arc{
		URL:              ts.URL,
		HTTPClientConfig: httpconfig.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
		ContentEncoding:  "identity",
		Log:              testutil.Logger{},
	}

	err := a.Init()
	require.NoError(t, err)

	err = a.Connect()
	require.NoError(t, err)

	// Write metrics from different measurements
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{"host": "server01"},
			map[string]interface{}{"usage_idle": float64(95.5)},
			time.Unix(1633024800, 0),
		),
		testutil.MustMetric(
			"mem",
			map[string]string{"host": "server01"},
			map[string]interface{}{"usage_percent": float64(75.0)},
			time.Unix(1633024800, 0),
		),
	}

	err = a.Write(metrics)
	require.NoError(t, err)
}

func TestSampleConfig(t *testing.T) {
	a := &Arc{}
	cfg := a.SampleConfig()
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
				HTTPClientConfig: httpconfig.HTTPClientConfig{Timeout: config.Duration(5 * time.Second)},
			},
			expectError: false,
		},
		{
			name: "default values",
			arc: &Arc{
				URL: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.arc.Init()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClose(t *testing.T) {
	a := &Arc{
		Log: testutil.Logger{},
	}
	err := a.Init()
	require.NoError(t, err)

	err = a.Connect()
	require.NoError(t, err)

	err = a.Close()
	require.NoError(t, err)
}
