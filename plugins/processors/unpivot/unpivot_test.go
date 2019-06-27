package unpivot

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestUnpivot(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		unpivot  *Unpivot
		metrics  []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "simple",
			unpivot: &Unpivot{
				TagKey:   "name",
				ValueKey: "value",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{
						"idle_time": int64(42),
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
						"value": int64(42),
					},
					now,
				),
			},
		},
		{
			name: "multi fields",
			unpivot: &Unpivot{
				TagKey:   "name",
				ValueKey: "value",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{
						"idle_time": int64(42),
						"idle_user": int64(43),
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
						"value": int64(42),
					},
					now,
				),
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_user",
					},
					map[string]interface{}{
						"value": int64(43),
					},
					now,
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.unpivot.Apply(tt.metrics...)
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.SortMetrics())
		})
	}
}
