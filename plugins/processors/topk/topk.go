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

type Measurements []telegraf.Metric

func (m Measurements) Len() int {
	return len(m)
}

func (m Measurements) Less(i, j int) bool {
	iv, iok := convert(m[i].Fields()["value"])
	jv, jok := convert(m[j].Fields()["value"])
	if  iok && jok && (iv < jv) {
		return true
	} else {
		return false
	}
}

func (m Measurements) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

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
