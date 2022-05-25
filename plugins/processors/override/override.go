//go:generate ../../../tools/readme_config_includer/generator
package override

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embedd the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type Override struct {
	NameOverride string
	NamePrefix   string
	NameSuffix   string
	Tags         map[string]string
}

func (*Override) SampleConfig() string {
	return sampleConfig
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
