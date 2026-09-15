package prometheus

import (
	"math"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type input struct {
	metric  telegraf.Metric
	addtime time.Time
}

func TestCollectionExpire(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		age      time.Duration
		input    []input
		expected []*dto.MetricFamily
	}{
		{
			name: "not expired",
			now:  time.Unix(1, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: new("cpu_time_idle"),
					Help: new(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   make([]*dto.LabelPair, 0),
							Untyped: &dto.Untyped{Value: new(42.0)},
						},
					},
				},
			},
		},
		{
			name: "update metric expiration",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				},
				{
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_idle": 43.0,
						},
						time.Unix(12, 0),
					),
					addtime: time.Unix(12, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: new("cpu_time_idle"),
					Help: new(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   make([]*dto.LabelPair, 0),
							Untyped: &dto.Untyped{Value: new(43.0)},
						},
					},
				},
			},
		},
		{
			name: "update metric expiration descending order",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_idle": 42.0,
						},
						time.Unix(12, 0),
					),
					addtime: time.Unix(12, 0),
				}, {
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_idle": 43.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: new("cpu_time_idle"),
					Help: new(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   make([]*dto.LabelPair, 0),
							Untyped: &dto.Untyped{Value: new(42.0)},
						},
					},
				},
			},
		},
		{
			name: "expired single metric in metric family",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: make([]*dto.MetricFamily, 0),
		},
		{
			name: "expired one metric in metric family",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_guest": 42.0,
						},
						time.Unix(15, 0),
					),
					addtime: time.Unix(15, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: new("cpu_time_guest"),
					Help: new(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   make([]*dto.LabelPair, 0),
							Untyped: &dto.Untyped{Value: new(42.0)},
						},
					},
				},
			},
		},
		{
			name: "histogram bucket updates",
			now:  time.Unix(0, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"http_request_duration_seconds_sum":   10.0,
							"http_request_duration_seconds_count": 2,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]any{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]any{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					// Next interval
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"http_request_duration_seconds_sum":   20.0,
							"http_request_duration_seconds_count": 4,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]any{
							"http_request_duration_seconds_bucket": 2.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]any{
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
					Name: new("http_request_duration_seconds"),
					Help: new(helpString),
					Type: dto.MetricType_HISTOGRAM.Enum(),
					Metric: []*dto.Metric{
						{
							Label: make([]*dto.LabelPair, 0),
							Histogram: &dto.Histogram{
								SampleCount: proto.Uint64(4),
								SampleSum:   new(20.0),
								Bucket: []*dto.Bucket{
									{
										UpperBound:      new(0.05),
										CumulativeCount: proto.Uint64(2),
									},
									{
										UpperBound:      new(math.Inf(1)),
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
			input: []input{
				{
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"http_request_duration_seconds_sum":   10.0,
							"http_request_duration_seconds_count": 2,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]any{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]any{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: make([]*dto.MetricFamily, 0),
		},
		{
			name: "histogram does not expire because of addtime from bucket",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]any{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"http_request_duration_seconds_sum":   10.0,
							"http_request_duration_seconds_count": 2,
						},
						time.Unix(0, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]any{
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
					Name: new("http_request_duration_seconds"),
					Help: new(helpString),
					Type: dto.MetricType_HISTOGRAM.Enum(),
					Metric: []*dto.Metric{
						{
							Label: make([]*dto.LabelPair, 0),
							Histogram: &dto.Histogram{
								SampleCount: proto.Uint64(2),
								SampleSum:   new(10.0),
								Bucket: []*dto.Bucket{
									{
										UpperBound:      new(math.Inf(1)),
										CumulativeCount: proto.Uint64(1),
									},
									{
										UpperBound:      new(0.05),
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
			input: []input{
				{
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"rpc_duration_seconds_sum":   1.0,
							"rpc_duration_seconds_count": 1,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]any{
							"rpc_duration_seconds": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					// Updated Summary
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"rpc_duration_seconds_sum":   2.0,
							"rpc_duration_seconds_count": 2,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]any{
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
					Name: new("rpc_duration_seconds"),
					Help: new(helpString),
					Type: dto.MetricType_SUMMARY.Enum(),
					Metric: []*dto.Metric{
						{
							Label: make([]*dto.LabelPair, 0),
							Summary: &dto.Summary{
								SampleCount: proto.Uint64(2),
								SampleSum:   new(2.0),
								Quantile: []*dto.Quantile{
									{
										Quantile: new(0.01),
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
			input: []input{
				{
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"rpc_duration_seconds_sum":   1.0,
							"rpc_duration_seconds_count": 1,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]any{
							"rpc_duration_seconds": 1.0,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				},
			},
			expected: make([]*dto.MetricFamily, 0),
		},
		{
			name: "summary does not expire because of quantile addtime",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"rpc_duration_seconds_sum":   1.0,
							"rpc_duration_seconds_count": 1,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"quantile": "0.5"},
						map[string]any{
							"rpc_duration_seconds": 10.0,
						},
						time.Unix(0, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(0, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]any{
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
					Name: new("rpc_duration_seconds"),
					Help: new(helpString),
					Type: dto.MetricType_SUMMARY.Enum(),
					Metric: []*dto.Metric{
						{
							Label: make([]*dto.LabelPair, 0),
							Summary: &dto.Summary{
								SampleSum:   proto.Float64(1),
								SampleCount: proto.Uint64(1),
								Quantile: []*dto.Quantile{
									{
										Quantile: new(0.5),
										Value:    proto.Float64(10),
									},
									{
										Quantile: new(0.01),
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
			input: []input{
				{
					metric: metric.New(
						"cpu",
						map[string]string{},
						map[string]any{
							"time_idle": 42.0,
						},
						time.Unix(0, 0),
					),
					addtime: time.Unix(15, 0),
				},
			},
			expected: []*dto.MetricFamily{
				{
					Name: new("cpu_time_idle"),
					Help: new(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   make([]*dto.LabelPair, 0),
							Untyped: &dto.Untyped{Value: new(42.0)},
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
		input    []input
		expected []*dto.MetricFamily
	}{
		{
			name: "histogram bucket updates",
			now:  time.Unix(23, 0),
			age:  10 * time.Second,
			input: []input{
				{
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"http_request_duration_seconds_sum":   10.0,
							"http_request_duration_seconds_count": 2,
						},
						time.Unix(15, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]any{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(15, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]any{
							"http_request_duration_seconds_bucket": 1.0,
						},
						time.Unix(15, 0),
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					// Next interval
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"http_request_duration_seconds_sum":   20.0,
							"http_request_duration_seconds_count": 4,
						},
						time.Unix(20, 0), // Updated timestamp
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "0.05"},
						map[string]any{
							"http_request_duration_seconds_bucket": 2.0,
						},
						time.Unix(20, 0), // Updated timestamp
						telegraf.Histogram,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"le": "+Inf"},
						map[string]any{
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
					Name: new("http_request_duration_seconds"),
					Help: new(helpString),
					Type: dto.MetricType_HISTOGRAM.Enum(),
					Metric: []*dto.Metric{
						{
							Label:       make([]*dto.LabelPair, 0),
							TimestampMs: new(time.Unix(20, 0).UnixNano() / int64(time.Millisecond)),
							Histogram: &dto.Histogram{
								SampleCount: proto.Uint64(4),
								SampleSum:   new(20.0),
								Bucket: []*dto.Bucket{
									{
										UpperBound:      new(0.05),
										CumulativeCount: proto.Uint64(2),
									},
									{
										UpperBound:      new(math.Inf(1)),
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
			input: []input{
				{
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"rpc_duration_seconds_sum":   1.0,
							"rpc_duration_seconds_count": 1,
						},
						time.Unix(15, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]any{
							"rpc_duration_seconds": 1.0,
						},
						time.Unix(15, 0),
						telegraf.Summary,
					),
					addtime: time.Unix(23, 0),
				}, {
					// Updated Summary
					metric: metric.New(
						"prometheus",
						map[string]string{},
						map[string]any{
							"rpc_duration_seconds_sum":   2.0,
							"rpc_duration_seconds_count": 2,
						},
						time.Unix(20, 0), // Updated timestamp
						telegraf.Summary,
					),
					addtime: time.Unix(23, 0),
				}, {
					metric: metric.New(
						"prometheus",
						map[string]string{"quantile": "0.01"},
						map[string]any{
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
					Name: new("rpc_duration_seconds"),
					Help: new(helpString),
					Type: dto.MetricType_SUMMARY.Enum(),
					Metric: []*dto.Metric{
						{
							Label:       make([]*dto.LabelPair, 0),
							TimestampMs: new(time.Unix(20, 0).UnixNano() / int64(time.Millisecond)),
							Summary: &dto.Summary{
								SampleCount: proto.Uint64(2),
								SampleSum:   new(2.0),
								Quantile: []*dto.Quantile{
									{
										Quantile: new(0.01),
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
			c := NewCollection(FormatConfig{ExportTimestamp: true})
			for _, item := range tt.input {
				c.Add(item.metric, item.addtime)
			}
			c.Expire(tt.now, tt.age)

			actual := c.GetProto()

			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestCollectionLegacyDropsUTF8OnlyNames(t *testing.T) {
	c := NewCollection(FormatConfig{NameSanitization: "legacy"})
	c.Add(
		metric.New(
			"温度-指标",
			map[string]string{"主机-名": "example.org"},
			map[string]any{"数值-值": 42.0},
			time.Unix(0, 0),
		),
		time.Unix(0, 0),
	)

	// In legacy mode, purely UTF-8 metric names are sanitized to underscores
	// which get trimmed to empty strings, causing the metric to be dropped.
	// In contrast, TestCollectionUTF8NameSanitization shows these names
	// are preserved in utf8 mode.
	require.Empty(t, c.GetProto())
}

func TestCollectionUTF8NameSanitization(t *testing.T) {
	c := NewCollection(FormatConfig{NameSanitization: "utf8"})
	c.Add(
		metric.New(
			"温度-指标",
			map[string]string{"主机-名": "example.org"},
			map[string]any{"数值-值": 42.0},
			time.Unix(0, 0),
		),
		time.Unix(0, 0),
	)

	expected := []*dto.MetricFamily{
		{
			Name: new("温度-指标_数值-值"),
			Help: new(helpString),
			Type: dto.MetricType_UNTYPED.Enum(),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{
							Name:  new("主机-名"),
							Value: new("example.org"),
						},
					},
					Untyped: &dto.Untyped{Value: proto.Float64(42)},
				},
			},
		},
	}

	require.Equal(t, expected, c.GetProto())
}

func TestCollectionUTF8FallbackForInvalidUTF8(t *testing.T) {
	c := NewCollection(FormatConfig{NameSanitization: "utf8"})
	c.Add(
		metric.New(
			"cpu",
			map[string]string{
				string([]byte{0xff, 'h', '-', '1'}): "example.org",
			},
			map[string]any{
				string([]byte{0xff, 't', '-', 'x'}): 42.0,
			},
			time.Unix(0, 0),
		),
		time.Unix(0, 0),
	)

	expected := []*dto.MetricFamily{
		{
			Name: new("cpu__t_x"),
			Help: new(helpString),
			Type: dto.MetricType_UNTYPED.Enum(),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{
							Name:  new("h_1"),
							Value: new("example.org"),
						},
					},
					Untyped: &dto.Untyped{Value: proto.Float64(42)},
				},
			},
		},
	}

	require.Equal(t, expected, c.GetProto())
}

func TestCollectionUTF8DropWhenFallbackBecomesEmpty(t *testing.T) {
	tests := []struct {
		name     string
		metric   telegraf.Metric
		expected []*dto.MetricFamily
	}{
		{
			name: "drop metric when metric name is empty after fallback",
			metric: metric.New(
				string([]byte{0xff}),
				map[string]string{},
				map[string]any{
					string([]byte{0xff}): 42.0,
				},
				time.Unix(0, 0),
			),
			expected: make([]*dto.MetricFamily, 0),
		},
		{
			name: "drop label when label name is empty after fallback",
			metric: metric.New(
				"cpu",
				map[string]string{
					string([]byte{0xff}): "example.org",
				},
				map[string]any{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
			expected: []*dto.MetricFamily{
				{
					Name: new("cpu_time_idle"),
					Help: new(helpString),
					Type: dto.MetricType_UNTYPED.Enum(),
					Metric: []*dto.Metric{
						{
							Label:   make([]*dto.LabelPair, 0),
							Untyped: &dto.Untyped{Value: proto.Float64(42)},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollection(FormatConfig{NameSanitization: "utf8"})
			c.Add(tt.metric, time.Unix(0, 0))
			require.Equal(t, tt.expected, c.GetProto())
		})
	}
}
