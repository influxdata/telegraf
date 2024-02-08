//go:generate ../../../tools/readme_config_includer/generator
package clone

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Clone struct {
	NameOverride string
	NamePrefix   string
	NameSuffix   string
	Tags         map[string]string
}

func (*Clone) SampleConfig() string {
	return sampleConfig
}

func (c *Clone) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0, 2*len(in))

	for _, original := range in {
		m := original.Copy()
		if len(c.NameOverride) > 0 {
			m.SetName(c.NameOverride)
		}
		if len(c.NamePrefix) > 0 {
			m.AddPrefix(c.NamePrefix)
		}
		if len(c.NameSuffix) > 0 {
			m.AddSuffix(c.NameSuffix)
		}
		for key, value := range c.Tags {
			m.AddTag(key, value)
		}
		out = append(out, m)
	}

	return append(out, in...)
}

func init() {
	processors.Add("clone", func() telegraf.Processor {
		return &Clone{}
	})
}
