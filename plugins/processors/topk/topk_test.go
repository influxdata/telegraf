package topk

import (
	"testing"
	"time"
	"reflect"

	"github.com/influxdata/telegraf"
)

func deepCopy(a []telegraf.Metric) []telegraf.Metric {
	ret := make([]telegraf.Metric, 0, len(a))
	for _, m := range(a) {
		ret = append(ret, m.Copy())
	}

	return ret
}

func belongs(m telegraf.Metric, ms []telegraf.Metric) bool {
	for _, i := range(ms){
		if reflect.DeepEqual(i, m) {
			return true
		}
	}
	return false
}

func subSet(a []telegraf.Metric, b []telegraf.Metric) bool {
	subset := true
	for _, m := range(a){
		if ! belongs(m, b) {
			subset = false
			break
		}
	}
	return subset
}

func equalSets(l1 []telegraf.Metric, l2 []telegraf.Metric) bool {
	return subSet(l1, l2) && subSet(l2, l1)
}

func runAndCompare(topk *TopK, metrics []telegraf.Metric, answer []telegraf.Metric, test_id string, t *testing.T) {
	// Sleep for `period`, otherwise the processor will only
	// cache the metrics, but it will not process them
	period := time.Second * time.Duration(topk.Period)
	time.Sleep(period)

	// Run the processor
	ret := topk.Apply(metrics...)
	topk.Reset()

	// The returned set mut be equal to the answer set
	if ! equalSets(ret, answer) {
		t.Error("\nExpected metrics for", test_id, ":\n",
			answer, "\nReturned metrics:\n", ret)
	}
}

// Smoke tests
func TestTopkSmokeTest(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	aggregators := []string{"avg", "sum", "max", "min"}

	for _,ag := range(aggregators) {
		topk.Aggregation = ag

		//The answer is equal to the original set for these particual scenarios
		runAndCompare(&topk, MetricsSet1, MetricsSet1, "SmokeAggregator_"+ag, t)
	}
}

// AggregationField + Avg aggregator
func TestTopkAvgAggregationField(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Aggregation = "avg"
	topk.AggregationField = "avgag"
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	runAndCompare(&topk, deepCopy(MetricsSet1), AvgAggregationFieldAns, "AvgAggregationField test", t)
}

// AggregationField + Sum aggregator
func TestTopkSumAggregationField(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Aggregation = "sum"
	topk.AggregationField = "sumag"
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	runAndCompare(&topk, deepCopy(MetricsSet1), SumAggregationFieldAns, "SumAggregationField test", t)
}

// AggregationField + Max aggregator
func TestTopkMaxAggregationField(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Aggregation = "max"
	topk.AggregationField = "maxag"
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	runAndCompare(&topk, deepCopy(MetricsSet1), MaxAggregationFieldAns, "MaxAggregationField test", t)
}

// AggregationField + Min aggregator
func TestTopkMinAggregationField(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Aggregation = "min"
	topk.AggregationField = "minag"
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	runAndCompare(&topk, deepCopy(MetricsSet1), MinAggregationFieldAns, "MinAggregationField test", t)
}

// GroupBy
// GroupBy + Fields
// GroupBy metric name
// GroupBy + GroupBy metric name
// DropNoGroup
// DropNonTop=false + PositionField
// BottomK
