package override

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Override struct {
	NameOverride string
	NamePrefix   string
	NameSuffix   string
	Tags         map[string]string
}

func (p *Override) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		if len(p.NameOverride) > 0 {
			metric.SetName(p.NameOverride)
		}
		if len(p.NamePrefix) > 0 {
			metric.AddPrefix(p.NamePrefix)
		}
		if len(p.NameSuffix) > 0 {
			metric.AddSuffix(p.NameSuffix)
		}
		for key, value := range p.Tags {
			metric.AddTag(key, value)
		}
	}
	return in
}

func init() {
	processors.Add("override", func() telegraf.Processor {
		return &Override{}
	})
}
