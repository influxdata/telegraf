package prometheusremotewrite

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestRemoteWriteSerialize(t *testing.T) {
	tests := []struct {
		name     string
		config   FormatConfig
		metric   telegraf.Metric
		expected []byte
	}{
		{
			name: "simple",
			metric: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
				},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
			expected: []byte(`
cpu_time_idle{host="example.org"} 42
`),
		},
		{
			name: "prometheus input untyped",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{
					"code":   "400",
					"method": "post",
				},
				map[string]interface{}{
					"http_requests_total": 3.0,
				},
				time.Unix(0, 0),
				telegraf.Untyped,
			),
			expected: []byte(`
http_requests_total{code="400", method="post"} 3
`),
		},
		{
			name: "prometheus input counter",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{
					"code":   "400",
					"method": "post",
				},
				map[string]interface{}{
					"http_requests_total": 3.0,
				},
				time.Unix(0, 0),
				telegraf.Counter,
			),
			expected: []byte(`
http_requests_total{code="400", method="post"} 3
`),
		},
		{
			name: "prometheus input gauge",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{
					"code":   "400",
					"method": "post",
				},
				map[string]interface{}{
					"http_requests_total": 3.0,
				},
				time.Unix(0, 0),
				telegraf.Gauge,
			),
			expected: []byte(`
http_requests_total{code="400", method="post"} 3
`),
		},
		{
			name: "prometheus input histogram no buckets",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{},
				map[string]interface{}{
					"http_request_duration_seconds_sum":   53423,
					"http_request_duration_seconds_count": 144320,
				},
				time.Unix(0, 0),
				telegraf.Histogram,
			),
			expected: []byte(`
http_request_duration_seconds_count 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_bucket{le="+Inf"} 144320
`),
		},
		{
			name: "prometheus input histogram only bucket",
			metric: testutil.MustMetric(
				"prometheus",
				map[string]string{
					"le": "0.5",
				},
				map[string]interface{}{
					"http_request_duration_seconds_bucket": 129389.0,
				},
				time.Unix(0, 0),
				telegraf.Histogram,
			),
			expected: []byte(`
http_request_duration_seconds_count 0
http_request_duration_seconds_sum 0
http_request_duration_seconds_bucket{le="+Inf"} 0
http_request_duration_seconds_bucket{le="0.5"} 129389
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSerializer(FormatConfig{
				MetricSortOrder: SortMetrics,
				StringHandling:  tt.config.StringHandling,
			})
			require.NoError(t, err)
			data, err := s.Serialize(tt.metric)
			require.NoError(t, err)
			actual, err := prompbToText(data)
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(tt.expected)),
				strings.TrimSpace(string(actual)))
		})
	}
}

func TestRemoteWriteSerializeBatch(t *testing.T) {
	tests := []struct {
		name     string
		config   FormatConfig
		metrics  []telegraf.Metric
		expected []byte
	}{
		{
			name: "simple",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "one.example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "two.example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_idle{host="one.example.org"} 42
cpu_time_idle{host="two.example.org"} 42
`),
		},
		{
			name: "multiple metric families",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "one.example.org",
					},
					map[string]interface{}{
						"time_idle":  42.0,
						"time_guest": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_guest{host="one.example.org"} 42
cpu_time_idle{host="one.example.org"} 42
`),
		},
		{
			name: "histogram",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"prometheus",
					map[string]string{},
					map[string]interface{}{
						"http_request_duration_seconds_sum":   53423,
						"http_request_duration_seconds_count": 144320,
					},
					time.Unix(0, 0),
					telegraf.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "0.05"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 24054.0,
					},
					time.Unix(0, 0),
					telegraf.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "0.1"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 33444.0,
					},
					time.Unix(0, 0),
					telegraf.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "0.2"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 100392.0,
					},
					time.Unix(0, 0),
					telegraf.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "0.5"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 129389.0,
					},
					time.Unix(0, 0),
					telegraf.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "1.0"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 133988.0,
					},
					time.Unix(0, 0),
					telegraf.Histogram,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"le": "+Inf"},
					map[string]interface{}{
						"http_request_duration_seconds_bucket": 144320.0,
					},
					time.Unix(0, 0),
					telegraf.Histogram,
				),
			},
			expected: []byte(`
http_request_duration_seconds_count 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_bucket{le="0.05"} 24054
http_request_duration_seconds_bucket{le="0.1"} 33444
http_request_duration_seconds_bucket{le="0.2"} 100392
http_request_duration_seconds_bucket{le="0.5"} 129389
http_request_duration_seconds_bucket{le="1"} 133988
`),
		},
		{
			name: "summary with quantile",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"prometheus",
					map[string]string{},
					map[string]interface{}{
						"rpc_duration_seconds_sum":   1.7560473e+07,
						"rpc_duration_seconds_count": 2693,
					},
					time.Unix(0, 0),
					telegraf.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.01"},
					map[string]interface{}{
						"rpc_duration_seconds": 3102.0,
					},
					time.Unix(0, 0),
					telegraf.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.05"},
					map[string]interface{}{
						"rpc_duration_seconds": 3272.0,
					},
					time.Unix(0, 0),
					telegraf.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.5"},
					map[string]interface{}{
						"rpc_duration_seconds": 4773.0,
					},
					time.Unix(0, 0),
					telegraf.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.9"},
					map[string]interface{}{
						"rpc_duration_seconds": 9001.0,
					},
					time.Unix(0, 0),
					telegraf.Summary,
				),
				testutil.MustMetric(
					"prometheus",
					map[string]string{"quantile": "0.99"},
					map[string]interface{}{
						"rpc_duration_seconds": 76656.0,
					},
					time.Unix(0, 0),
					telegraf.Summary,
				),
			},
			expected: []byte(`
rpc_duration_seconds_count 2693
rpc_duration_seconds_sum 17560473
rpc_duration_seconds{quantile="0.01"} 3102
rpc_duration_seconds{quantile="0.05"} 3272
rpc_duration_seconds{quantile="0.5"} 4773
rpc_duration_seconds{quantile="0.9"} 9001
rpc_duration_seconds{quantile="0.99"} 76656
`),
		},
		{
			name: "newer sample",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 43.0,
					},
					time.Unix(1, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_idle 43
`),
		},
		{
			name: "colons are not replaced in metric name from measurement",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu::xyzzy",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu::xyzzy_time_idle 42
`),
		},
		{
			name: "colons are not replaced in metric name from field",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time:idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time:idle 42
`),
		},
		{
			name: "invalid label",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host-name": "example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_idle{host_name="example.org"} 42
`),
		},
		{
			name: "colons are replaced in label name",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host:name": "example.org",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_idle{host_name="example.org"} 42
`),
		},
		{
			name: "discard strings",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
						"cpu":       "cpu0",
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_idle 42
`),
		},
		{
			name: "string as label",
			config: FormatConfig{
				MetricSortOrder: SortMetrics,
				StringHandling:  StringAsLabel,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
						"cpu":       "cpu0",
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_idle{cpu="cpu0"} 42
`),
		},
		{
			name: "string as label duplicate tag",
			config: FormatConfig{
				MetricSortOrder: SortMetrics,
				StringHandling:  StringAsLabel,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 42.0,
						"cpu":       "cpu1",
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_idle{cpu="cpu0"} 42
`),
		},
		{
			name: "replace characters when using string as label",
			config: FormatConfig{
				MetricSortOrder: SortMetrics,
				StringHandling:  StringAsLabel,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"host:name": "example.org",
						"time_idle": 42.0,
					},
					time.Unix(1574279268, 0),
				),
			},
			expected: []byte(`
cpu_time_idle{host_name="example.org"} 42
`),
		},
		{
			name: "multiple fields grouping",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_guest":  8106.04,
						"time_system": 26271.4,
						"time_user":   92904.33,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu1",
					},
					map[string]interface{}{
						"time_guest":  8181.63,
						"time_system": 25351.49,
						"time_user":   96912.57,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu2",
					},
					map[string]interface{}{
						"time_guest":  7470.04,
						"time_system": 24998.43,
						"time_user":   96034.08,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu3",
					},
					map[string]interface{}{
						"time_guest":  7517.95,
						"time_system": 24970.82,
						"time_user":   94148,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
cpu_time_guest{cpu="cpu0"} 8106.04
cpu_time_system{cpu="cpu0"} 26271.4
cpu_time_user{cpu="cpu0"} 92904.33
cpu_time_guest{cpu="cpu1"} 8181.63
cpu_time_system{cpu="cpu1"} 25351.49
cpu_time_user{cpu="cpu1"} 96912.57
cpu_time_guest{cpu="cpu2"} 7470.04
cpu_time_system{cpu="cpu2"} 24998.43
cpu_time_user{cpu="cpu2"} 96034.08
cpu_time_guest{cpu="cpu3"} 7517.95
cpu_time_system{cpu="cpu3"} 24970.82
cpu_time_user{cpu="cpu3"} 94148
`),
		},
		{
			name: "summary with no quantile",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"prometheus",
					map[string]string{},
					map[string]interface{}{
						"rpc_duration_seconds_sum":   1.7560473e+07,
						"rpc_duration_seconds_count": 2693,
					},
					time.Unix(0, 0),
					telegraf.Summary,
				),
			},
			expected: []byte(`
rpc_duration_seconds_count 2693
rpc_duration_seconds_sum 17560473
`),
		},
		{
			name: "empty label string value",
			config: FormatConfig{
				MetricSortOrder: SortMetrics,
				StringHandling:  StringAsLabel,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"prometheus",
					map[string]string{
						"cpu": "",
					},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []byte(`
			time_idle 42
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSerializer(FormatConfig{
				MetricSortOrder: SortMetrics,
				StringHandling:  tt.config.StringHandling,
			})
			require.NoError(t, err)
			data, err := s.SerializeBatch(tt.metrics)
			require.NoError(t, err)
			actual, err := prompbToText(data)
			require.NoError(t, err)

			require.Equal(t,
				strings.TrimSpace(string(tt.expected)),
				strings.TrimSpace(string(actual)))
		})
	}
}

func prompbToText(data []byte) ([]byte, error) {
	var buf = bytes.Buffer{}
	protobuff, err := snappy.Decode(nil, data)
	if err != nil {
		return nil, err
	}
	var req prompb.WriteRequest
	err = req.Unmarshal(protobuff)
	if err != nil {
		return nil, err
	}
	samples := protoToSamples(&req)
	for _, sample := range samples {
		_, err = buf.Write([]byte(fmt.Sprintf("%s %s\n", sample.Metric.String(), sample.Value.String())))
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func protoToSamples(req *prompb.WriteRequest) model.Samples {
	var samples model.Samples
	for _, ts := range req.Timeseries {
		metric := make(model.Metric, len(ts.Labels))
		for _, l := range ts.Labels {
			metric[model.LabelName(l.Name)] = model.LabelValue(l.Value)
		}

		for _, s := range ts.Samples {
			samples = append(samples, &model.Sample{
				Metric:    metric,
				Value:     model.SampleValue(s.Value),
				Timestamp: model.Time(s.Timestamp),
			})
		}
	}
	return samples
}
