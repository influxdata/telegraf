package template

import (
	"github.com/influxdata/telegraf"
	"text/template"
	"strings"
)

type TemplateProcessor struct {
	Tag         string `toml:"tag"`
	Template    string `toml:"template"`
}

func (r *TemplateProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// create template
	tmpl, err := template.New("test").Parse(r.Template)
	if err != nil { panic(err) }

	// for each metric in "in" array
	for _, metric := range in {
		var b strings.Builder
		//newM := TemplateMetric{metric}

		// supply TemplateMetric and Template from configuration to Template.Execute
		err := tmpl.Execute(&b, metric)
		if err != nil { panic(err) }

		metric.AddTag(r.Tag, b.String())
	}

// convert/wrap metric in TemplateMetric
// convert TemplateMetric back to metric?
	return in
}
