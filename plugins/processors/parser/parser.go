package parser

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Parser struct {
	config      parsers.Config
	parseFields []string `toml:"parse_fields"`
	Parser      parsers.Parser
}

// holds a default sample config
var SampleConfig = `

## specify the name of the field[s] whose value will be parsed
parse_fields = []

[processors.parser.config]
  data_format = "logfmt"
  ## additional configurations for parser go here
`

// returns the default config
func (p *Parser) SampleConfig() string {
	return SampleConfig
}

// returns a brief description of the processor
func (p *Parser) Description() string {
	return "Parse a value in a specified field/tag(s) and add the result in a new metric"
}

func (p *Parser) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	if p.Parser == nil {
		var err error
		p.Parser, err = parsers.NewParser(&p.config)
		if err != nil {
			log.Printf("E! [processors.parser] could not create parser: %v", err)
			return metrics
		}
	}

	for _, metric := range metrics {
		for _, key := range p.parseFields {
			value := metric.Fields()[key]
			strVal := fmt.Sprintf("%v", value)
			nMetrics, err := p.parseField(strVal)
			if err != nil {
				log.Printf("E! [processors.parser] could not parse field %v: %v", key, err)
				return metrics
			}
			metrics = append(metrics, nMetrics...)
		}
	}
	return metrics

}

func (p *Parser) parseField(value string) ([]telegraf.Metric, error) {
	return p.Parser.Parse([]byte(value))
}

func init() {
	processors.Add("parser", func() telegraf.Processor {
		return &Parser{}
	})
}
