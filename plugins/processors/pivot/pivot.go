package pivot

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Pivot struct {
	TagKey   string `toml:"tag_key"`
	ValueKey string `toml:"value_key"`
}

func (p *Pivot) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, m := range metrics {
		key, ok := m.GetTag(p.TagKey)
		if !ok {
			continue
		}

		value, ok := m.GetField(p.ValueKey)
		if !ok {
			continue
		}

		m.RemoveTag(p.TagKey)
		m.RemoveField(p.ValueKey)
		m.AddField(key, value)
	}
	return metrics
}

func init() {
	processors.Add("pivot", func() telegraf.Processor {
		return &Pivot{}
	})
}
