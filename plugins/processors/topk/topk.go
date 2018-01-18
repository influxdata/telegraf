package topk

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TopK struct {
	Period            int
	K                 int
	GroupBy           []string `toml:"group_by"`
	GroupByMetricName bool     `toml:"group_by_metric_name"`
	Fields            []string
	Aggregation       string
	Bottomk           bool
	SimpleTopk        bool   `toml:"simple_topk"`
	DropNoGroup       bool   `toml:"drop_no_group"`
	DropNonTop        bool   `toml:"drop_non_top"`
	GroupByTag        string `toml:"group_by_tag"`
	PositionField     string `toml:"position_field"`
	AggregationField  string `toml:"aggregation_field"`

	cache           map[string][]telegraf.Metric
	metricRegex     *regexp.Regexp
	tagsRegexes     map[string]*regexp.Regexp
	lastAggregation time.Time
}

func NewTopK() TopK {
	// Create object
	topk := TopK{}

	// Setup defaults
	topk.Period = 10
	topk.K = 10
	topk.Fields = []string{"value"}
	topk.Aggregation = "avg"
	topk.GroupBy = []string{}
	topk.GroupByMetricName = false
	topk.GroupByTag = ""
	topk.SimpleTopk = false
	topk.DropNoGroup = true
	topk.DropNonTop = true
	topk.PositionField = ""
	topk.AggregationField = ""

	// Initialize cache
	topk.Reset()

	return topk
}

func NewTopKProcessor() telegraf.Processor {
	topk := NewTopK()
	return &topk
}

var sampleConfig = `
[[processors.topk]]
  period = 10                  # How many seconds between aggregations. Default: 10
  k = 10                       # How many top metrics to return. Default: 10

  # Metrics are grouped based on their tags and name. The plugin aggregates the selected fields of
  # these groups of metrics and sorts the groups based these aggregations
  group_by = ["process_name"]  # Over which tags should the aggregation be done. Default: []
  group_by_metric_name = false # Wheter or not to also group by metric name. Default: false

  # The plugin can aggregate over several fields. If more than one field is specified, an aggregation is calculated per group per field
  # The plugin returns a metric if it is in a group in the top k groups ordered by any of the aggregations of the selected fields
  # This effectively means that more than K metrics may be returned. If you need to return only the top k metrics regardless of grouping, use the simple_topk setting
  fields = ["memory_rss"]      # Over which fields are the top k are calculated. Default: ["value"]
  aggregation = "avg"          # What aggregation to use. Default: "avg". Options: sum, avg, min, max

  bottomk = false              # Instead of the top k largest metrics, return the bottom k lowest metrics. Default: false
  simple_topk = false          # If true, this will override any GroupBy options and assign each metric its own individual group. Default: false
  drop_no_group = true         # Drop any metrics that do fit in any group (due to nonexistent tags). Default: true
  drop_non_top = true          # Drop the metrics that do not make the cut for the top k. Default: true

  group_by_tag = ""            # The plugin assigns each metric a GroupBy tag generated from its name and tags.
                               # If this setting is different than "" the plugin will add a tag (which name will be the value of
                               # this setting) to each metric with the value of the calculated GroupBy tag.
                               # Useful for debugging. Default: ""

  position_field = ""          # This settings provides a way to know the position of each metric in the top k.
                               # If set to a value different than "", then a field (which name will be prefixed with the value of this setting) will be
                               # added to each every metric for each field over which an aggregation was made.
                               # This field will contain the ranking of the group that the metric belonged to. When aggregating
                               # over several fields, several fields will be added (one for each field over which an aggregation was calculated).
                               # Useful for debugging. Default: ""

  aggregation_field = ""       # This setting provies a way know the what values the plugin is generating when aggregating the fields
                               # If set to a value different than "", then a field (which name will be prefixed with the value of this setting) will be added
                               # to each metric which was part of a field aggregation.
                               # The value of the added field will be the value of the result of the aggregation operation for that metric's group. When aggregating
                               # over several fields, several fields will be added (one for each field over which an aggregation was calculated).
                               # Useful for debugging.Default: ""
`

type MetricAggregation struct {
	groupbykey string
	values     map[string]float64
}

type Aggregations struct {
	metrics []MetricAggregation
	field   string
}

func (ags Aggregations) Len() int {
	return len(ags.metrics)
}

func (ags Aggregations) Less(i, j int) bool {
	iv := ags.metrics[i].values[ags.field]
	jv := ags.metrics[j].values[ags.field]
	if iv < jv {
		return true
	} else {
		return false
	}
}

func (ags Aggregations) Swap(i, j int) {
	ags.metrics[i], ags.metrics[j] = ags.metrics[j], ags.metrics[i]
}

func sortMetrics(metrics []MetricAggregation, field string, reverse bool) {
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
	t.lastAggregation = time.Now()
}

func (t *TopK) Description() string {
	return "Print all metrics that pass through this filter."
}

func (t *TopK) generateGroupByKey(m telegraf.Metric) string {
	groupkey := ""

	if t.SimpleTopk {
		return strconv.FormatUint(m.HashID(), 16)
	}

	if t.GroupByMetricName {
		groupkey += m.Name() + "&"
	}
	for _, tag := range t.GroupBy {
		tagValue, ok := m.Tags()[tag]
		if ok {
			groupkey += tag + "=" + tagValue + "&"
		}
	}

	if groupkey == "" && !t.DropNoGroup {
		groupkey = "<<default_groupby_key>>"
	}

	return groupkey
}

