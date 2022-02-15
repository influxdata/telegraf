package prometheus

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type Input struct {
	metric  telegraf.Metric
	addtime time.Time
}

func TestCollectionExpire(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		age      time.Duration
		input    []Input
		expected []*dto.MetricFamily
	}{
		{
			name: "not expired",
			now:  time.Unix(1, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("cpu_time_idle"),
					Help: proto.String(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   []*dto.LabelPair{},
							Untyped: &dto.Untyped{Value: proto.Float64(42.0)},
						},
					},
				},
			},
		},
		{
			name: "update metric expiration",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				},
				{
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_idle": 43.0,
						},
						time.Unix(12, 0),
					),
					addtime: time.Unix(12, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("cpu_time_idle"),
					Help: proto.String(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   []*dto.LabelPair{},
							Untyped: &dto.Untyped{Value: proto.Float64(43.0)},
						},
					},
				},
			},
		},
		{
			name: "update metric expiration descending order",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_idle": 42.0,
						},
						time.Unix(12, 0),
					),
					addtime: time.Unix(12, 0),
				}, {
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_idle": 43.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("cpu_time_idle"),
					Help: proto.String(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   []*dto.LabelPair{},
							Untyped: &dto.Untyped{Value: proto.Float64(42.0)},
						},
					},
				},
			},
		},
		{
			name: "expired single metric in metric family",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{},
		},
		{
			name: "expired one metric in metric family",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_guest": 42.0,
						},
						time.Unix(15, 0),
					),
					addtime: time.Unix(15, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("cpu_time_guest"),
					Help: proto.String(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   []*dto.LabelPair{},
							Untyped: &dto.Untyped{Value: proto.Float64(42.0)},
						},
					},
				},
			},
		},
		{
			name: "histogram bucket updates",
			now:  time.Unix(0, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"http_request_duration_seconds_sum":   10.0,
							"http_request_duration_seconds_count": 2,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					// Next interval
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"http_request_duration_seconds_sum":   20.0,
							"http_request_duration_seconds_count": 4,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 2.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 2.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("http_request_duration_seconds"),
					Help: proto.String(helpString),
					Type: dto.MetricType_HISTOGRAM.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{},
							Histogram: &dto.Histogram{
								SampleCount: proto.Uint64(4),
								SampleSum:   proto.Float64(20.0),
								Bucket: []*dto.Bucket{
									{
										UpperBound:      proto.Float64(0.05),
										CumulativeCount: proto.Uint64(2),
									},
									{
										UpperBound:      proto.Float64(math.Inf(1)),
										CumulativeCount: proto.Uint64(2),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "entire histogram expires",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"http_request_duration_seconds_sum":   10.0,
							"http_request_duration_seconds_count": 2,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{},
		},
		{
			name: "histogram does not expire because of addtime from bucket",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"http_request_duration_seconds_sum":   10.0,
							"http_request_duration_seconds_count": 2,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(15, 0), // More recent addtime causes entire metric to stay valid
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("http_request_duration_seconds"),
					Help: proto.String(helpString),
					Type: dto.MetricType_HISTOGRAM.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{},
							Histogram: &dto.Histogram{
								SampleCount: proto.Uint64(2),
								SampleSum:   proto.Float64(10.0),
								Bucket: []*dto.Bucket{
									{
										UpperBound:      proto.Float64(math.Inf(1)),
										CumulativeCount: proto.Uint64(1),
									},
									{
										UpperBound:      proto.Float64(0.05),
										CumulativeCount: proto.Uint64(1),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "summary quantile updates",
			now:  time.Unix(0, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"rpc_duration_seconds_sum":   1.0,
							"rpc_duration_seconds_count": 1,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]interface{}{
							"rpc_duration_seconds": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					// Updated Summary
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"rpc_duration_seconds_sum":   2.0,
							"rpc_duration_seconds_count": 2,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]interface{}{
							"rpc_duration_seconds": 2.0,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("rpc_duration_seconds"),
					Help: proto.String(helpString),
					Type: dto.MetricType_SUMMARY.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{},
							Summary: &dto.Summary{
								SampleCount: proto.Uint64(2),
								SampleSum:   proto.Float64(2.0),
								Quantile: []*dto.Quantile{
									{
										Quantile: proto.Float64(0.01),
										Value:    proto.Float64(2),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Entire summary expires",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"rpc_duration_seconds_sum":   1.0,
							"rpc_duration_seconds_count": 1,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]interface{}{
							"rpc_duration_seconds": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{},
		},
		{
			name: "summary does not expire because of quantile addtime",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"rpc_duration_seconds_sum":   1.0,
							"rpc_duration_seconds_count": 1,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"quantile": "0.5"},
						map[string]interface{}{
							"rpc_duration_seconds": 10.0,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]interface{}{
							"rpc_duration_seconds": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(15, 0), // Recent addtime keeps entire metric around
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("rpc_duration_seconds"),
					Help: proto.String(helpString),
					Type: dto.MetricType_SUMMARY.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{},
							Summary: &dto.Summary{
								SampleSum:   proto.Float64(1),
								SampleCount: proto.Uint64(1),
								Quantile: []*dto.Quantile{
									{
										Quantile: proto.Float64(0.5),
										Value:    proto.Float64(10),
									},
									{
										Quantile: proto.Float64(0.01),
										Value:    proto.Float64(1),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "expire based on add time",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"cpu",
						map[string]string{},
						map[string]interface{}{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(15, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("cpu_time_idle"),
					Help: proto.String(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   []*dto.LabelPair{},
							Untyped: &dto.Untyped{Value: proto.Float64(42.0)},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollection(FormatConfig{})
			for _, item := range tt.input {
				c.Add(item.metric, item.addtime)
			}
			c.Expire(tt.now, tt.age)

			actual := c.GetProto()

			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestExportTimestamps(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		age      time.Duration
		input    []Input
		expected []*dto.MetricFamily
	}{
		{
			name: "histogram bucket updates",
			now:  time.Unix(23, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"http_request_duration_seconds_sum":   10.0,
							"http_request_duration_seconds_count": 2,
						},
						time.Unix(15, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(15, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(15, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					// Next interval
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"http_request_duration_seconds_sum":   20.0,
							"http_request_duration_seconds_count": 4,
						},
						time.Unix(20, 0), // Updated timestamp
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 2.0,
						},
						time.Unix(20, 0), // Updated timestamp
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]interface{}{
							"http_request_duration_seconds_bucket": 2.0,
						},
						time.Unix(20, 0), // Updated timestamp
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("http_request_duration_seconds"),
					Help: proto.String(helpString),
					Type: dto.MetricType_HISTOGRAM.Enum(),
					Metric: []*dto.Metric{
						{
							Label:       []*dto.LabelPair{},
							TimestampMs: proto.Int64(time.Unix(20, 0).UnixNano() / int64(time.Millisecond)),
							Histogram: &dto.Histogram{
								SampleCount: proto.Uint64(4),
								SampleSum:   proto.Float64(20.0),
								Bucket: []*dto.Bucket{
									{
										UpperBound:      proto.Float64(0.05),
										CumulativeCount: proto.Uint64(2),
									},
									{
										UpperBound:      proto.Float64(math.Inf(1)),
										CumulativeCount: proto.Uint64(2),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "summary quantile updates",
			now:  time.Unix(23, 0),
			age:  10 * time.Second,
			input: []Input{
				{
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"rpc_duration_seconds_sum":   1.0,
							"rpc_duration_seconds_count": 1,
						},
						time.Unix(15, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]interface{}{
							"rpc_duration_seconds": 1.0,
						},
						time.Unix(15, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(23, 0),
				}, {
					// Updated Summary
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{},
						map[string]interface{}{
							"rpc_duration_seconds_sum":   2.0,
							"rpc_duration_seconds_count": 2,
						},
						time.Unix(20, 0), // Updated timestamp
						telegraf.Summary,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: testutil.MustMetric(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]interface{}{
							"rpc_duration_seconds": 2.0,
						},
						time.Unix(20, 0), // Updated timestamp
						telegraf.Summary,
					),
					addtime: time.Unix(23, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: proto.String("rpc_duration_seconds"),
					Help: proto.String(helpString),
					Type: dto.MetricType_SUMMARY.Enum(),
					Metric: []*dto.Metric{
						{
							Label:       []*dto.LabelPair{},
							TimestampMs: proto.Int64(time.Unix(20, 0).UnixNano() / int64(time.Millisecond)),
							Summary: &dto.Summary{
								SampleCount: proto.Uint64(2),
								SampleSum:   proto.Float64(2.0),
								Quantile: []*dto.Quantile{
									{
										Quantile: proto.Float64(0.01),
										Value:    proto.Float64(2),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollection(FormatConfig{TimestampExport: ExportTimestamp})
			for _, item := range tt.input {
				c.Add(item.metric, item.addtime)
			}
			c.Expire(tt.now, tt.age)

			actual := c.GetProto()

			require.Equal(t, tt.expected, actual)
		})
	}
}
