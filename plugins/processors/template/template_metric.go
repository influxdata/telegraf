package template

import (
	"time"
	"github.com/influxdata/telegraf"
)

type TemplateMetric struct {
	metric telegraf.Metric
}

func (m *TemplateMetric) Measurement() string {
	return m.Measurement()
}

func (m *TemplateMetric) Tag(key string) string {
	return m.Tag(key)
}

func (m *TemplateMetric) Field(key string) interface{} {
	panic("not implemented")
}

func (m *TemplateMetric) Time() time.Time {
	panic("not implemented")
}