func (t *TopK) groupBy(m telegraf.Metric) {
	// Generate the metric group key
	groupkey := t.generateGroupByKey(m)

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
		t.groupBy(m)
	}

	// If enough time has passed
	elapsed := time.Since(t.lastAggregation)
	if elapsed >= time.Second*time.Duration(t.Period) {
		// Generate aggregations list using the selected fields
		aggregations := make([]MetricAggregation, 0, 100)
		var aggregator func([]telegraf.Metric, []string) map[string]float64 = t.getAggregationFunction(t.Aggregation)
		for k, ms := range t.cache {
			aggregations = append(aggregations, MetricAggregation{groupbykey: k, values: aggregator(ms, t.Fields)})
		}

		// Get the top K metrics for each field and add them to the return value
		addedKeys := make(map[string]bool)
		aggField := t.AggregationField
		posField := t.PositionField
		groupTag := t.GroupByTag
		for _, field := range t.Fields {

			// Sort the aggregations
			sortMetrics(aggregations, field, t.Bottomk)

			// Create a one dimentional list with the top K metrics of each key
			for i, ag := range aggregations[0:min(t.K, len(aggregations))] {

				// Check whether of not we need to add fields of tags to the selected metrics
				if aggField != "" || posField != "" || groupTag != "" {
					for _, m := range t.cache[ag.groupbykey] {
						if aggField != "" && m.HasField(field) {
							m.AddField(aggField+"_"+field, ag.values[field])
						}
						if posField != "" {
							m.AddField(posField+"_"+field, i+1) //+1 to it starts from 1
						}
						if groupTag != "" {
							m.AddTag(groupTag, ag.groupbykey) //+1 to it starts from 1
						}
					}
				}

				// Add metrics if we have not already appended them to the return value
				_, ok := addedKeys[ag.groupbykey]
				if !ok {
					ret = append(ret, t.cache[ag.groupbykey]...)
					addedKeys[ag.groupbykey] = true
				}
			}
		}

		//Lastly, if we were instructed to not drop the bottom metrics, append them as is to the output
		if !t.DropNonTop {
			for _, ag := range aggregations {
				_, ok := addedKeys[ag.groupbykey]
				if !ok {
					ret = append(ret, t.cache[ag.groupbykey]...)
					addedKeys[ag.groupbykey] = true
				}
			}
		}

		t.Reset()

		return ret
	}

	return []telegraf.Metric{}
}

func min(a, b int) int {
	if a > b {
		return b
	}
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
func (t *TopK) getAggregationFunction(aggOperation string) func([]telegraf.Metric, []string) map[string]float64 {

	// This is a function aggregates a set of metrics using a given aggregation function
	var aggregator = func(ms []telegraf.Metric, fields []string, f func(map[string]float64, float64, string)) map[string]float64 {
		agg := make(map[string]float64)
		// Compute the sums of the selected fields over all the measurements collected for this metric
		for _, m := range ms {
			for _, field := range fields {
				fieldVal, ok := m.Fields()[field]
				if !ok {
					continue // Skip if this metric doesn't have this field set
				}
				val, ok := convert(fieldVal)
				if !ok {
					panic(fmt.Sprintf("Cannot convert value '%s' from metric '%s' with tags '%s'",
						m.Fields()[field], m.Name(), m.Tags()))
				}
				f(agg, val, field)
			}
		}
		return agg
	}

	switch aggOperation {
	case "sum":
		return func(ms []telegraf.Metric, fields []string) map[string]float64 {
			sum := func(agg map[string]float64, val float64, field string) {
				agg[field] += val
			}
			return aggregator(ms, fields, sum)
		}

	case "min":
		return func(ms []telegraf.Metric, fields []string) map[string]float64 {
			min := func(agg map[string]float64, val float64, field string) {
				// If this field has not been set, set it to the maximum float64
				_, ok := agg[field]
				if !ok {
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
			max := func(agg map[string]float64, val float64, field string) {
				// If this field has not been set, set it to the minimum float64
				_, ok := agg[field]
				if !ok {
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
			avgCounters := make(map[string]float64)
			// Compute the sums of the selected fields over all the measurements collected for this metric
			for _, m := range ms {
				for _, field := range fields {
					fieldVal, ok := m.Fields()[field]
					if !ok {
						continue // Skip if this metric doesn't have this field set
					}
					val, ok := convert(fieldVal)
					if !ok {
						panic(fmt.Sprintf("Cannot convert value '%s' from metric '%s' with tags '%s'",
							m.Fields()[field], m.Name(), m.Tags()))
					}
					avg[field] += val
					avgCounters[field] += 1
				}
			}
			// Divide by the number of recorded measurements collected for every field
			noMeasurementsFound := true // Canary to check if no field with values was found, so we can return nil
			for k, _ := range avg {
				if avgCounters[k] == 0 {
					avg[k] = 0
					continue
				}
				avg[k] = avg[k] / avgCounters[k]
				noMeasurementsFound = noMeasurementsFound && false
			}

			if noMeasurementsFound {
				return nil
			}
			return avg
		}

	default:
		panic(fmt.Sprintf("Unknown aggregation function '%s'", t.Aggregation))
	}
}
