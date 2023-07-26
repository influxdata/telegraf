//go:generate ../../../tools/readme_config_includer/generator
package template

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type TemplateProcessor struct {
	Tag      string          `toml:"tag"`
	Template string          `toml:"template"`
	Log      telegraf.Logger `toml:"-"`

	tmplTag   *template.Template
	tmplValue *template.Template
}

func (*TemplateProcessor) SampleConfig() string {
	return sampleConfig
}

func (r *TemplateProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// for each metric in "in" array
	for _, metric := range in {
		m, ok := metric.(telegraf.TemplateMetric)
		if !ok {
			r.Log.Errorf("metric of type %T is not a template metric", metric)
			continue
		}
		newM := TemplateMetric{m}

		var b strings.Builder
		if err := r.tmplTag.Execute(&b, &newM); err != nil {
			r.Log.Errorf("failed to execute tag name template: %v", err)
			continue
		}
		tag := b.String()

		b.Reset()
		if err := r.tmplValue.Execute(&b, &newM); err != nil {
			r.Log.Errorf("failed to execute value template: %v", err)
			continue
		}
		value := b.String()

		metric.AddTag(tag, value)
	}
	return in
}

func (r *TemplateProcessor) Init() error {
	var err error

	r.tmplTag, err = template.New("tag template").Parse(r.Tag)
	if err != nil {
		return fmt.Errorf("creating tag name template failed: %w", err)
	}

	r.tmplValue, err = template.New("value template").Parse(r.Template)
	if err != nil {
		return fmt.Errorf("creating value template failed: %w", err)
	}
	return nil
}

func init() {
	processors.Add("template", func() telegraf.Processor {
		return &TemplateProcessor{}
	})
}
