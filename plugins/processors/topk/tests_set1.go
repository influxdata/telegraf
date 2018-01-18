package topk

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"time"
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

var ansAvg11 = metric11.Copy()
var ansAvg12 = metric12.Copy()
var ansAvg13 = metric13.Copy()
var ansAvg14 = metric14.Copy()
var ansAvg15 = metric15.Copy()
var AvgAggregationFieldAns = []telegraf.Metric{ansAvg11, ansAvg12, ansAvg13, ansAvg14, ansAvg15}

var ansSum11 = metric11.Copy()
var ansSum12 = metric12.Copy()
var ansSum13 = metric13.Copy()
var ansSum14 = metric14.Copy()
var ansSum15 = metric15.Copy()
var SumAggregationFieldAns = []telegraf.Metric{ansSum11, ansSum12, ansSum13, ansSum14, ansSum15}

var ansMax11 = metric11.Copy()
var ansMax12 = metric12.Copy()
var ansMax13 = metric13.Copy()
var ansMax14 = metric14.Copy()
var ansMax15 = metric15.Copy()
var MaxAggregationFieldAns = []telegraf.Metric{ansMax11, ansMax12, ansMax13, ansMax14, ansMax15}

var ansMin11 = metric11.Copy()
var ansMin12 = metric12.Copy()
var ansMin13 = metric13.Copy()
var ansMin14 = metric14.Copy()
var ansMin15 = metric15.Copy()
var MinAggregationFieldAns = []telegraf.Metric{ansMin11, ansMin12, ansMin13, ansMin14, ansMin15}

func setupTestSet1() {
	// Expected answer for the TopkAvgAggretationField test
	ansAvg11.AddField("avgag_a", float64(28.044))
	ansAvg12.AddField("avgag_a", float64(28.044))
	ansAvg13.AddField("avgag_a", float64(28.044))
	ansAvg14.AddField("avgag_a", float64(28.044))
	ansAvg15.AddField("avgag_a", float64(28.044))

	// Expected answer for the TopkSumAggretationField test
	ansSum11.AddField("sumag_a", float64(140.22))
	ansSum12.AddField("sumag_a", float64(140.22))
	ansSum13.AddField("sumag_a", float64(140.22))
	ansSum14.AddField("sumag_a", float64(140.22))
	ansSum15.AddField("sumag_a", float64(140.22))

	// Expected answer for the TopkMaxAggretationField test
	ansMax11.AddField("maxag_a", float64(50.5))
	ansMax12.AddField("maxag_a", float64(50.5))
	ansMax13.AddField("maxag_a", float64(50.5))
	ansMax14.AddField("maxag_a", float64(50.5))
	ansMax15.AddField("maxag_a", float64(50.5))

	// Expected answer for the TopkMinAggretationField test
	ansMin11.AddField("minag_a", float64(0.3))
	ansMin12.AddField("minag_a", float64(0.3))
	ansMin13.AddField("minag_a", float64(0.3))
	ansMin14.AddField("minag_a", float64(0.3))
	ansMin15.AddField("minag_a", float64(0.3))
}
