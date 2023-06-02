package processors

import (
	"fmt"
	"github.com/influxdata/telegraf"
)

const (
	metricTypeTagName = "metric_type"
)

// MetricType processor removes metric_type tag if it is present and appends it to metric name as a suffix
type MetricType struct {
}

// Process implements the core of MetricType processor logic
func (m *MetricType) Process(metric telegraf.Metric) error {
	if tagValue, set := metric.GetTag(metricTypeTagName); set {
		metric.RemoveTag(metricTypeTagName)

		for _, f := range metric.FieldList() {
			f.Key = fmt.Sprintf("%s.%s", f.Key, tagValue)
		}
	}
	return nil
}

// Close clears the resources processor used, no-op in this case
func (m *MetricType) Close() {}

// NewMetricType creates a new MetricType processor
func NewMetricType() MetricProcessor {
	return &MetricType{}
}
