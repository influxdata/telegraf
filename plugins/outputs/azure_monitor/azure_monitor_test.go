package azure_monitor

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestAggregate(t *testing.T) {
	tests := []struct {
		name                  string
		stringdim             bool
		metrics               []telegraf.Metric
		addTime               time.Time
		pushTime              time.Time
		expected              []telegraf.Metric
		expectedOutsideWindow int64
	}{
		{
			name: "add metric outside window is dropped",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			addTime:               time.Unix(3600, 0),
			pushTime:              time.Unix(3600, 0),
			expectedOutsideWindow: 1,
		},
		{
			name: "metric not sent until period expires",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(0, 0),
		},
		{
			name:      "add strings as dimensions",
			stringdim: true,
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "localhost",
					},
					map[string]interface{}{
						"value":   42,
						"message": "howdy",
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu-value",
					map[string]string{
						"host":    "localhost",
						"message": "howdy",
					},
					map[string]interface{}{
						"min":   42.0,
						"max":   42.0,
						"sum":   42.0,
						"count": 1,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "add metric to cache and push",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu-value",
					map[string]string{},
					map[string]interface{}{
						"min":   42.0,
						"max":   42.0,
						"sum":   42.0,
						"count": 1,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "added metric are aggregated",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 84,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 2,
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu-value",
					map[string]string{},
					map[string]interface{}{
						"min":   2.0,
						"max":   84.0,
						"sum":   128.0,
						"count": 3,
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msiEndpoint, err := adal.GetMSIVMEndpoint()
			require.NoError(t, err)
			t.Setenv("MSI_ENDPOINT", msiEndpoint)

			// Setup plugin
			plugin := &AzureMonitor{
				Region:               "test",
				ResourceID:           "/test",
				StringsAsDimensions:  tt.stringdim,
				TimestampLimitPast:   config.Duration(30 * time.Minute),
				TimestampLimitFuture: config.Duration(-1 * time.Minute),
				Log:                  testutil.Logger{},
				timeFunc:             func() time.Time { return tt.addTime },
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Reset statistics
			plugin.MetricOutsideWindow.Set(0)

			// Add the data
			for _, m := range tt.metrics {
				plugin.Add(m)
			}

			// Push out the data at a later time
			plugin.timeFunc = func() time.Time { return tt.pushTime }
			metrics := plugin.Push()
			plugin.Reset()

			// Check the results
			require.Equal(t, tt.expectedOutsideWindow, plugin.MetricOutsideWindow.Get())
			testutil.RequireMetricsEqual(t, tt.expected, metrics)
		})
	}
}

func TestWrite(t *testing.T) {
	// Set up a fake environment for Authorizer
	// This used to fake an MSI environment, but since https://github.com/Azure/go-autorest/pull/670/files it's no longer possible,
	// So we fake a user/password authentication
	t.Setenv("AZURE_CLIENT_ID", "fake")
	t.Setenv("AZURE_USERNAME", "fake")
	t.Setenv("AZURE_PASSWORD", "fake")

	tests := []struct {
		name            string
		metrics         []telegraf.Metric
		expectedCalls   uint64
		expectedMetrics uint64
		errmsg          string
	}{
		{
			name: "if not an azure metric nothing is sent",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			errmsg: "translating metric(s) failed",
		},
		{
			name: "single azure metric",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu-value",
					map[string]string{},
					map[string]interface{}{
						"min":   float64(42),
						"max":   float64(42),
						"sum":   float64(42),
						"count": int64(1),
					},
					time.Unix(0, 0),
				),
			},
			expectedCalls:   1,
			expectedMetrics: 1,
		},
		{
			name: "multiple azure metric",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu-value",
					map[string]string{},
					map[string]interface{}{
						"min":   float64(42),
						"max":   float64(42),
						"sum":   float64(42),
						"count": int64(1),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu-value",
					map[string]string{},
					map[string]interface{}{
						"min":   float64(42),
						"max":   float64(42),
						"sum":   float64(42),
						"count": int64(1),
					},
					time.Unix(60, 0),
				),
			},
			expectedCalls:   1,
			expectedMetrics: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test server to collect the sent metrics
			var calls atomic.Uint64
			var metrics atomic.Uint64
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)

				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Logf("cannot create gzip reader: %v", err)
					t.Fail()
					return
				}

				scanner := bufio.NewScanner(gz)
				for scanner.Scan() {
					var m azureMonitorMetric
					if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						t.Logf("cannot unmarshal JSON: %v", err)
						t.Fail()
						return
					}
					metrics.Add(1)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			// Setup the plugin
			plugin := AzureMonitor{
				EndpointURL:          "http://" + ts.Listener.Addr().String(),
				Region:               "test",
				ResourceID:           "/test",
				TimestampLimitPast:   config.Duration(30 * time.Minute),
				TimestampLimitFuture: config.Duration(-1 * time.Minute),
				Log:                  testutil.Logger{},
				timeFunc:             func() time.Time { return time.Unix(120, 0) },
			}
			require.NoError(t, plugin.Init())

			// Override with testing setup
			plugin.preparer = autorest.CreatePreparer(autorest.NullAuthorizer{}.WithAuthorization())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			err := plugin.Write(tt.metrics)
			if tt.errmsg != "" {
				require.ErrorContains(t, err, tt.errmsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedCalls, calls.Load())
			require.Equal(t, tt.expectedMetrics, metrics.Load())
		})
	}
}

