package parser

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Parser struct {
	parsers.Config
	DropOriginal bool            `toml:"drop_original"`
	Merge        string          `toml:"merge"`
	ParseFields  []string        `toml:"parse_fields"`
	Log          telegraf.Logger `toml:"-"`
	parser       telegraf.Parser
}

func (p *Parser) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	if p.parser == nil {
		var err error
		p.parser, err = parsers.NewParser(&p.Config)
		if err != nil {
			p.Log.Errorf("could not create parser: %v", err)
			return metrics
		}
		models.SetLoggerOnPlugin(p.parser, p.Log)
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
							p.Log.Errorf("could not parse field %s: %v", key, err)
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
						p.Log.Errorf("field '%s' not a string, skipping", key)
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
	return p.parser.Parse([]byte(value))
}

func init() {
	processors.Add("parser", func() telegraf.Processor {
		return &Parser{DropOriginal: false}
	})
}
