package prometheus

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestCollectionExpire(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		age      time.Duration
		metrics  []telegraf.Metric
		expected []*dto.MetricFamily
	}{
		{
			name: "not expired",
			now:  time.Unix(1, 0),
			age:  10 * time.Second,
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
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
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []*dto.MetricFamily{},
		},
		{
			name: "expired one metric in metric family",
			now:  time.Unix(20, 0),
			age:  10 * time.Second,
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_guest": 42.0,
					},
					time.Unix(15, 0),
				),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollection(FormatConfig{})
			for _, metric := range tt.metrics {
				c.Add(metric)
			}
			c.Expire(tt.now, tt.age)

			actual := c.GetProto()

			require.Equal(t, tt.expected, actual)
		})
	}
}