func TestWriteTimelimits(t *testing.T) {
	// Set up a fake environment for Authorizer
	// This used to fake an MSI environment, but since https://github.com/Azure/go-autorest/pull/670/files it's no longer possible,
	// So we fake a user/password authentication
	t.Setenv("AZURE_CLIENT_ID", "fake")
	t.Setenv("AZURE_USERNAME", "fake")
	t.Setenv("AZURE_PASSWORD", "fake")

	// Setup input metrics
	tref := time.Now().Truncate(time.Minute)
	inputs := []telegraf.Metric{
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "too old",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(-time.Hour),
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "30 min in the past",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(-30*time.Minute),
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "20 min in the past",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(-20*time.Minute),
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "10 min in the past",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(-10*time.Minute),
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "now",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref,
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "1 min in the future",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(1*time.Minute),
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "2 min in the future",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(2*time.Minute),
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "4 min in the future",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(4*time.Minute),
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "5 min in the future",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(5*time.Minute),
		),
		metric.New(
			"cpu-value",
			map[string]string{
				"status": "too far in the future",
			},
			map[string]interface{}{
				"min":   float64(42),
				"max":   float64(42),
				"sum":   float64(42),
				"count": int64(1),
			},
			tref.Add(time.Hour),
		),
	}

	// Error message for status 400
	msg := `{"error":{"code":"BadRequest","message":"'time' should not be older than 30 minutes and not more than 4 minutes in the future\r\n"}}`

	tests := []struct {
		name          string
		input         []telegraf.Metric
		limitPast     time.Duration
		limitFuture   time.Duration
		expectedCount int
		expectedError string
	}{
		{
			name:          "only good metrics",
			input:         inputs[1 : len(inputs)-2],
			limitPast:     48 * time.Hour,
			limitFuture:   48 * time.Hour,
			expectedCount: len(inputs) - 3,
		},
		{
			name:          "metrics out of bounds",
			input:         inputs,
			limitPast:     48 * time.Hour,
			limitFuture:   48 * time.Hour,
			expectedCount: len(inputs),
			expectedError: "400 Bad Request: " + msg,
		},
		{
			name:          "default limit",
			input:         inputs,
			limitPast:     20 * time.Minute,
			limitFuture:   -1 * time.Minute,
			expectedCount: 2,
			expectedError: "metric(s) outside of acceptable time window",
		},
		{
			name:          "permissive limit",
			input:         inputs,
			limitPast:     30 * time.Minute,
			limitFuture:   5 * time.Minute,
			expectedCount: len(inputs) - 2,
			expectedError: "metric(s) outside of acceptable time window",
		},
		{
			name:          "very strict",
			input:         inputs,
			limitPast:     19*time.Minute + 59*time.Second,
			limitFuture:   3*time.Minute + 59*time.Second,
			expectedCount: len(inputs) - 6,
			expectedError: "metric(s) outside of acceptable time window",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Counter for the number of received metrics
			var count atomic.Int32

			// Setup test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()

				reader, err := gzip.NewReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Logf("unzipping content failed: %v", err)
					t.Fail()
					return
				}
				defer reader.Close()

				status := http.StatusOK
				scanner := bufio.NewScanner(reader)
				for scanner.Scan() {
					var data map[string]interface{}
					if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						t.Logf("decoding JSON failed: %v", err)
						t.Fail()
						return
					}

					timestamp, err := time.Parse(time.RFC3339, data["time"].(string))
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						t.Logf("decoding time failed: %v", err)
						t.Fail()
						return
					}
					if timestamp.Before(tref.Add(-30*time.Minute)) || timestamp.After(tref.Add(5*time.Minute)) {
						status = http.StatusBadRequest
					}
					count.Add(1)
				}
				w.WriteHeader(status)
				if status == 400 {
					//nolint:errcheck // Ignoring returned error as it is not relevant for the test
					w.Write([]byte(msg))
				}
			}))
			defer ts.Close()

			// Setup plugin
			plugin := AzureMonitor{
				EndpointURL:          "http://" + ts.Listener.Addr().String(),
				Region:               "test",
				ResourceID:           "/test",
				TimestampLimitPast:   config.Duration(tt.limitPast),
				TimestampLimitFuture: config.Duration(tt.limitFuture),
				Log:                  testutil.Logger{},
				timeFunc:             func() time.Time { return tref },
			}
			require.NoError(t, plugin.Init())

			// Override with testing setup
			plugin.preparer = autorest.CreatePreparer(autorest.NullAuthorizer{}.WithAuthorization())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Test writing
			err := plugin.Write(tt.input)
			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.expectedError)
			}
			require.Equal(t, tt.expectedCount, int(count.Load()))
		})
	}
}
