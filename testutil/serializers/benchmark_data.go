package testutil

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// BenchmarkMetrics returns a set of metrics for benchmarking.
func BenchmarkMetrics() [4]telegraf.Metric {
	return [4]telegraf.Metric{
		metric.New("cpu",
			map[string]string{
				"cpu":  "cpu0",
				"host": "realHost",
			},
			map[string]interface{}{
				"usage_idle": 91.5,
			},
			time.Unix(1787161794, 0),
		),
		metric.New("cpu",
			map[string]string{
				"cpu":  "cpu0",
				"host": "realHost",
			},
			map[string]interface{}{
				"usage_idle": 91,
			},
			time.Unix(1787161794, 0),
		),
		metric.New("cpu",
			map[string]string{
				"cpu":  "cpu0",
				"host": "realHost",
			},
			map[string]interface{}{
				"usage_idle": true,
			},
			time.Unix(1787161794, 0),
		),
		metric.New("cpu",
			map[string]string{
				"cpu":  "cpu0",
				"host": "realHost",
			},
			map[string]interface{}{
				"usage_idle": false,
			},
			time.Unix(1787161794, 0),
		),
	}
}
