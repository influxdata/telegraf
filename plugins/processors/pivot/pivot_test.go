package pivot

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestPivot(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		pivot    *Pivot
		metrics  []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "simple",
			pivot: &Pivot{
				TagKey:   "name",
				ValueKey: "value",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{
						"idle_time": int64(42),
					},
					now,
				),
			},
		},
		{
			name: "missing tag",
			pivot: &Pivot{
				TagKey:   "name",
				ValueKey: "value",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"foo": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"foo": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
		},
		{
			name: "missing field",
			pivot: &Pivot{
				TagKey:   "name",
				ValueKey: "value",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"foo": int64(42),
					},
					now,
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"foo": int64(42),
					},
					now,
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.pivot.Apply(tt.metrics...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}
