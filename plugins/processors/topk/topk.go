package topk

import (
	"sort"
	"time"
	"regexp"
	"fmt"
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TopK struct {
	Period             int
	K                  int
	Fields             []string
	Aggregation        string
	GroupBy            []string `toml:"group_by"`
	GroupByMetricName  bool `toml:"group_by_metric_name"`
	GroupByTag         string `toml:"group_by_tag"`
	DropNoGroup        bool `toml:"drop_no_group"`
        Bottomk            bool
	DropNonTop         bool `toml:"drop_non_top"`
	PositionField      string `toml:"position_field"`
	AggregationField   string `toml:"aggregation_field"`

	cache map[string][]telegraf.Metric
	metric_regex *regexp.Regexp
	tags_regexes map[string]*regexp.Regexp
	last_aggregation time.Time
}

func NewTopK() TopK {
	// Create object
	topk := TopK{}

	// Setup defaults
	topk.Period = 10
	topk.K = 10
	topk.Fields = []string{"value"}
	topk.Aggregation = "avg"
	topk.GroupBy = nil
	topk.GroupByMetricName = false
	topk.GroupByTag = ""
	topk.DropNoGroup = true
	topk.DropNonTop = true
	topk.PositionField = ""
	topk.AggregationField = ""

	// Initialize cache
	topk.Reset()

	return topk
}

func NewTopKProcessor() telegraf.Processor{
	topk := NewTopK()
	return &topk
}

var sampleConfig = `
[[processors.topk]]
  metric = "cpu"               # Which metric to filter. No default. Mandatory
  period = 10                  # How many seconds between aggregations. Default: 10
  k = 10                       # How many top metrics to return. Default: 10
  field = "user"               # Over which field is the aggregation done. Default: "value"
  tags = ["node-1", "east"]    # List of tags regexes to match against. Default: "*"
  aggregation = "avg"          # What aggregation to use over time. Default: "avg". Options: sum, avg, min, max
  revert_tag_match = false     # Whether or not to invert the tag match
  drop_non_matching = false    # Whether or not to drop all non matching measurements (for the selected metric only). Default: False
  drop_non_top = true          # Whether or not to drop measurements that do not reach the top k: Default: True
  position_field = "telegraf_topk_position"       # Field to add to the top k measurements, with their position as value. Default: "" (deactivated)
  aggregation_field = "telegraf_topk_aggregation" # Field with the value of the computed aggregation. Default: "" (deactivated)
`

type MetricAggregation struct {
	groupbykey string
	values map[string]float64
}

type Aggregations struct {
	metrics []MetricAggregation
	field string
}

func (ags Aggregations) Len() int {
	return len(ags.metrics)
}

func (ags Aggregations) Less(i, j int) bool {
	iv := ags.metrics[i].values[ags.field]
	jv := ags.metrics[j].values[ags.field]
	if (iv < jv) {
		return true
	} else {
		return false
	}
}

func (ags Aggregations) Swap(i, j int) {
	ags.metrics[i], ags.metrics[j] = ags.metrics[j], ags.metrics[i]
}

func sort_metrics(metrics []MetricAggregation, field string, reverse bool){
	aggs := Aggregations{metrics: metrics, field: field}
	if reverse {
		sort.Sort(aggs)
	} else {
		sort.Sort(sort.Reverse(aggs))
	}
}

func (t *TopK) SampleConfig() string {
	return sampleConfig
}

func (t *TopK) Reset() {
	t.cache = make(map[string][]telegraf.Metric)
	t.last_aggregation = time.Now()
}

func (t *TopK) Description() string {
	return "Print all metrics that pass through this filter."
}

func (t *TopK) generate_groupby_key(m telegraf.Metric) string {
	groupkey := ""
	if t.GroupByMetricName {
		groupkey += m.Name() + "&"
	}
	for _, tag := range(t.GroupBy) {
		tag_value, ok := m.Tags()[tag]
		if ok {
			groupkey += tag + "=" + tag_value + "&"
		}
	}

	if groupkey == "" && ! t.DropNoGroup {
		groupkey = "<<default_groupby_key>>"
	}

	return groupkey
}

func (t *TopK) group_by(m telegraf.Metric) {
	// Generate the metric group key
	groupkey := t.generate_groupby_key(m)

	// If the groupkey is empty, it means we are supposed to drop this metric
	if groupkey == "" {
		return
	}

	// Initialize the key with an empty list if necessary
	if _, ok := t.cache[groupkey]; !ok {
		t.cache[groupkey] = make([]telegraf.Metric, 0, 10)
	}

	// Append the metric to the corresponding key list
	t.cache[groupkey] = append(t.cache[groupkey], m)
}

func (t *TopK) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// Add the metrics received to our internal cache
	var ret []telegraf.Metric = make([]telegraf.Metric, 0, 0)
	for _, m := range in {
		t.group_by(m)
	}

	// If enough time has passed
	elapsed := time.Since(t.last_aggregation)
	if elapsed >= time.Second * time.Duration(t.Period) {
		// Generate aggregations list using the selected fields
		aggregations := make([]MetricAggregation, 0, 100)
		var aggregator func([]telegraf.Metric, []string) map[string]float64 = t.get_aggregation_function(t.Aggregation);
		for k, ms := range t.cache {
			aggregations = append(aggregations, MetricAggregation{groupbykey: k, values: aggregator(ms, t.Fields)})
		}

		// Get the top K metrics for each field and add them to the return value
		added_keys := make(map[string]bool)
		agg_field := t.AggregationField
		pos_field := t.PositionField
		group_tag := t.GroupByTag
		for _, field := range(t.Fields) {

			// Sort the aggregations
			sort_metrics(aggregations, field, t.Bottomk)

			// Create a one dimentional list with the top K metrics of each key
			for i, ag := range aggregations[0:min(t.K, len(aggregations))] {

				// Check whether of not we need to add fields of tags to the selected metrics
				if agg_field != "" || pos_field != "" || group_tag != "" {
					for _, m := range(t.cache[ag.groupbykey]) {
						if agg_field != "" && m.HasField(field){
							m.AddField(agg_field+"_"+field, ag.values[field])
						}
						if pos_field != "" {
							m.AddField(pos_field+"_"+field, i+1) //+1 to it starts from 1
						}
						if group_tag != "" {
							m.AddTag(group_tag, ag.groupbykey) //+1 to it starts from 1
						}
					}
				}

				// Add metrics if we have not already appended them to the return value
				_, ok := added_keys[ag.groupbykey]
				if ! ok {
					ret = append(ret, t.cache[ag.groupbykey]...)
					added_keys[ag.groupbykey] = true
				}
			}
		}

		//Lastly, if we were instructed to not drop the bottom metrics, append them as is to the output
		if ! t.DropNonTop {
			for _, ag := range aggregations {
				_, ok := added_keys[ag.groupbykey]
				if ! ok {
					ret = append(ret, t.cache[ag.groupbykey]...)
				}
			}
		}

		t.Reset()

		return ret
	}

	return []telegraf.Metric{}
}

