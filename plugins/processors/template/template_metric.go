package template

import (
	"time"
	"github.com/influxdata/telegraf"
)

type TemplateMetric struct {
	metric telegraf.Metric
}

func (m *TemplateMetric) Measurement() string {
	panic("not implemented")
}

func (m *TemplateMetric) Tag(key string) string {
	panic("not implemented")
}

func (m *TemplateMetric) Field(key string) interface{} {
	panic("not implemented")
}

func (m *TemplateMetric) Time() time.Time {
	panic("not implemented")
}