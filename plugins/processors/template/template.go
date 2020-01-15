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
	tmpl	 *template.Template
}

func (r *TemplateProcessor) Description() string {
	return ""
}

func (r *TemplateProcessor) SampleConfig() string {
	return ""
}

func (r *TemplateProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// for each metric in "in" array
	for _, metric := range in {
		var b strings.Builder
		//newM := TemplateMetric{metric}

		// supply TemplateMetric and Template from configuration to Template.Execute
		err := r.tmpl.Execute(&b, metric)
		if err != nil {
			panic(err)
		}

		metric.AddTag(r.Tag, b.String())
	}

	// convert/wrap metric in TemplateMetric
	// convert TemplateMetric back to metric?
	return in
}

func (r *TemplateProcessor) Init() error{
	// create template
	t, err := template.New("test").Parse(r.Template)
	
	r.tmpl = t
	return err
}

func init() {
	processors.Add("printer", func() telegraf.Processor {
		return &TemplateProcessor{}
	})
}
