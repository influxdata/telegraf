package topk

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/internal"
)

type TopK struct {
	Period               internal.Duration
	K                    int
	GroupBy              []string `toml:"group_by"`
	Fields               []string
	Aggregation          string
	Bottomk              bool
	DropNonTop           bool     `toml:"drop_non_top"`
	AddGroupByTag        string   `toml:"add_groupby_tag"`
	AddRankFields        []string `toml:"add_rank_fields"`
	RankFieldSuffix      string   `toml:"rank_field_suffix"`
	AddAggregateFields   []string `toml:"add_aggregate_fields"`
	AggregateFieldSuffix string   `toml:"aggregate_field_suffix"`

	cache           map[string][]telegraf.Metric
	tagsGlobs       filter.Filter
	rankFieldSet    map[string]bool
	aggFieldSet     map[string]bool
	lastAggregation time.Time
}

func New() *TopK {
	// Create object
	topk := TopK{}

	// Setup defaults
	topk.Period = internal.Duration{Duration: time.Second*time.Duration(10)}
	topk.K = 10
	topk.Fields = []string{"value"}
	topk.Aggregation = "mean"
	topk.GroupBy = []string{"*"}
	topk.AddGroupByTag = ""
	topk.DropNonTop = true
	topk.AddRankFields = []string{""}
	topk.RankFieldSuffix = "_rank"
	topk.AddAggregateFields = []string{""}
	topk.AggregateFieldSuffix = "_aggregate"

	// Initialize cache
	topk.Reset()

	return &topk
}

var sampleConfig = `
  ## How many seconds between aggregations
  # period = 10

  ## How many top metrics to return
  # k = 10

  ## Over which tags should the aggregation be done. Globs can be specified, in
  ## which case any tag matching the glob will aggregated over. If set to an
  ## empty list is no aggregation over tags is done
  # group_by = ['*']

  ## Over which fields are the top k are calculated
  # fields = ["value"]

  ## What aggregation to use. Options: sum, mean, min, max
  # aggregation = "mean"

  ## Instead of the top k largest metrics, return the bottom k lowest metrics
  # bottomk = false

  ## Drop the metrics that do not make the cut for the top k
  # drop_non_top = true          

  ## The plugin assigns each metric a GroupBy tag generated from its name and
  ## tags. If this setting is different than "" the plugin will add a
  ## tag (which name will be the value of this setting) to each metric with
  ## the value of the calculated GroupBy tag. Useful for debugging
  # add_groupby_tag = ""          

  ## These settings provide a way to know the position of each metric in
  ## the top k. The 'add_rank_field' setting allows to specify for which
  ## fields the position is required. If the list is non empty, then a field
  ## will be added to each every metric for each field present in the 
  ## 'add_rank_field'. This field will contain the ranking of the group that
  ## the metric belonged to when aggregated over that field.
  ## The name of the field will be set to the name of the aggregation field,
  ## suffixed by the value of the 'rank_field_suffix' setting
  # add_rank_fields = []
  # rank_field_suffix = "_rank"

  ## These settings provide a way to know what values the plugin is generating
  ## when aggregating metrics. The 'add_agregate_field' setting allows to
  ## specify for which fields the final aggregation value is required. If the
  ## list is non empty, then a field will be added to each every metric for
  ## each field present in the 'add_aggregate_field'. This field will contain
  ## the computed aggregation for the group that the metric belonged to when
  ## aggregated over that field.
  ## The name of the field will be set to the name of the aggregation field,
  ## suffixed by the value of the 'aggregate_field_suffix' setting
  # add_aggregate_fields = []
  # aggregate_field_suffix = "_aggregate"
`

type MetricAggregation struct {
	groupbykey string
	values     map[string]float64
}

