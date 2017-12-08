package topk

import (
	"testing"
	"time"
	"reflect"

	"github.com/influxdata/telegraf"
)

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

// Smoke avg
func TestTopkAvgSmokeTest(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	period := time.Second * time.Duration(topk.Period)
	time.Sleep(period)

	ret := topk.Apply(MetricsSet1...)
	answer := MetricsSet1

	if ! equalSets(ret, answer) {
		t.Error("\nExpected metrics:\n", answer, "\nReturned metrics:\n", ret)
	}
}

// Smoke sum
func TestTopkSumSmokeTest(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Aggregation = "sum"
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	period := time.Second * time.Duration(topk.Period)
	time.Sleep(period)

	ret := topk.Apply(MetricsSet1...)
	answer := MetricsSet1

	if ! equalSets(ret, answer) {
		t.Error("\nExpected metrics:\n", answer, "\nReturned metrics:\n", ret)
	}
}

// Smoke max
func TestTopkMaxSmokeTest(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Aggregation = "max"
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	period := time.Second * time.Duration(topk.Period)
	time.Sleep(period)

	ret := topk.Apply(MetricsSet1...)
	answer := MetricsSet1

	if ! equalSets(ret, answer) {
		t.Error("\nExpected metrics:\n", answer, "\nReturned metrics:\n", ret)
	}
}

// Smoke min
func TestTopkMinSmokeTest(t *testing.T) {
	var topk TopK
	topk = NewTopK()
	topk.Period = 1
	topk.Aggregation = "min"
	topk.Fields = []string{"a"}
	topk.GroupBy = []string{"tag_name"}

	period := time.Second * time.Duration(topk.Period)
	time.Sleep(period)

	ret := topk.Apply(MetricsSet1...)
	answer := MetricsSet1

	if ! equalSets(ret, answer) {
		t.Error("\nExpected metrics:\n", answer, "\nReturned metrics:\n", ret)
	}
}

// AggregationField + avg
// AggregationField + sum
// AggregationField + max
// AggregationField + min

// GroupBy
// GroupBy + Fields
// GroupBy metric name
// GroupBy + GroupBy metric name
// DropNoGroup
// DropNonTop=false + PositionField
// BottomK
