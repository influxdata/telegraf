package topk

import (
	"sort"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TopK struct {
	Metric             string
	Period             int
	K                  int
	Field              string
	Aggregation        string
	Tags               []string
        RevertTagMatch     bool `toml:"revert_tag_match"`
        DropNonMatching    bool `toml:"drop_non_matching"`
	DropNonTop         bool `toml:"top"`
	PositionField      string `toml:"position_field"`
	AggregationField   string `toml:"aggregation_field"`

	cache map[uint64][]telegraf.Metric
	last_report time.Time
}

func NewTopK() telegraf.Processor{
	// Create object
	topk := &TopK{}

	// Setup defaults
	topk.Period = 10
	topk.K = 10
	topk.Aggregation = "avg"
	topk.Field = "value"
	topk.Tags = []string{"*"}
	topk.RevertTagMatch = false
	topk.DropNonMatching = false
	topk.DropNonTop = true
	topk.PositionField = ""
	topk.AggregationField = ""

	// Initialize cache
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
	t.last_report = time.Now()
}

func (t *TopK) Description() string {
	return "Print all metrics that pass through this filter."
}

func (t *TopK) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// Add the metrics received to our internal cache
	for _, m := range in {

		// Initialize the key with an empty list if necessary
		if _, ok := t.cache[m.HashID()]; !ok {
			t.cache[m.HashID()] = make([]telegraf.Metric, 0, 10)
		}

		// Append the metric to the corresponding key list
		t.cache[m.HashID()] = append(t.cache[m.HashID()], m)
	}

	// If enough time has passed
	elapsed := time.Since(t.last_report)
	if elapsed >= time.Second*10 {
		// Sort the keys by the selected field
		for _, ms := range t.cache {
			sort.Reverse(Measurements(ms))
		}
		
		// Create a one dimentional list with the top K metrics of each key
		ret := make([]telegraf.Metric, 0, 100)
		for _, ms := range t.cache {
			ret = append(ret, ms[0:min(len(ms), 10)]...)
		}

		t.Reset()

		return ret
	}

	return []telegraf.Metric{}
}

func min(a, b int) int   {
	if a > b { return b }
	return a
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
