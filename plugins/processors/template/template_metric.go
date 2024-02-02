package template

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
)

var (
	onceTagList   sync.Once
	onceFieldList sync.Once
)

type TemplateMetric struct {
	metric telegraf.TemplateMetric
}

func (m *TemplateMetric) Name() string {
	return m.metric.Name()
}

func (m *TemplateMetric) Tag(key string) string {
	return m.metric.Tag(key)
}

func (m *TemplateMetric) Field(key string) interface{} {
	return m.metric.Field(key)
}

func (m *TemplateMetric) Time() time.Time {
	return m.metric.Time()
}

func (m *TemplateMetric) Tags() map[string]string {
	return m.metric.Tags()
}

func (m *TemplateMetric) Fields() map[string]interface{} {
	return m.metric.Fields()
}

func (m *TemplateMetric) String() string {
	return m.metric.String()
}

func (m *TemplateMetric) TagList() map[string]string {
	onceTagList.Do(func() {
		models.PrintOptionValueDeprecationNotice(
			telegraf.Warn, "processors.template", "template", "{{.TagList}}",
			telegraf.DeprecationInfo{
				Since:     "1.28.0",
				RemovalIn: "1.34.0",
				Notice:    "use '{{.Tags}}' instead",
			},
		)
	})
	return m.metric.Tags()
}

func (m *TemplateMetric) FieldList() map[string]interface{} {
	onceFieldList.Do(func() {
		models.PrintOptionValueDeprecationNotice(
			telegraf.Warn, "processors.template", "template", "{{.FieldList}}",
			telegraf.DeprecationInfo{
				Since:     "1.28.0",
				RemovalIn: "1.34.0",
				Notice:    "use '{{.Fields}}' instead",
			},
		)
	})
	return m.metric.Fields()
}
