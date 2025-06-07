package template

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

var (
	onceTagList   sync.Once
	onceFieldList sync.Once
)

type templateMetric struct {
	metric telegraf.TemplateMetric
}

func (m *templateMetric) Name() string {
	return m.metric.Name()
}

func (m *templateMetric) Tag(key string) string {
	return m.metric.Tag(key)
}

func (m *templateMetric) Field(key string) interface{} {
	return m.metric.Field(key)
}

func (m *templateMetric) Time() time.Time {
	return m.metric.Time()
}

func (m *templateMetric) Tags() map[string]string {
	return m.metric.Tags()
}

func (m *templateMetric) Fields() map[string]interface{} {
	return m.metric.Fields()
}

func (m *templateMetric) String() string {
	return m.metric.String()
}

func (m *templateMetric) TagList() map[string]string {
	onceTagList.Do(func() {
		config.PrintOptionValueDeprecationNotice(
			"processors.template", "template", "{{.TagList}}",
			telegraf.DeprecationInfo{
				Since:     "1.28.0",
				RemovalIn: "1.34.0",
				Notice:    "use '{{.Tags}}' instead",
			},
		)
	})
	return m.metric.Tags()
}

func (m *templateMetric) FieldList() map[string]interface{} {
	onceFieldList.Do(func() {
		config.PrintOptionValueDeprecationNotice(
			"processors.template", "template", "{{.FieldList}}",
			telegraf.DeprecationInfo{
				Since:     "1.28.0",
				RemovalIn: "1.34.0",
				Notice:    "use '{{.Fields}}' instead",
			},
		)
	})
	return m.metric.Fields()
}
