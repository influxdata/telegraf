package topk

import (
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var metric1, _ = metric.New(
	"first_metric_name",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(15.3),
		"b": float64(40),
	},
	time.Now(),
)

var metric2, _ = metric.New(
	"first_metric_name",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(40),
	},
	time.Now(),
)

var metric3, _ = metric.New(
	"first_metric_name",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(0.3),
		"c": float64(400),
	},
	time.Now(),
)

var metric4, _ = metric.New(
	"first_metric_name",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(24.12),
		"b": float64(40),
	},
	time.Now(),
)

var metric5, _ = metric.New(
	"first_metric_name",
	map[string]string{"tag_name": "tag_value1"},
	map[string]interface{}{
		"a": float64(50.8),
		"h": float64(1),
		"u": float64(2.4),
	},
	time.Now(),
)

var MetricsSet1 = [5]telegraf.Metric{metric1, metric2, metric3, metric4, metric5}
