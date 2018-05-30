package topk

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"time"
)

///// Test set 1 /////
var metric11, _ = metric.New(
	"m1",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(15.3),
		"b": float64(40),
	},
	time.Now(),
)

var metric12, _ = metric.New(
	"m1",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50),
	},
	time.Now(),
)

var metric13, _ = metric.New(
	"m1",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(0.3),
		"c": float64(400),
	},
	time.Now(),
)

var metric14, _ = metric.New(
	"m1",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(24.12),
		"b": float64(40),
	},
	time.Now(),
)

var metric15, _ = metric.New(
	"m1",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50.5),
		"h": float64(1),
		"u": float64(2.4),
	},
	time.Now(),
)

var MetricsSet1 = []telegraf.Metric{metric11, metric12, metric13, metric14, metric15}

///// Test set 2 /////
var metric21, _ = metric.New(
	"metric1",
	map[string]string{
		"id":   "1",
		"tag1": "ONE",
		"tag2": "FIVE",
		"tag3": "SIX",
		"tag4": "EIGHT",
	},
	map[string]interface{}{
		"value": float64(31.31),
		"A":     float64(95.36),
		"C":     float64(72.41),
	},
	time.Now(),
)

var metric22, _ = metric.New(
	"metric1",
	map[string]string{
		"id":   "2",
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "THREE",
		"tag4": "EIGHT",
	},
	map[string]interface{}{
		"value": float64(59.43),
		"A":     float64(0.6),
	},
	time.Now(),
)

var metric23, _ = metric.New(
	"metric1",
	map[string]string{
		"id":   "3",
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "SIX",
		"tag5": "TEN",
	},
	map[string]interface{}{
		"value": float64(74.18),
		"A":     float64(77.42),
		"B":     float64(60.96),
	},
	time.Now(),
)

var metric24, _ = metric.New(
	"metric2",
	map[string]string{
		"id":   "4",
		"tag1": "ONE",
		"tag2": "FIVE",
		"tag3": "THREE",
	},
	map[string]interface{}{
		"value": float64(72),
		"B":     float64(22.1),
		"C":     float64(30.8),
	},
	time.Now(),
)

var metric25, _ = metric.New(
	"metric2",
	map[string]string{
		"id":   "5",
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "SEVEN",
		"tag4": "NINE",
	},
	map[string]interface{}{
		"value": float64(87.92),
		"B":     float64(81.55),
		"C":     float64(45.1),
	},
	time.Now(),
)

var metric26, _ = metric.New(
	"metric2",
	map[string]string{
		"id":   "6",
		"tag1": "TWO",
		"tag2": "FIVE",
		"tag3": "SEVEN",
		"tag4": "NINE",
	},
	map[string]interface{}{
		"value": float64(75.3),
		"A":     float64(29.45),
		"C":     float64(4.86),
	},
	time.Now(),
)

var MetricsSet2 = []telegraf.Metric{metric21, metric22, metric23, metric24, metric25, metric26}
