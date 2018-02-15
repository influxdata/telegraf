package topk

import (
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
)

// Key, value pair that represents a telegraf.Metric field
type field struct {
	key string
	val interface{}
}

func fieldList(fields ...field) []field {
	return fields
}

// Key, value pair that represents a telegraf.Metric tags
type tag struct {
	key string
	val string
}

func tagList(tags ...tag) []tag {
	return tags
}

// Abstraction of a change in a single metric
type metricChange struct {
	newFields []field // Fieldsthat should be added to the metric
	newTags   []tag   // Tags that should be added to the metric
	runHash   bool    // Sometimes the metrics' HashID must be run so the deep comparison works
}

// Generate a new set of metrics from a set of changes. This is used to generate an answer which will be
// compare against the output of the processor
// NOTE: A `changeSet` is a map where the keys are the indices of the metrics to keep, and the values
//       are list of new tags and fields to be added to the metric in that index.
//       THE ORDERING OF THE NEW TAGS AND FIELDS MATTERS. When using reflect.DeepEqual to compare metrics,
//       comparing metrics that have the same fields/tags added in different orders will return false, although
//       they are semantically equal.
//       Therefore the fields and tags must be in the same order that the processor would add them
func generateAns(input []telegraf.Metric, changeSet map[int]metricChange) []telegraf.Metric {
	answer := []telegraf.Metric{}

	// For every input metric, we check if there is a change we need to apply
	// If there is no change for a given input metric, the metric is dropped
	for i, metric := range input {
		change, ok := changeSet[i]
		if ok {
			// Deep copy the metric
			newMetric := metric.Copy()

			// Add new fields
			if change.newFields != nil {
				for _, p := range change.newFields {
					newMetric.AddField(p.key, p.val)
				}
			}

			// Add new tags
			if change.newTags != nil {
				for _, p := range change.newTags {
					newMetric.AddTag(p.key, p.val)
				}
			}

			// Run the hash function if required
			if change.runHash {
				newMetric.HashID()
			}

			answer = append(answer, newMetric)
		}
	}

	return answer
}

func deepCopy(a []telegraf.Metric) []telegraf.Metric {
	ret := make([]telegraf.Metric, 0, len(a))
	for _, m := range a {
		ret = append(ret, m.Copy())
	}

	return ret
}

func belongs(m telegraf.Metric, ms []telegraf.Metric) bool {
	for _, i := range ms {
		if reflect.DeepEqual(i, m) {
			return true
		}
	}
	return false
}

func subSet(a []telegraf.Metric, b []telegraf.Metric) bool {
	subset := true
	for _, m := range a {
		if !belongs(m, b) {
			subset = false
			break
		}
	}
	return subset
}

func equalSets(l1 []telegraf.Metric, l2 []telegraf.Metric) bool {
	return subSet(l1, l2) && subSet(l2, l1)
}

func runAndCompare(topk *TopK, metrics []telegraf.Metric, answer []telegraf.Metric, testID string, t *testing.T) {
	// Sleep for `period`, otherwise the processor will only
	// cache the metrics, but it will not process them
	period := time.Second * time.Duration(topk.Period)
	time.Sleep(period)

	// Run the processor
	ret := topk.Apply(metrics...)
	topk.Reset()

	// The returned set mut be equal to the answer set
	if !equalSets(ret, answer) {
		t.Error("\nExpected metrics for", testID, ":\n",
			answer, "\nReturned metrics:\n", ret)
	}
}

// Smoke tests
func TestTopkAggregatorsSmokeTests(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	aggregators := []string{"mean", "sum", "max", "min"}

	//The answer is equal to the original set for these particual scenarios
	input := MetricsSet1
	answer := MetricsSet1

	for _, ag := range aggregators {
		topk.Aggregation = ag

		runAndCompare(&topk, input, answer, "SmokeAggregator_"+ag, t)
	}
}

