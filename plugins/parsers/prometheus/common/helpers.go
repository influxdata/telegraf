package common

import (
	"github.com/influxdata/telegraf"
	dto "github.com/prometheus/client_model/go"
)

func ValueType(mt dto.MetricType) telegraf.ValueType {
	switch mt {
	case dto.MetricType_COUNTER:
		return telegraf.Counter
	case dto.MetricType_GAUGE:
		return telegraf.Gauge
	case dto.MetricType_SUMMARY:
		return telegraf.Summary
	case dto.MetricType_HISTOGRAM:
		return telegraf.Histogram
	default:
		return telegraf.Untyped
	}
}

// Get labels from metric
func MakeLabels(m *dto.Metric, defaultTags map[string]string) map[string]string {
	result := map[string]string{}

	for key, value := range defaultTags {
		result[key] = value
	}

	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}

	return result
}
