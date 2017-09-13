package topk

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TopK struct {
}

var sampleConfig = `
`

func (p *TopK) SampleConfig() string {
	return sampleConfig
}

func (p *TopK) Description() string {
	return "Print all metrics that pass through this filter."
}

func (p *TopK) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		fmt.Println(metric.String())
	}
	return in
}

func init() {
	processors.Add("topk", func() telegraf.Processor {
		return &TopK{}
	})
}
