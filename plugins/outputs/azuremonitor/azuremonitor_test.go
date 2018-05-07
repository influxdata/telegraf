package azuremonitor

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// MockMetrics returns a mock []telegraf.Metric object for using in unit tests
// of telegraf output sinks.
func getMockMetrics() []telegraf.Metric {
	metrics := make([]telegraf.Metric, 0)
	// Create a new point batch
	metrics = append(metrics, getTestMetric(1.0))
	return metrics
}

// TestMetric Returns a simple test point:
//     measurement -> "test1" or name
//     tags -> "tag1":"value1"
//     value -> value
//     time -> time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
func getTestMetric(value interface{}, name ...string) telegraf.Metric {
	if value == nil {
		panic("Cannot use a nil value")
	}
	measurement := "test1"
	if len(name) > 0 {
		measurement = name[0]
	}
	tags := map[string]string{"tag1": "value1"}
	pt, _ := metric.New(
		measurement,
		tags,
		map[string]interface{}{"value": value},
		time.Now().UTC(),
	)
	return pt
}