// AddAggregateField + Mean aggregator
func TestTopkMeanAddAggregateField(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.Aggregation = "mean"
	topk.AggregateFieldSuffix = "_meanag"
	topk.AddAggregateField = []string{"a"}
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet1)

	// Generate the answer
	chng := fieldList(field{"a_meanag", float64(28.044)})
	changeSet := map[int]metricChange{
		0: metricChange{newFields: chng},
		1: metricChange{newFields: chng},
		2: metricChange{newFields: chng},
		3: metricChange{newFields: chng},
		4: metricChange{newFields: chng},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "MeanAddAggregateField test", t)
}

// AddAggregateField + Sum aggregator
func TestTopkSumAddAggregateField(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.Aggregation = "sum"
	topk.AggregateFieldSuffix = "_sumag"
	topk.AddAggregateField = []string{"a"}
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet1)

	// Generate the answer
	chng := fieldList(field{"a_sumag", float64(140.22)})
	changeSet := map[int]metricChange{
		0: metricChange{newFields: chng},
		1: metricChange{newFields: chng},
		2: metricChange{newFields: chng},
		3: metricChange{newFields: chng},
		4: metricChange{newFields: chng},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "SumAddAggregateField test", t)
}

// AddAggregateField + Max aggregator
func TestTopkMaxAddAggregateField(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.Aggregation = "max"
	topk.AggregateFieldSuffix = "_maxag"
	topk.AddAggregateField = []string{"a"}
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet1)

	// Generate the answer
	chng := fieldList(field{"a_maxag", float64(50.5)})
	changeSet := map[int]metricChange{
		0: metricChange{newFields: chng},
		1: metricChange{newFields: chng},
		2: metricChange{newFields: chng},
		3: metricChange{newFields: chng},
		4: metricChange{newFields: chng},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "MaxAddAggregateField test", t)
}

// AddAggregateField + Min aggregator
func TestTopkMinAddAggregateField(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.Aggregation = "min"
	topk.AggregateFieldSuffix = "_minag"
	topk.AddAggregateField = []string{"a"}
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet1)

	// Generate the answer
	chng := fieldList(field{"a_minag", float64(0.3)})
	changeSet := map[int]metricChange{
		0: metricChange{newFields: chng},
		1: metricChange{newFields: chng},
		2: metricChange{newFields: chng},
		3: metricChange{newFields: chng},
		4: metricChange{newFields: chng},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "MinAddAggregateField test", t)
}

