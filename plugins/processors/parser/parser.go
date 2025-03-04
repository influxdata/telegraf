//go:generate ../../../tools/readme_config_includer/generator
package parser

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	gobin "encoding/binary"
	"fmt"
	"slices"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Parser struct {
	DropOriginal bool            `toml:"drop_original"`
	Merge        string          `toml:"merge"`
	ParseFields  []string        `toml:"parse_fields"`
	Base64Fields []string        `toml:"parse_fields_base64"`
	ParseTags    []string        `toml:"parse_tags"`
	Log          telegraf.Logger `toml:"-"`
	parser       telegraf.Parser
}

func (p *Parser) Init() error {
	switch p.Merge {
	case "", "override", "override-with-timestamp":
	default:
		return fmt.Errorf("unrecognized merge value: %s", p.Merge)
	}

	return nil
}

func (*Parser) SampleConfig() string {
	return sampleConfig
}

func (p *Parser) SetParser(parser telegraf.Parser) {
	p.parser = parser
}

func (p *Parser) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	results := make([]telegraf.Metric, 0, len(metrics))
	for _, metric := range metrics {
		var newMetrics []telegraf.Metric
		if !p.DropOriginal {
			newMetrics = append(newMetrics, metric)
		} else {
			metric.Drop()
		}

		// parse fields
		for _, field := range metric.FieldList() {
			plain := slices.Contains(p.ParseFields, field.Key)
			b64 := slices.Contains(p.Base64Fields, field.Key)

			if !plain && !b64 {
				continue
			}

			if plain && b64 {
				p.Log.Errorf("field %s is listed in both parse fields and base64 fields; skipping", field.Key)
				continue
			}

			value, err := toBytes(field.Value)
			if err != nil {
				p.Log.Errorf("could not convert field %s: %v; skipping", field.Key, err)
				continue
			}

			if b64 {
				decoded := make([]byte, base64.StdEncoding.DecodedLen(len(value)))
				n, err := base64.StdEncoding.Decode(decoded, value)
				if err != nil {
					p.Log.Errorf("could not decode base64 field %s: %v; skipping", field.Key, err)
					continue
				}
				value = decoded[:n]
			}

			fromFieldMetric, err := p.parser.Parse(value)
			if err != nil {
				p.Log.Errorf("could not parse field %s: %v", field.Key, err)
				continue
			}

			for _, m := range fromFieldMetric {
				// The parser get the parent plugin's name as
				// default measurement name. Thus, in case the
				// parsed metric does not provide a name itself,
				// the parser  will return 'parser' as we are in
				// processors.parser. In those cases we want to
				// keep the original metric name.
				if m.Name() == "" || m.Name() == "parser" {
					m.SetName(metric.Name())
				}
			}

			// multiple parsed fields shouldn't create multiple
			// metrics so we'll merge tags/fields down into one
			// prior to returning.
			newMetrics = append(newMetrics, fromFieldMetric...)
		}

		// parse tags
		for _, key := range p.ParseTags {
			if value, ok := metric.GetTag(key); ok {
				fromTagMetric, err := p.parseValue(value)
				if err != nil {
					p.Log.Errorf("could not parse tag %s: %v", key, err)
				}

				for _, m := range fromTagMetric {
					// The parser get the parent plugin's name as
					// default measurement name. Thus, in case the
					// parsed metric does not provide a name itself,
					// the parser  will return 'parser' as we are in
					// processors.parser. In those cases we want to
					// keep the original metric name.
					if m.Name() == "" || m.Name() == "parser" {
						m.SetName(metric.Name())
					}
				}

				newMetrics = append(newMetrics, fromTagMetric...)
			}
		}

		if len(newMetrics) == 0 {
			continue
		}

		if p.Merge == "override" {
			results = append(results, merge(newMetrics[0], newMetrics[1:]))
		} else if p.Merge == "override-with-timestamp" {
			results = append(results, mergeWithTimestamp(newMetrics[0], newMetrics[1:]))
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

func mergeWithTimestamp(base telegraf.Metric, metrics []telegraf.Metric) telegraf.Metric {
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			base.AddField(field.Key, field.Value)
		}
		for _, tag := range metric.TagList() {
			base.AddTag(tag.Key, tag.Value)
		}
		base.SetName(metric.Name())
		if !metric.Time().IsZero() {
			base.SetTime(metric.Time())
		}
	}
	return base
}

func (p *Parser) parseValue(value string) ([]telegraf.Metric, error) {
	return p.parser.Parse([]byte(value))
}

func toBytes(value interface{}) ([]byte, error) {
	if v, ok := value.(string); ok {
		return []byte(v), nil
	}

	var buf bytes.Buffer
	if err := gobin.Write(&buf, internal.HostEndianness, value); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func init() {
	processors.Add("parser", func() telegraf.Processor {
		return &Parser{DropOriginal: false}
	})
}
