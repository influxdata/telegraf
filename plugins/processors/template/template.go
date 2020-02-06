package template

import (
	"strings"
	"text/template"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TemplateProcessor struct {
	Tag      string          `toml:"tag"`
	Template string          `toml:"template"`
	Log      telegraf.Logger `toml:"-"`
	tmpl     *template.Template
}

const sampleConfig = `
  ## Tag to set with the output of the template.
  tag = "topic"

  ## Go template used to create the tag value.  In order to ease TOML
  ## escaping requirements, you may wish to use single quotes around the
  ## template string.
  template = '{{ .Tag "hostname" }}.{{ .Tag "level" }}'
`

func (r *TemplateProcessor) SampleConfig() string {
	return sampleConfig
}

func (r *TemplateProcessor) Description() string {
	return "Uses a Go template to create a new tag"
}

func (r *TemplateProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// for each metric in "in" array
	for _, metric := range in {
		var b strings.Builder
		newM := TemplateMetric{metric}

		// supply TemplateMetric and Template from configuration to Template.Execute
		err := r.tmpl.Execute(&b, &newM)
		if err != nil {
			r.Log.Errorf("failed to execute template: %v", err)
			continue
		}

		metric.AddTag(r.Tag, b.String())
	}
	return in
}

func (r *TemplateProcessor) Init() error {
	// create template
	t, err := template.New("configured_template").Parse(r.Template)

	r.tmpl = t
	return err
}

func init() {
	processors.Add("template", func() telegraf.Processor {
		return &TemplateProcessor{}
	})
}
