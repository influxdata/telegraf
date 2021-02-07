package parser

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Parser struct {
	parsers.Config
	DropOriginal bool     `toml:"drop_original"`
	Merge        string   `toml:"merge"`
	ParseFields  []string `toml:"parse_fields"`
	Parser       parsers.Parser
}

var SampleConfig = `
  ## The name of the fields whose value will be parsed.
  parse_fields = []

  ## If true, incoming metrics are not emitted.
  drop_original = false

  ## If set to override, emitted metrics will be merged by overriding the
  ## original metric using the newly parsed metrics.
  merge = "override"

  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (p *Parser) SampleConfig() string {
	return SampleConfig
}

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

	results := []telegraf.Metric{}

	for _, metric := range metrics {
		newMetrics := []telegraf.Metric{}
		if !p.DropOriginal {
			newMetrics = append(newMetrics, metric)
		}

		for _, key := range p.ParseFields {
			for _, field := range metric.FieldList() {
				if field.Key == key {
					switch value := field.Value.(type) {
					case string:
						fromFieldMetric, err := p.parseField(value)
						if err != nil {
							log.Printf("E! [processors.parser] could not parse field %s: %v", key, err)
						}

						for _, m := range fromFieldMetric {
							if m.Name() == "" {
								m.SetName(metric.Name())
							}
						}

						// multiple parsed fields shouldn't create multiple
						// metrics so we'll merge tags/fields down into one
						// prior to returning.
						newMetrics = append(newMetrics, fromFieldMetric...)
					default:
						log.Printf("E! [processors.parser] field '%s' not a string, skipping", key)
					}
				}
			}
		}

		if len(newMetrics) == 0 {
			continue
		}

		if p.Merge == "override" {
			results = append(results, merge(newMetrics[0], newMetrics[1:]))
		} else {
			results = append(results, newMetrics...)
		}
	}
	return results
}

func merge(base telegraf.Metric, metrics []telegraf.Metric) telegraf.Metric {
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			base.AddField(field.Key, field.Value)
		}
		for _, tag := range metric.TagList() {
			base.AddTag(tag.Key, tag.Value)
		}
		base.SetName(metric.Name())
	}
	return base
}

func (p *Parser) parseField(value string) ([]telegraf.Metric, error) {
	return p.Parser.Parse([]byte(value))
}

func init() {
	processors.Add("parser", func() telegraf.Processor {
		return &Parser{DropOriginal: false}
	})
}