// GroupBy
func TestTopkGroupby1(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 3
	topk.Aggregation = "sum"
	topk.AggregateFieldSuffix = "_sumag"
	topk.AddAggregateField = []string{"value"}
	topk.GroupBy = []string{"tag[13]"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		2: metricChange{newFields: fieldList(field{"value_sumag", float64(74.18)})},
		3: metricChange{newFields: fieldList(field{"value_sumag", float64(72)})},
		4: metricChange{newFields: fieldList(field{"value_sumag", float64(163.22)})},
		5: metricChange{newFields: fieldList(field{"value_sumag", float64(163.22)})},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "GroupBy test 1", t)
}
func TestTopkGroupby2(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 1
	topk.Aggregation = "mean"
	topk.AggregateFieldSuffix = "_mean"
	topk.AddAggregateField = []string{"value"}
	topk.GroupBy = []string{"tag1"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	chng := fieldList(field{"value_mean", float64(74.20750000000001)})
	changeSet := map[int]metricChange{
		1: metricChange{newFields: chng},
		2: metricChange{newFields: chng},
		4: metricChange{newFields: chng},
		5: metricChange{newFields: chng},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "GroupBy test 2", t)
}
func TestTopkGroupby3(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 1
	topk.Aggregation = "min"
	topk.AggregateFieldSuffix = "_minaggfield"
	topk.AddAggregateField = []string{"value"}
	topk.GroupBy = []string{"tag4"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	chng := fieldList(field{"value_minaggfield", float64(75.3)})
	changeSet := map[int]metricChange{
		4: metricChange{newFields: chng},
		5: metricChange{newFields: chng},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "GroupBy test 3", t)
}
func TestTopkGroupby4(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 1
	topk.Aggregation = "min"
	topk.GroupBy = []string{"tag9"} // This is a nonexistent tag in this test set
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	answer := []telegraf.Metric{} // This test should drop all metrics

	// Run the test
	runAndCompare(&topk, input, answer, "GroupBy test 4", t)
}

// GroupBy + Fields
func TestTopkGroupbyFields1(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 2
	topk.Aggregation = "mean"
	topk.AggregateFieldSuffix = "_mean"
	topk.AddAggregateField = []string{"A"}
	topk.GroupBy = []string{"tag1", "tag2"}
	topk.Fields = []string{"A"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		0: metricChange{newFields: fieldList(field{"A_mean", float64(95.36)})},
		1: metricChange{newFields: fieldList(field{"A_mean", float64(39.01)})},
		2: metricChange{newFields: fieldList(field{"A_mean", float64(39.01)})},
		3: metricChange{},
		4: metricChange{},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "GroupBy Fields test 1", t)
}

func TestTopkGroupbyFields2(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 2
	topk.Aggregation = "sum"
	topk.AggregateFieldSuffix = "_sum"
	topk.AddAggregateField = []string{"B", "C"}
	topk.GroupBy = []string{"tag1", "tag3"}
	topk.Fields = []string{"B", "C"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		0: metricChange{newFields: fieldList(field{"C_sum", float64(72.41)})},
		2: metricChange{newFields: fieldList(field{"B_sum", float64(60.96)})},
		4: metricChange{newFields: fieldList(field{"B_sum", float64(81.55)}, field{"C_sum", float64(49.96)})},
		5: metricChange{newFields: fieldList(field{"C_sum", float64(49.96)})},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "GroupBy Fields test 2", t)
}

// GroupBy metric name
func TestTopkGroupbyMetricName1(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 1
	topk.Aggregation = "sum"
	topk.AggregateFieldSuffix = "_sigma"
	topk.AddAggregateField = []string{"value"}
	topk.GroupBy = []string{}

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	chng := fieldList(field{"value_sigma", float64(235.22000000000003)})
	changeSet := map[int]metricChange{
		3: metricChange{newFields: chng},
		4: metricChange{newFields: chng},
		5: metricChange{newFields: chng},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "GroupBy by metric name test 1", t)
}

func TestTopkGroupbyMetricName2(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 2
	topk.Aggregation = "sum"
	topk.AggregateFieldSuffix = "_SUM"
	topk.AddAggregateField = []string{"A", "value"}
	topk.GroupBy = []string{"tag[12]"}
	topk.Fields = []string{"A", "value"}

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		0: metricChange{newFields: fieldList(field{"A_SUM", float64(95.36)})},
		1: metricChange{newFields: fieldList(field{"A_SUM", float64(78.02)}, field{"value_SUM", float64(133.61)})},
		2: metricChange{newFields: fieldList(field{"A_SUM", float64(78.02)}, field{"value_SUM", float64(133.61)})},
		4: metricChange{newFields: fieldList(field{"value_SUM", float64(87.92)})},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "GroupBy by metric name test 2", t)
}

// DropNoGroup
func TestTopkDropNoGroupFalse(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 1
	topk.Aggregation = "sum"
	topk.GroupBy = []string{"tag5"}
	topk.DropNoGroup = false
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		0: metricChange{},
		1: metricChange{},
		3: metricChange{},
		4: metricChange{},
		5: metricChange{},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "DropNoGroup False test", t)
}

