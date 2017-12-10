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

// Expected answer for the TopkAvgAggretationField test
var ans_avg_11, _ = metric.New(
	"1one",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(15.3),
		"b": float64(40),
		"avgag_a": float64(28.044),
	},
	metric11.Time(),
)

var ans_avg_12, _ = metric.New(
	"1two",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50),
		"avgag_a": float64(28.044),
	},
	metric12.Time(),
)

var ans_avg_13, _ = metric.New(
	"1three",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(0.3),
		"c": float64(400),
		"avgag_a": float64(28.044),
	},
	metric13.Time(),
)

var ans_avg_14, _ = metric.New(
	"1four",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(24.12),
		"b": float64(40),
		"avgag_a": float64(28.044),
	},
	metric14.Time(),
)

var ans_avg_15, _ = metric.New(
	"1five",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50.5),
		"h": float64(1),
		"u": float64(2.4),
		"avgag_a": float64(28.044),
	},
	metric15.Time(),
)

var AvgAggregationFieldAns = []telegraf.Metric{ans_avg_11, ans_avg_12, ans_avg_13, ans_avg_14, ans_avg_15}


// Expected answer for the TopkSumAggretationField test
var ans_sum_11, _ = metric.New(
	"1one",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(15.3),
		"b": float64(40),
		"sumag_a": float64(140.22),
	},
	metric11.Time(),
)

var ans_sum_12, _ = metric.New(
	"1two",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50),
		"sumag_a": float64(140.22),
	},
	metric12.Time(),
)

var ans_sum_13, _ = metric.New(
	"1three",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(0.3),
		"c": float64(400),
		"sumag_a": float64(140.22),
	},
	metric13.Time(),
)

var ans_sum_14, _ = metric.New(
	"1four",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(24.12),
		"b": float64(40),
		"sumag_a": float64(140.22),
	},
	metric14.Time(),
)

var ans_sum_15, _ = metric.New(
	"1five",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50.5),
		"h": float64(1),
		"u": float64(2.4),
		"sumag_a": float64(140.22),
	},
	metric15.Time(),
)

var SumAggregationFieldAns = []telegraf.Metric{ans_sum_11, ans_sum_12, ans_sum_13, ans_sum_14, ans_sum_15}


// Expected answer for the TopkSumAggretationField test
var ans_max_11, _ = metric.New(
	"1one",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(15.3),
		"b": float64(40),
		"maxag_a": float64(50.5),
	},
	metric11.Time(),
)

var ans_max_12, _ = metric.New(
	"1two",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50),
		"maxag_a": float64(50.5),
	},
	metric12.Time(),
)

var ans_max_13, _ = metric.New(
	"1three",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(0.3),
		"c": float64(400),
		"maxag_a": float64(50.5),
	},
	metric13.Time(),
)

var ans_max_14, _ = metric.New(
	"1four",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(24.12),
		"b": float64(40),
		"maxag_a": float64(50.5),
	},
	metric14.Time(),
)

var ans_max_15, _ = metric.New(
	"1five",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50.5),
		"h": float64(1),
		"u": float64(2.4),
		"maxag_a": float64(50.5),
	},
	metric15.Time(),
)

var MaxAggregationFieldAns = []telegraf.Metric{ans_max_11, ans_max_12, ans_max_13, ans_max_14, ans_max_15}


// Expected answer for the TopkSumAggretationField test
var ans_min_11, _ = metric.New(
	"1one",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(15.3),
		"b": float64(40),
		"minag_a": float64(0.3),
	},
	metric11.Time(),
)

var ans_min_12, _ = metric.New(
	"1two",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50),
		"minag_a": float64(0.3),
	},
	metric12.Time(),
)

var ans_min_13, _ = metric.New(
	"1three",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(0.3),
		"c": float64(400),
		"minag_a": float64(0.3),
	},
	metric13.Time(),
)

var ans_min_14, _ = metric.New(
	"1four",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(24.12),
		"b": float64(40),
		"minag_a": float64(0.3),
	},
	metric14.Time(),
)

var ans_min_15, _ = metric.New(
	"1five",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50.5),
		"h": float64(1),
		"u": float64(2.4),
		"minag_a": float64(0.3),
	},
	metric15.Time(),
)

var MinAggregationFieldAns = []telegraf.Metric{ans_min_11, ans_min_12, ans_min_13, ans_min_14, ans_min_15}
