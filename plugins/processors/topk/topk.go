package topk

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TopK struct {
	cache map[uint64][]telegraf.Metric
}

func NewTopK() telegraf.Processor{
	topk := &TopK{}
	topk.Reset()
	return topk
}

var sampleConfig = `
`

func (t *TopK) SampleConfig() string {
	return sampleConfig
}

func (t *TopK) Reset() {
	t.cache = make(map[uint64][]telegraf.Metric)
}

func (t *TopK) Description() string {
	return "Print all metrics that pass through this filter."
}

func (t *TopK) Apply(in ...telegraf.Metric) []telegraf.Metric {
	return in
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func init() {
	processors.Add("topk", func() telegraf.Processor {
		return NewTopK()
	})
}
