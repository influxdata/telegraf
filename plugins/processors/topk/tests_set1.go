package topk

import (
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var metric11, _ = metric.New(
	"1one",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(15.3),
		"b": float64(40),
	},
	time.Now(),
)

var metric12, _ = metric.New(
	"1two",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50),
	},
	time.Now(),
)

var metric13, _ = metric.New(
	"1three",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(0.3),
		"c": float64(400),
	},
	time.Now(),
)

var metric14, _ = metric.New(
	"1four",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(24.12),
		"b": float64(40),
	},
	time.Now(),
)

var metric15, _ = metric.New(
	"1five",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50.5),
		"h": float64(1),
		"u": float64(2.4),
	},
	time.Now(),
)

var MetricsSet1 = []telegraf.Metric{metric11, metric12, metric13, metric14, metric15}

var ans_avg_11 = metric11.Copy()
var ans_avg_12 = metric12.Copy()
var ans_avg_13 = metric13.Copy()
var ans_avg_14 = metric14.Copy()
var ans_avg_15 = metric15.Copy()
var AvgAggregationFieldAns = []telegraf.Metric{ans_avg_11, ans_avg_12, ans_avg_13, ans_avg_14, ans_avg_15}

var ans_sum_11 = metric11.Copy()
var ans_sum_12 = metric12.Copy()
var ans_sum_13 = metric13.Copy()
var ans_sum_14 = metric14.Copy()
var ans_sum_15 = metric15.Copy()
var SumAggregationFieldAns = []telegraf.Metric{ans_sum_11, ans_sum_12, ans_sum_13, ans_sum_14, ans_sum_15}

var ans_max_11 = metric11.Copy()
var ans_max_12 = metric12.Copy()
var ans_max_13 = metric13.Copy()
var ans_max_14 = metric14.Copy()
var ans_max_15 = metric15.Copy()
var MaxAggregationFieldAns = []telegraf.Metric{ans_max_11, ans_max_12, ans_max_13, ans_max_14, ans_max_15}

var ans_min_11 = metric11.Copy()
var ans_min_12 = metric12.Copy()
var ans_min_13 = metric13.Copy()
var ans_min_14 = metric14.Copy()
var ans_min_15 = metric15.Copy()
var MinAggregationFieldAns = []telegraf.Metric{ans_min_11, ans_min_12, ans_min_13, ans_min_14, ans_min_15}

func setupTestSet1(){
	// Expected answer for the TopkAvgAggretationField test
	ans_avg_11.AddField("avgag_a", float64(28.044))
	ans_avg_12.AddField("avgag_a", float64(28.044))
	ans_avg_13.AddField("avgag_a", float64(28.044))
	ans_avg_14.AddField("avgag_a", float64(28.044))
	ans_avg_15.AddField("avgag_a", float64(28.044))

	// Expected answer for the TopkSumAggretationField test
	ans_sum_11.AddField("sumag_a", float64(140.22))
	ans_sum_12.AddField("sumag_a", float64(140.22))
	ans_sum_13.AddField("sumag_a", float64(140.22))
	ans_sum_14.AddField("sumag_a", float64(140.22))
	ans_sum_15.AddField("sumag_a", float64(140.22))

	// Expected answer for the TopkMaxAggretationField test
	ans_max_11.AddField("maxag_a", float64(50.5))
	ans_max_12.AddField("maxag_a", float64(50.5))
	ans_max_13.AddField("maxag_a", float64(50.5))
	ans_max_14.AddField("maxag_a", float64(50.5))
	ans_max_15.AddField("maxag_a", float64(50.5))

	// Expected answer for the TopkMinAggretationField test
	ans_min_11.AddField("minag_a", float64(0.3))
	ans_min_12.AddField("minag_a", float64(0.3))
	ans_min_13.AddField("minag_a", float64(0.3))
	ans_min_14.AddField("minag_a", float64(0.3))
	ans_min_15.AddField("minag_a", float64(0.3))
}
