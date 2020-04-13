package prometheus

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	inputs "github.com/influxdata/telegraf/plugins/inputs/prometheus"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestMetricVersion1(t *testing.T) {
	Logger := testutil.Logger{Name: "outputs.prometheus_client"}
	tests := []struct {
		name     string
		output   *PrometheusClient
		metrics  []telegraf.Metric
		expected []byte
	}{
		{
			name: "simple",
			output: &PrometheusClient{
				Listen:            ":0",
				MetricVersion:     1,
				CollectorsExclude: []string{"gocollector", "process"},
				Path:              "/metrics",
				Log:               Logger,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "prometheus untyped",
			output: &PrometheusClient{
				Listen:            ":0",
				MetricVersion:     1,
				CollectorsExclude: []string{"gocollector", "process"},
				Path:              "/metrics",
				Log:               Logger,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu_time_idle",
					map[string]string{
						"host": "example.org",
					},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "prometheus counter",
			output: &PrometheusClient{
				Listen:            ":0",
				MetricVersion:     1,
				CollectorsExclude: []string{"gocollector", "process"},
				Path:              "/metrics",
				Log:               Logger,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu_time_idle",
					map[string]string{
						"host": "example.org",
					},
					map[string]interface{}{
						"counter": 42.0,
					},
					time.Unix(0, 0),
					telegraf.Counter,
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle counter
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "replace characters when using string as label",
			output: &PrometheusClient{
				Listen:            ":0",
				MetricVersion:     1,
				CollectorsExclude: []string{"gocollector", "process"},
				Path:              "/metrics",
				StringAsLabel:     true,
				Log:               Logger,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu_time_idle",
					map[string]string{},
					map[string]interface{}{
						"host:name": "example.org",
						"counter":   42.0,
					},
					time.Unix(0, 0),
					telegraf.Counter,
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle counter
cpu_time_idle{host_name="example.org"} 42
`),
		},
		{
			name: "prometheus gauge",
			output: &PrometheusClient{
				Listen:            ":0",
				MetricVersion:     1,
				CollectorsExclude: []string{"gocollector", "process"},
				Path:              "/metrics",
				Log:               Logger,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu_time_idle",
					map[string]string{
						"host": "example.org",
					},
					map[string]interface{}{
						"gauge": 42.0,
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			},
			expected: []byte(`
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle gauge
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "prometheus histogram",
			output: &PrometheusClient{
				Listen:            ":0",
				MetricVersion:     1,
				CollectorsExclude: []string{"gocollector", "process"},
				Path:              "/metrics",
				Log:               Logger,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"http_request_duration_seconds",
					map[string]string{},
					map[string]interface{}{
						"sum":   53423,
						"0.05":  24054,
						"0.1":   33444,
						"0.2":   100392,
						"0.5":   129389,
						"1":     133988,
						"+Inf":  144320,
						"count": 144320,
					},
					time.Unix(0, 0),
					telegraf.Histogram,
				),
			},
			expected: []byte(`
# HELP http_request_duration_seconds Telegraf collected metric
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 24054
http_request_duration_seconds_bucket{le="0.1"} 33444
http_request_duration_seconds_bucket{le="0.2"} 100392
http_request_duration_seconds_bucket{le="0.5"} 129389
http_request_duration_seconds_bucket{le="1"} 133988
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_count 144320
`),
		},
		{
			name: "prometheus summary",
			output: &PrometheusClient{
				Listen:            ":0",
				MetricVersion:     1,
				CollectorsExclude: []string{"gocollector", "process"},
				Path:              "/metrics",
				Log:               Logger,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"rpc_duration_seconds",
					map[string]string{},
					map[string]interface{}{
						"0.01":  3102,
						"0.05":  3272,
						"0.5":   4773,
						"0.9":   9001,
						"0.99":  76656,
						"count": 2693,
						"sum":   17560473,
					},
					time.Unix(0, 0),
					telegraf.Summary,
				),
			},
			expected: []byte(`
# HELP rpc_duration_seconds Telegraf collected metric
# TYPE rpc_duration_seconds summary
rpc_duration_seconds{quantile="0.01"} 3102
rpc_duration_seconds{quantile="0.05"} 3272
rpc_duration_seconds{quantile="0.5"} 4773
rpc_duration_seconds{quantile="0.9"} 9001
rpc_duration_seconds{quantile="0.99"} 76656
rpc_duration_seconds_sum 1.7560473e+07
rpc_duration_seconds_count 2693
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.output.Init()
			require.NoError(t, err)

			err = tt.output.Connect()
			require.NoError(t, err)

			defer func() {
				err := tt.output.Close()
				require.NoError(t, err)
			}()

			err = tt.output.Write(tt.metrics)
			require.NoError(t, err)

			resp, err := http.Get(tt.output.URL())
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t,
				strings.TrimSpace(string(tt.expected)),
				strings.TrimSpace(string(body)))
		})
	}
}

func TestRoundTripMetricVersion1(t *testing.T) {
	Logger := testutil.Logger{Name: "outputs.prometheus_client"}
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "untyped",
			data: []byte(`
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle untyped
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "counter",
			data: []byte(`
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle counter
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "gauge",
			data: []byte(`
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle gauge
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "multi",
			data: []byte(`
# HELP cpu_time_guest Telegraf collected metric
# TYPE cpu_time_guest gauge
cpu_time_guest{host="one.example.org"} 42
cpu_time_guest{host="two.example.org"} 42
# HELP cpu_time_idle Telegraf collected metric
# TYPE cpu_time_idle gauge
cpu_time_idle{host="one.example.org"} 42
cpu_time_idle{host="two.example.org"} 42
`),
		},
		{
			name: "histogram",
			data: []byte(`
# HELP http_request_duration_seconds Telegraf collected metric
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 24054
http_request_duration_seconds_bucket{le="0.1"} 33444
http_request_duration_seconds_bucket{le="0.2"} 100392
http_request_duration_seconds_bucket{le="0.5"} 129389
http_request_duration_seconds_bucket{le="1"} 133988
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_count 144320
`),
		},
		{
			name: "summary",
			data: []byte(`
# HELP rpc_duration_seconds Telegraf collected metric
# TYPE rpc_duration_seconds summary
rpc_duration_seconds{quantile="0.01"} 3102
rpc_duration_seconds{quantile="0.05"} 3272
rpc_duration_seconds{quantile="0.5"} 4773
rpc_duration_seconds{quantile="0.9"} 9001
rpc_duration_seconds{quantile="0.99"} 76656
rpc_duration_seconds_sum 1.7560473e+07
rpc_duration_seconds_count 2693
`),
		},
	}

	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	url := fmt.Sprintf("http://%s", ts.Listener.Addr())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(tt.data)
			})

			input := &inputs.Prometheus{
				URLs:          []string{url},
				URLTag:        "",
				MetricVersion: 1,
			}
			var acc testutil.Accumulator
			err := input.Start(&acc)
			require.NoError(t, err)
			err = input.Gather(&acc)
			require.NoError(t, err)
			input.Stop()

			metrics := acc.GetTelegrafMetrics()

			output := &PrometheusClient{
				Listen:            "127.0.0.1:0",
				Path:              defaultPath,
				MetricVersion:     1,
				Log:               Logger,
				CollectorsExclude: []string{"gocollector", "process"},
			}
			err = output.Init()
			require.NoError(t, err)
			err = output.Connect()
			require.NoError(t, err)
			defer func() {
				err = output.Close()
				require.NoError(t, err)
			}()
			err = output.Write(metrics)
			require.NoError(t, err)

			resp, err := http.Get(output.URL())
			require.NoError(t, err)

			actual, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t,
				strings.TrimSpace(string(tt.data)),
				strings.TrimSpace(string(actual)))
		})
	}
}
