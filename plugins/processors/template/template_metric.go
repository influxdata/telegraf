package template

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
)

type TemplateMetric struct {
	metric telegraf.Metric
}

func (m *TemplateMetric) Name() string {
	return m.metric.Name()
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

func (m *TemplateMetric) String() string {
	return fmt.Sprint(m.metric)
}

func (m *TemplateMetric) TagList() map[string]string {
	return m.metric.Tags()
}

func (m *TemplateMetric) FieldList() map[string]interface{} {
	return m.metric.Fields()
}
