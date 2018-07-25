package fieldparser

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
)

type FieldParser struct {
	config      parsers.Config
	parseFields []string `toml:"parse_fields"`
	parseTags   []string `toml:"parse_tags"`
	Parser      parsers.Parser
}

// holds a default sample config
var SampleConfig = `
## specify the name of the tag[s] whose value will be parsed
parse_tags = []

## specify the name of the field[s] whose value will be parsed
parse_fields = []

[processors.fieldparser.config]
  data_format = "logfmt"
  ## additional configurations for parser go here
`

// returns the default config
func (p *FieldParser) SampleConfig() string {
	return SampleConfig
}

// returns a brief description of the processor
func (p *FieldParser) Description() string {
	return "Parse a value in a specified field/tag(s) and add the result in a new metric"
}

func (p *FieldParser) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	newMetrics := make([]telegraf.Metric, 0)

	//load input metrics into newMetrics
	newMetrics = append(newMetrics, metrics...)
	if p.Parser == nil {
		var err error
		p.Parser, err = parsers.NewParser(&p.config)
		if err != nil {
			log.Printf("E! [processors.fieldparser] could not create parser: %v", err)
			return newMetrics
		}
	}

	for _, metric := range metrics {
		for _, key := range p.parseFields {
			value := metric.Fields()[key]
			nMetrics, err := p.parseField(value.(string))
			if err != nil {
				log.Printf("E! [processors.fieldparser] could not parse field %v: %v", key, err)
				return newMetrics
			}
			newMetrics = append(newMetrics, nMetrics...)
		}
		for _, key := range p.parseTags {
			value := metric.Tags()[key]
			nMetrics, err := p.parseField(value)
			if err != nil {
				log.Printf("E! [processors.fieldparser] could not parse field %v: %v", key, err)
				return newMetrics
			}
			newMetrics = append(newMetrics, nMetrics...)
		}
	}
	return newMetrics

}

func (p *FieldParser) parseField(value string) ([]telegraf.Metric, error) {
	return p.Parser.Parse([]byte(value))
}

func init() {
	processors.Add("fieldparser", func() telegraf.Processor {
		return &FieldParser{}
	})
}
