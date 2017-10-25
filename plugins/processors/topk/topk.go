package topk

import (
	"sort"
	"time"
	"regexp"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TopK struct {
	Metric             string
	Period             int
	K                  int
	Fields             []string
	Tags               map[string]string
	Aggregation        string
	GroupBy            []string `toml:"group_by"`
	GroupByMetricName  bool `toml:"group_by_metric_name"`
        Bottomk            bool
        RevertMetricMatch  bool `toml:"revert_metric_match"`
        RevertTagMatch     bool `toml:"revert_tag_match"`
        DropNonMatching    bool `toml:"drop_non_matching"`
	DropNonTop         bool `toml:"drop_non_top"`
	PositionField      string `toml:"position_field"`
	AggregationField   string `toml:"aggregation_field"`

	cache map[string][]telegraf.Metric
	metric_regex *regexp.Regexp
	tags_regexes map[string]*regexp.Regexp
	last_aggregation time.Time
}

func NewTopK() telegraf.Processor{
	// Create object
	topk := &TopK{}

	// Setup defaults
	topk.Metric = ".*"
	topk.Period = 10
	topk.K = 10
	topk.Fields = nil
	topk.Tags = nil
	topk.Aggregation = "avg"
	topk.GroupBy = nil
	topk.GroupByMetricName = false
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
		sort.Reverse(aggs)
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

func (t *TopK) init_regexes() {
	// Compile regex for the metric name
	if (t.metric_regex == nil) {
		var err error
		t.metric_regex, err = regexp.Compile(t.Metric)
		if (err != nil) {
			panic(fmt.Sprintf("TopK processor could not parse metric name regex '%s'", t.Metric))
		}
	}

	// Compile regexes for the tags
	if (t.Tags != nil) && (t.tags_regexes == nil) {
		t.tags_regexes = make(map[string]*regexp.Regexp)
		for key, regex := range t.Tags {
			regex, err := regexp.Compile(regex)
			if (err != nil) {
				panic(fmt.Sprintf("TopK processor could not parse tag regex '%s'", t.Metric))
			}
			t.tags_regexes[key] = regex
		}
	}
}

func (t *TopK) match_metric(m telegraf.Metric) bool {
	// Run metric name against our metric regex
	match_name := t.metric_regex.MatchString(m.Name())
	if ! (match_name != t.RevertMetricMatch) { return false }

	// Run every tag against our tags regexes
	if t.Tags == nil { return true }
	match_tags := false
	for key, value := range m.Tags() {
		if _, ok := t.tags_regexes[key]; ok {
			match_tags = t.tags_regexes[key].MatchString(value) && (match_tags || ok)
		}
	}
	if ! (match_tags != t.RevertTagMatch) { return false }

	return true
}

func (t *TopK) generate_groupby_key(m telegraf.Metric) string {
	groupkey := ""
	if t.GroupByMetricName {
		groupkey += m.Name() + "&"
	}
	for _, tag := range(t.GroupBy) {
		groupkey += tag + "=" + m.Tags()[tag] + "&"
	}

	if groupkey == "" {
		groupkey = "<<default_groupby_key>>"
	}

	return groupkey
}

func (t *TopK) group_by(m telegraf.Metric) {
	// Generate the metric group key
	groupkey := t.generate_groupby_key(m)

	// Initialize the key with an empty list if necessary
	if _, ok := t.cache[groupkey]; !ok {
		t.cache[groupkey] = make([]telegraf.Metric, 0, 10)
	}

	// Append the metric to the corresponding key list
	t.cache[groupkey] = append(t.cache[groupkey], m)
}

func (t *TopK) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// Generate the regexp structs that we use to match the metrics
	t.init_regexes()

	// Add the metrics received to our internal cache
	var ret []telegraf.Metric = nil
	for _, m := range in {
		if (t.match_metric(m)){
			t.group_by(m)
		} else {
			// If the metric didn't match, add it to the return value, so we don't drop it
			if (ret == nil) {
				ret = make([]telegraf.Metric, 0, len(in))
			}
			ret = append(ret, m)
		}
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
		for _, field := range(t.Fields) {
			// Sort the aggregations
			sort_metrics(aggregations, field, t.Bottomk)

			// Create a one dimentional list with the top K metrics of each key
			added_keys := make(map[string]bool)
			for _, ag := range aggregations[0:min(t.K, len(aggregations))] {
				_, ok := added_keys[ag.groupbykey]
				if ! ok { // Check that we haven't already added these metrics
					ret = append(ret, t.cache[ag.groupbykey]...)
					added_keys[ag.groupbykey] = true
				}
			}
		}

		t.Reset()

		return ret
	}

	return ret
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


// Here we have the function that generates the aggregation functions
func (t *TopK) get_aggregation_function(agg_operation string) func([]telegraf.Metric, []string) map[string]float64 {
	switch agg_operation {
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
						fmt.Println(m)
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
					avg[k] = 0 // FIX. We have no way of knowing if a bucket was ever touched
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
