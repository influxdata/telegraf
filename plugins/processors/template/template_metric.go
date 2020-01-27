package template

import (
	"github.com/influxdata/telegraf"
	"time"
)

type TemplateMetric struct {
	metric telegraf.Metric
}

func (m *TemplateMetric) Measurement() string {
	return m.Measurement()
}

func (m *TemplateMetric) Tag(key string) string {
	tagString, _ := m.metric.GetTag(key)
	return tagString
}

func (m *TemplateMetric) Field(key string) interface{} {
	field, _ := m.metric.GetField(key)
	return field
}

func (m *TemplateMetric) Time() time.Time {
	return m.metric.Time()
}
