package global_fields

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type (
	globalFieldsConfig struct {
		Fields []*globalField `toml:"field"`
	}

	globalField struct {
		Name  string      `toml:"name"`
		Value interface{} `toml:"value"`
	}
)

var sampleConfig = `
[[processors.global_fields]]
  [[processors.global_fields.field]]
    Name = "owner"
    Value = "Mr T."
  [[processors.global_fields.field]]
    Name = "age"
    Value = 67
`

func (p *globalFieldsConfig) SampleConfig() string {
	return sampleConfig
}

func (p *globalFieldsConfig) Description() string {
	return "Adds fields to all metrics"
}

func (p *globalFieldsConfig) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		for _, field := range p.Fields {
			if !metric.HasField(field.Name) {
				metric.AddField(field.Name, field.Value)
			}
		}
	}
	return metrics
}

func init() {
	processors.Add("global_fields", func() telegraf.Processor {
		return &globalFieldsConfig{}
	})
}
