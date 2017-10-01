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
	last_aggregation time.Time
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
[[processors.topk]]
  metric = "cpu"               # Which metric to filter. No default. Mandatory
  period = 10                  # How many seconds between aggregations. Default: 10
  k = 10                       # How many top metrics to return. Default: 10
  field = "user"               # Over which field is the aggregation done. Default: "value"
  tags = ["node-1", "east"]    # List of tags regexes to match against. Default: "*"
  aggregation = "avg"          # What aggregation to use over time. Default: "avg". Options: sum, avg, min, ma
  revert_tag_match = false     # Whether or not to invert the tag match
  drop_non_matching = false    # Whether or not to drop all non matching measurements (for the selected metric only). Default: False
  drop_non_top = true          # Whether or not to drop measurements that do not reach the top k: Default: True
  position_field = "telegraf_topk_position"       # Field to add to the top k measurements, with their position as value. Default: "" (deactivated)
  aggregation_field = "telegraf_topk_aggregation" # Field with the value of the computed aggregation. Default: "" (deactivated)
`

type Measurements struct {
	metrics []telegraf.Metric
	field   string
}

func (m Measurements) Len() int {
	return len(m.metrics)
}

func (m Measurements) Less(i, j int) bool {
	iv, iok := convert(m.metrics[i].Fields()["value"])
	jv, jok := convert(m.metrics[j].Fields()["value"])
	if  iok && jok && (iv < jv) {
		return true
	} else {
		return false
	}
}

func (m Measurements) Swap(i, j int) {
	m.metrics[i], m.metrics[j] = m.metrics[j], m.metrics[i]
}

func (t *TopK) SampleConfig() string {
	return sampleConfig
}

func (t *TopK) Reset() {
	t.cache = make(map[uint64][]telegraf.Metric)
	t.last_aggregation = time.Now()
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
	elapsed := time.Since(t.last_aggregation)
	if elapsed >= time.Second * time.Duration(t.Period) {
		// Sort the keys by the selected field TODO: Make the field configurable
		if t.Field == "" { t.Field = "value"} // Setup the default value for the field to sort
		for _, ms := range t.cache {
			sort.Reverse(Measurements{metrics: ms, field: t.Field})
		}
		
		// Create a one dimentional list with the top K metrics of each key
		ret := make([]telegraf.Metric, 0, 100)
		for _, ms := range t.cache {
			ret = append(ret, ms[0:min(len(ms), t.K)]...)
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