// DropNonTop=false + RankField
func TestTopkDontDropBottom(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 3
	topk.Aggregation = "sum"
	topk.AggregateFieldSuffix = "_sumag"
	topk.AddAggregateField = []string{"value"}
	topk.GroupBy = []string{"tag1", "tag3"}
	topk.DropNonTop = false
	topk.RankFieldSuffix = "_aggpos"
	topk.AddRankField = []string{"value"}
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		0: metricChange{},
		1: metricChange{},
		2: metricChange{newFields: fieldList(field{"value_sumag", float64(74.18)}, field{"value_aggpos", 2})},
		3: metricChange{newFields: fieldList(field{"value_sumag", float64(72)}, field{"value_aggpos", 3})},
		4: metricChange{newFields: fieldList(field{"value_sumag", float64(163.22)}, field{"value_aggpos", 1})},
		5: metricChange{newFields: fieldList(field{"value_sumag", float64(163.22)}, field{"value_aggpos", 1})},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "DontDropBottom test", t)
}

// BottomK
func TestTopkBottomk(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 3
	topk.Aggregation = "sum"
	topk.GroupBy = []string{"tag1", "tag3"}
	topk.Bottomk = true
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		0: metricChange{},
		1: metricChange{},
		3: metricChange{},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "Bottom k test", t)
}

// GroupByKeyTag
func TestTopkGroupByKeyTag(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 3
	topk.Aggregation = "sum"
	topk.GroupBy = []string{"tag1", "tag3"}
	topk.AddGroupByTag = "gbt"
	topk.DropNonTop = false
	topk.DropNoGroup = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		0: metricChange{},
		1: metricChange{},
		2: metricChange{newTags: tagList(tag{"gbt", "metric1&tag1=TWO&tag3=SIX&"})},
		3: metricChange{newTags: tagList(tag{"gbt", "metric2&tag1=ONE&tag3=THREE&"})},
		4: metricChange{newTags: tagList(tag{"gbt", "metric2&tag1=TWO&tag3=SEVEN&"})},
		5: metricChange{newTags: tagList(tag{"gbt", "metric2&tag1=TWO&tag3=SEVEN&"})},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "GroupByKeyTag test", t)
}

// No drops
func TestTopkNodrops1(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 3
	topk.Aggregation = "sum"
	topk.AddAggregateField = []string{"value"}
	topk.AddRankField = []string{"value"}
	topk.GroupBy = []string{"tag1", "tag3"}
	topk.DropNonTop = false
	topk.DropNoGroup = false
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		0: metricChange{},
		1: metricChange{},
		2: metricChange{newFields: fieldList(field{"value_aggregate", float64(74.18)}, field{"value_rank", 2})},
		3: metricChange{newFields: fieldList(field{"value_aggregate", float64(72)}, field{"value_rank", 3})},
		4: metricChange{newFields: fieldList(field{"value_aggregate", float64(163.22)}, field{"value_rank", 1})},
		5: metricChange{newFields: fieldList(field{"value_aggregate", float64(163.22)}, field{"value_rank", 1})},
	}
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "NoDrops test 1", t)
}

func TestTopkNodrops2(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 3
	topk.Aggregation = "sum"
	topk.GroupBy = []string{"tag1", "tag3"}
	topk.DropNonTop = false
	topk.DropNoGroup = false
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	answer := deepCopy(MetricsSet2)

	// Run the test
	runAndCompare(&topk, input, answer, "NoDrops test 2", t)
}

// Simple topk
func TestTopkSimpleTopk(t *testing.T) {

	// Build the processor
	var topk TopK
	topk = *New()
	topk.Period = 1
	topk.K = 3
	topk.Aggregation = "sum"
	topk.SimpleTopk = true
	topk.GroupByMetricName = false

	// Get the input
	input := deepCopy(MetricsSet2)

	// Generate the answer
	changeSet := map[int]metricChange{
		2: metricChange{runHash: true},
		4: metricChange{runHash: true},
		5: metricChange{runHash: true},
	}
	// Generate the answer
	answer := generateAns(input, changeSet)

	// Run the test
	runAndCompare(&topk, input, answer, "SimpleTopk test", t)
}
