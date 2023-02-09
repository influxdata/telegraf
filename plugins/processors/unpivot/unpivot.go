//go:generate ../../../tools/readme_config_includer/generator
package unpivot

import (
	_ "embed"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Unpivot struct {
	MetricMode string `toml:"metric_mode"`
	TagKey     string `toml:"tag_key"`
	ValueKey   string `toml:"value_key"`
}

func copyWithoutFields(metric telegraf.Metric) telegraf.Metric {
	m := metric.Copy()

	fieldKeys := make([]string, 0, len(m.FieldList()))
	for _, field := range m.FieldList() {
		fieldKeys = append(fieldKeys, field.Key)
	}

	for _, fk := range fieldKeys {
		m.RemoveField(fk)
	}

	return m
}

func (*Unpivot) SampleConfig() string {
	return sampleConfig
}

func (p *Unpivot) Init() error {
	switch p.MetricMode {
	case "", "original":
		p.MetricMode = "original"
	case "field":
	default:
		return fmt.Errorf("unrecognized metric mode: %q", p.MetricMode)
	}

	if p.TagKey == "" {
		p.TagKey = "name"
	}
	if p.ValueKey == "" {
		p.ValueKey = "value"
	}

	return nil
}

func (p *Unpivot) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	fieldCount := 0
	for _, m := range metrics {
		fieldCount += len(m.FieldList())
	}
	results := make([]telegraf.Metric, 0, fieldCount)

	for _, m := range metrics {
		base := copyWithoutFields(m)
		for _, field := range m.FieldList() {
			newMetric := base.Copy()
			newMetric.AddField(p.ValueKey, field.Value)

			switch p.MetricMode {
			case "field":
				newMetric.SetName(field.Key)
			case "", "original":
				newMetric.AddTag(p.TagKey, field.Key)
			}

			results = append(results, newMetric)
		}
		m.Accept()
	}
	return results
}

func init() {
	processors.Add("unpivot", func() telegraf.Processor {
		return &Unpivot{}
	})
}
