package template

import (
	"github.com/influxdata/telegraf"
)

type Replace struct {
	Measurement string `toml:"measurement"`
	Tag         string `toml:"tag"`
	Field       string `toml:"field"`
	Tempate     string `toml:"template"`
}

type Template struct {
	metric telegraf.Metric
}

func (r *Template) Apply(in ...telegraf.Metric) []telegraf.Metric {
// for each metric in "in" array
// convert/wrap metric in TemplateMetric
// supply TemplateMetric and Template from configuration to Template.Execute
// convert TemplateMetric back to metric?
	return in
}
