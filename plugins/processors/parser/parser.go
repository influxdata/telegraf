package parser

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Parser struct {
	Config      parsers.Config `toml:"config"`
	Original    string         `toml:"original"` // merge, replace, or keep (default)
	ParseFields []string       `toml:"parse_fields"`
	Parser      parsers.Parser `toml:"parser"`
}

// holds a default sample config
var SampleConfig = `
  ## specify the name of the field[s] whose value will be parsed
  parse_fields = []

  ## specify what to do with the original message. [merge|replace|keep] default=keep
  original = "keep"

  [processors.parser.config]
    # data_format = "logfmt"
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
		p.Parser, err = parsers.NewParser(&p.Config)
		if err != nil {
			log.Printf("E! [processors.parser] could not create parser: %v", err)
			return metrics
		}
	}

	rMetrics := []telegraf.Metric{}
	if p.Original != "replace" {
		rMetrics = metrics
	}

	name := ""

	for _, metric := range metrics {
		if n := metric.Name(); n != "" {
			name = n
		}
		sMetrics := []telegraf.Metric{}
		for _, key := range p.ParseFields {
			if value, ok := metric.Fields()[key]; ok {
				strVal := fmt.Sprintf("%v", value)
				nMetrics, err := p.parseField(strVal)
				if err != nil {
					log.Printf("E! [processors.parser] could not parse field %v: %v", key, err)
					switch p.Original {
					case "keep":
						return metrics
					case "merge":
						nMetrics = metrics
					}
				}
				sMetrics = append(sMetrics, nMetrics...)
			} else {
				fmt.Println("key not found", key)
			}
		}
		rMetrics = append(rMetrics, p.mergeTagsFields(sMetrics...)...)
	}
	if p.Original == "merge" {
		rMetrics = p.mergeTagsFields(rMetrics...)
	}

	return p.setName(name, rMetrics...)

}

func (p Parser) setName(name string, metrics ...telegraf.Metric) []telegraf.Metric {
	if len(metrics) == 0 {
		return nil
	}

	for i := range metrics {
		metrics[i].SetName(name)
	}

	return metrics
}

func (p Parser) mergeTagsFields(metrics ...telegraf.Metric) []telegraf.Metric {
	if len(metrics) == 0 {
		return nil
	}

	rMetric := metrics[0]
	for _, metric := range metrics {
		for key, field := range metric.Fields() {
			rMetric.AddField(key, field)
		}
		for key, tag := range metric.Tags() {
			rMetric.AddTag(key, tag)
		}
	}
	return []telegraf.Metric{rMetric}
}

func (p *Parser) parseField(value string) ([]telegraf.Metric, error) {
	return p.Parser.Parse([]byte(value))
}

func init() {
	processors.Add("parser", func() telegraf.Processor {
		return &Parser{Original: "keep"}
	})
}
