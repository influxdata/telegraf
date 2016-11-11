package utils

import "github.com/influxdata/telegraf"

func IsSelected(metric telegraf.Metric, configuredMetrics []string) bool {
	for _, m := range configuredMetrics {
		if m == metric.Name() {
			return true
		}
	}
	return false
}
