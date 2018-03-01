package override

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
## NOTE This processor will override names, name prefixes, name suffixes and
## values of tags, that are already present in the metric passed through this
## filter.

## All modifications on inputs and aggregators can be overridden:
# name_override = "new name"
#	name_prefix = "new name_prefix"
#	name_suffix = "new name_suffix"

## Tags to be added (all values must be strings)
# [processors.overide.tags]
#   additional_tag = "tag_value"
`

type Override struct {
	NameOverride string
	NamePrefix   string
	NameSuffix   string
	Tags         map[string]string
}

func (p *Override) SampleConfig() string {
	return sampleConfig
}

func (p *Override) Description() string {
	return "Add all configured tags to all metrics that pass through this filter."
}

func (p *Override) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		if len(p.NameOverride) > 0 {
			metric.SetName(p.NameOverride)
		}
		if len(p.NamePrefix) > 0 {
			metric.SetPrefix(p.NamePrefix)
		}
		if len(p.NameSuffix) > 0 {
			metric.SetSuffix(p.NameSuffix)
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
