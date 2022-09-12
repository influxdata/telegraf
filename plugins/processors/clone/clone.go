//go:generate ../../../tools/readme_config_includer/generator
package clone

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
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
	cloned := []telegraf.Metric{}

	for _, metric := range in {
		cloned = append(cloned, metric.Copy())

		if len(c.NameOverride) > 0 {
			metric.SetName(c.NameOverride)
		}
		if len(c.NamePrefix) > 0 {
			metric.AddPrefix(c.NamePrefix)
		}
		if len(c.NameSuffix) > 0 {
			metric.AddSuffix(c.NameSuffix)
		}
		for key, value := range c.Tags {
			metric.AddTag(key, value)
		}
	}
	return append(in, cloned...)
}

func init() {
	processors.Add("clone", func() telegraf.Processor {
		return &Clone{}
	})
}
