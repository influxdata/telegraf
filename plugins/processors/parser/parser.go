package parser

import (
	"log"
	"reflect"

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

	toReturnMetrics := []telegraf.Metric{}
	if p.Original != "replace" {
		toReturnMetrics = metrics
	}

	name := ""

	for _, metric := range metrics {
		if n := metric.Name(); n != "" {
			name = n
		}
		combinedParsedMetric := []telegraf.Metric{}
		for _, key := range p.ParseFields {
			fields := metric.FieldList()
			for _, field := range fields {
				if field.Key == key {
					if reflect.TypeOf(field.Value).String() != "string" {
						log.Printf("E! [processors.parser] field '%v' not a string, skipping", key)
						continue
					}
					fromFieldMetric, err := p.parseField(field.Value.(string))
					if err != nil {
						log.Printf("E! [processors.parser] could not parse field %v: %v", key, err)
						switch p.Original {
						case "keep":
							continue
						case "merge":
							fromFieldMetric = metrics
						}
					}
					// multiple parsed fields shouldn't create multiple metrics so we'll merge tags/fields down into one prior to returning.
					combinedParsedMetric = append(combinedParsedMetric, fromFieldMetric...)
				}
			}
		}
		toReturnMetrics = append(toReturnMetrics, p.mergeTagsFields(combinedParsedMetric...)...)
	}
	if p.Original == "merge" {
		toReturnMetrics = p.mergeTagsFields(toReturnMetrics...)
	}

	return p.setName(name, toReturnMetrics...)

}

func (p Parser) setName(name string, metrics ...telegraf.Metric) []telegraf.Metric {
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
		for _, field := range metric.FieldList() {
			rMetric.AddField(field.Key, field.Value)
		}
		for _, tag := range metric.TagList() {
			rMetric.AddTag(tag.Key, tag.Value)
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
