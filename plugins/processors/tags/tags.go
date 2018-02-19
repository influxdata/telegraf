package tags

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
## NOTE This processor will overwrite values of tags, that are already
## present in the metric passed through this filter.

## Tags to be added (all values must be strings)
# [processors.tags.tags]
#   additional_tag = "tag_value"
`

type TagAdder struct {
	NameOverride string
	NamePrefix   string
	NameSuffix   string
	Tags         map[string]string
}

func (p *TagAdder) SampleConfig() string {
	return sampleConfig
}

func (p *TagAdder) Description() string {
	return "Add all configured tags to all metrics that pass through this filter."
}

func (a *TagAdder) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		if len(a.NameOverride) > 0 {
			metric.SetName(a.NameOverride)
		}
		if len(a.NamePrefix) > 0 {
			metric.SetPrefix(a.NamePrefix)
		}
		if len(a.NameSuffix) > 0 {
			metric.SetSuffix(a.NameSuffix)
		}
		for key, value := range a.Tags {
			metric.AddTag(key, value)
		}
	}
	return in
}

func init() {
	processors.Add("tags", func() telegraf.Processor {
		return &TagAdder{}
	})
}
