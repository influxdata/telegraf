package template

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"strings"
	"text/template"
)

type TemplateProcessor struct {
	Tag      string `toml:"tag"`
	Template string `toml:"template"`
	tmpl     *template.Template
}

const sampleConfig = `
  ## Concatenate two tags to create a new tag
  # [[processors.template]]
  #   ## Tag to create
  #   tag = "topic"
  #   ## Template to create tag
  # Note: Single quotes (') are used, so the double quotes (") don't need escaping (\")
  #   template = '{{.Tag "hostname"}}.{{ .Tag "level" }}'
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
			panic(err)
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
	processors.Add("printer", func() telegraf.Processor {
		return &TemplateProcessor{}
	})
}