func sortMetrics(metrics []MetricAggregation, field string, reverse bool) {
	less := func(i, j int) bool {
		iv := metrics[i].values[field]
		jv := metrics[j].values[field]
		if iv < jv {
			return true
		} else {
			return false
		}
	}

	if reverse {
		sort.SliceStable(metrics, less)
	} else {
		sort.SliceStable(metrics, func(i, j int) bool { return !less(i, j) })
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

func (t *TopK) generateGroupByKey(m telegraf.Metric) (string, error) {
	// Create the filter.Filter objects if they have not been created
	if t.tagsGlobs == nil && len(t.GroupBy) > 0 {
		var err error
		t.tagsGlobs, err = filter.Compile(t.GroupBy)
		if err != nil {
			return "", fmt.Errorf("could not compile pattern: %v %v", t.GroupBy, err)
		}
	}

	groupkey := m.Name() + "&"

	if len(t.GroupBy) > 0 {
		tags := m.Tags()
		keys := make([]string, 0, len(tags))
		for tag, value := range tags {
			if t.tagsGlobs.Match(tag) {
				keys = append(keys, tag+"="+value+"&")
			}
		}
		// Sorting the selected tags is necessary because dictionaries
		// do not ensure any specific or deterministic ordering
		sort.SliceStable(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, str := range keys {
			groupkey += str
		}
	}

	return groupkey, nil
}

func (t *TopK) groupBy(m telegraf.Metric) {
	// Generate the metric group key
	groupkey, err := t.generateGroupByKey(m)
	if err != nil {
		// If we could not generate the groupkey, fail hard
		// by dropping this and all subsequent metrics
		log.Print("E! [processors.topk]: could not generate group key: %v", err)
		return
	}

	// Initialize the key with an empty list if necessary
	if _, ok := t.cache[groupkey]; !ok {
		t.cache[groupkey] = make([]telegraf.Metric, 0, 10)
	}

	// Append the metric to the corresponding key list
	t.cache[groupkey] = append(t.cache[groupkey], m)

	// Add the generated groupby key tag to the metric if requested
	if t.AddGroupTag != "" {
		m.AddTag(t.AddGroupTag, groupkey)
	}
}

func (t *TopK) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// Init any internal datastructures that are not initialized yet
	if t.rankFieldSet == nil {
		t.rankFieldSet = make(map[string]bool)
		for _, f := range t.AddRankFields {
			t.rankFieldSet[f] = true
		}
	}
	if t.aggFieldSet == nil {
		t.aggFieldSet = make(map[string]bool)
		for _, f := range t.AddAggregateFields {
			t.aggFieldSet[f] = true
		}
	}

	// Add the metrics received to our internal cache
	for _, m := range in {
		t.groupBy(m)
	}

	// If enough time has passed
	elapsed := time.Since(t.lastAggregation)
	if elapsed >= t.Period.Duration {
		return push()
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
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func (t *TopK) push() []telegraf.Metric {
	// Generate aggregations list using the selected fields
	aggregations := make([]MetricAggregation, 0, 100)
	aggregator, err := t.getAggregationFunction(t.Aggregation)
	if err != nil {
		// If we could not generate the aggregation
		// function, fail hard by dropping all metrics
		log.Print("E! [processors.topk]: %v", err)
		return []telegraf.Metric{}
	}
	for k, ms := range t.cache {
		aggregations = append(aggregations, MetricAggregation{groupbykey: k, values: aggregator(ms, t.Fields)})
	}

	// The return value that will hold the returned metrics
	var ret []telegraf.Metric = make([]telegraf.Metric, 0, 0)

	// Get the top K metrics for each field and add them to the return value
	addedKeys := make(map[string]bool)
	aggFieldSuffix := t.AggregateFieldSuffix
	rankFieldSuffix := t.RankFieldSuffix
	groupTag := t.AddGroupByTag
	for _, field := range t.Fields {

		// Sort the aggregations
		sortMetrics(aggregations, field, t.Bottomk)

		// Create a one dimentional list with the top K metrics of each key
		for i, ag := range aggregations[0:min(t.K, len(aggregations))] {

			// Check whether of not we need to add fields of tags to the selected metrics
			if len(t.aggFieldSet) != 0 || len(t.rankFieldSet) != 0 || groupTag != "" {
				for _, m := range t.cache[ag.groupbykey] {

					// Add the aggregation final value if requested
					_, addAggField := t.aggFieldSet[field]
					if addAggField && m.HasField(field) {
						m.AddField(field+aggFieldSuffix, ag.values[field])
					}

					// Add the rank relative to the current field if requested
					_, addRankField := t.rankFieldSet[field]
					if addRankField && m.HasField(field) {
						m.AddField(field+rankFieldSuffix, i+1)
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


// Function that generates the aggregation functions
func (t *TopK) getAggregationFunction(aggOperation string) (func([]telegraf.Metric, []string) map[string]float64, error) {

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
					log.Printf("Cannot convert value '%s' from metric '%s' with tags '%s'",
						m.Fields()[field], m.Name(), m.Tags())
					continue
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
		}, nil

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
		}, nil

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
		}, nil

	case "mean":
		return func(ms []telegraf.Metric, fields []string) map[string]float64 {
			mean := make(map[string]float64)
			meanCounters := make(map[string]float64)
			// Compute the sums of the selected fields over all the measurements collected for this metric
			for _, m := range ms {
				for _, field := range fields {
					fieldVal, ok := m.Fields()[field]
					if !ok {
						continue // Skip if this metric doesn't have this field set
					}
					val, ok := convert(fieldVal)
					if !ok {
						log.Printf("Cannot convert value '%s' from metric '%s' with tags '%s'",
							m.Fields()[field], m.Name(), m.Tags())
						continue
					}
					mean[field] += val
					meanCounters[field] += 1
				}
			}
			// Divide by the number of recorded measurements collected for every field
			noMeasurementsFound := true // Canary to check if no field with values was found, so we can return nil
			for k, _ := range mean {
				if meanCounters[k] == 0 {
					mean[k] = 0
					continue
				}
				mean[k] = mean[k] / meanCounters[k]
				noMeasurementsFound = noMeasurementsFound && false
			}

			if noMeasurementsFound {
				return nil
			}
			return mean
		}, nil

	default:
		return nil, fmt.Errorf("Unknown aggregation function '%s'. No metrics will be processed", t.Aggregation)
	}
}

func init() {
	processors.Add("topk", func() telegraf.Processor {
		return New()
	})
}