func min(a, b int) int {
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
		return NewTopKProcessor()
	})
}


// Here we have the function that generates the aggregation functions
func (t *TopK) get_aggregation_function(agg_operation string) func([]telegraf.Metric, []string) map[string]float64 {

	// This is a function aggregates a set of metrics using a given aggregation function
	var aggregator = func(ms []telegraf.Metric, fields []string, f func(map[string]float64, float64, string)) map[string]float64 {
		agg := make(map[string]float64)
		// Compute the sums of the selected fields over all the measurements collected for this metric
		for _, m := range ms {
			for _, field := range(fields){
				field_val, ok := m.Fields()[field]
				if ! ok {
					continue // Skip if this metric doesn't have this field set
				}
				val, ok := convert(field_val)
				if ! ok {
					panic(fmt.Sprintf("Cannot convert value '%s' from metric '%s' with tags '%s'",
						m.Fields()[field], m.Name(), m.Tags()))
				}
				f(agg, val, field)
			}
		}
		return agg
	}

	switch agg_operation {
	case "sum":
	return func(ms []telegraf.Metric, fields []string) map[string]float64 {
		sum := func(agg map[string]float64, val float64, field string){
			agg[field] += val
		}
		return aggregator(ms, fields, sum)
	}

        case "min":
	return func(ms []telegraf.Metric, fields []string) map[string]float64 {
		min := func(agg map[string]float64, val float64, field string){
			// If this field has not been set, set it to the maximum float64
			_, ok := agg[field]
			if ! ok {
				agg[field] = math.MaxFloat64
			}

			// Check if we've found a new minimum
			if agg[field] > val {
				agg[field] = val
			}
		}
		return aggregator(ms, fields, min)
	}

        case "max":
	return func(ms []telegraf.Metric, fields []string) map[string]float64 {
		max := func(agg map[string]float64, val float64, field string){
			// If this field has not been set, set it to the minimum float64
			_, ok := agg[field]
			if ! ok {
				agg[field] = -math.MaxFloat64
			}

			// Check if we've found a new maximum
			if agg[field] < val {
				agg[field] = val
			}
		}
		return aggregator(ms, fields, max)
	}

	case "avg":
		return func(ms []telegraf.Metric, fields []string) map[string]float64 {
			avg := make(map[string]float64)
			avg_counters := make(map[string]float64)
			// Compute the sums of the selected fields over all the measurements collected for this metric
			for _, m := range ms {
				for _, field := range(fields){
					field_val, ok := m.Fields()[field]
					if ! ok {
						continue // Skip if this metric doesn't have this field set
					}
					val, ok := convert(field_val)
					if ! ok {
						panic(fmt.Sprintf("Cannot convert value '%s' from metric '%s' with tags '%s'",
							m.Fields()[field], m.Name(), m.Tags()))
					}
					avg[field] += val
					avg_counters[field] += 1
				}
			}
			// Divide by the number of recorded measurements collected for every field
			no_measurements_found := true // Canary to check if no field with values was found, so we can return nil
			for k, _ := range(avg){
				if (avg_counters[k] == 0) {
					avg[k] = 0
					continue
				}
				avg[k] = avg[k] / avg_counters[k]
				no_measurements_found = no_measurements_found && false
			}

			if no_measurements_found {
				return nil
			}
			return avg
	}

	default:
		panic(fmt.Sprintf("Unknown aggregation function '%s'", t.Aggregation))
	}
}

