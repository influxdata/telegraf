package tags

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
## NOTE This processor will overwrite values of tags, that are already
## present in the metric passed through this filter.

## Tags to be added (all values must be strings)
# [processors.tags.add]
#   additional_tag = "tag_value"
`

type TagAdder struct {
	Add map[string]string
}

func (p *TagAdder) SampleConfig() string {
	return sampleConfig
}

func (p *TagAdder) Description() string {
	return "Add all configured tags to all metrics that pass through this filter."
}

func (a *TagAdder) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for key, value := range a.Add {
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
