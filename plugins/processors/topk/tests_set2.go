package topk

import (
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var metric21, _ = metric.New(
	"metric1",
	map[string]string{
		"tag1": "ONE",
		"tag2": "FIVE",
		"tag3": "SIX",
	},
	map[string]interface{}{
		"value": float64(31.31),
		"A": float64(95.36),
		"C": float64(72.41),
	},
	time.Now(),
)

var metric22, _ = metric.New(
	"metric1",
	map[string]string{
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "THREE",
		
	},
	map[string]interface{}{
		"value": float64(59.43),
		"A": float64(0.6),
	},
	time.Now(),
)

var metric23, _ = metric.New(
	"metric1",
	map[string]string{
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "SIX",
	},
	map[string]interface{}{
		"value": float64(74.18),
		"A": float64(77.42),
		"B": float64(60.96),
	},
	time.Now(),
)

var metric24, _ = metric.New(
	"metric2",
	map[string]string{
		"tag1": "ONE",
		"tag2": "FIVE",
		"tag3": "THREE",
	},
	map[string]interface{}{
		"value": float64(72),
		"B": float64(22.1),
		"C": float64(30.8),
	},
	time.Now(),
)

var metric25, _ = metric.New(
	"metric2",
	map[string]string{
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "SEVEN",
	},
	map[string]interface{}{
		"value": float64(87.92),
		"B": float64(81.55),
		"C": float64(45.1),
	},
	time.Now(),
)

var metric26, _ = metric.New(
	"metric2",
	map[string]string{
		"tag1": "TWO",
		"tag2": "FIVE",
		"tag3": "SEVEN",
	},
	map[string]interface{}{
		"value": float64(75.3),
		"A": float64(29.45),
		"C": float64(4.86),
	},
	time.Now(),
)

var MetricsSet2 = []telegraf.Metric{metric21, metric22, metric23, metric24, metric25, metric26}


var ans23groupby1 = metric23.Copy()
var ans24groupby1 = metric24.Copy()
var ans25groupby1 = metric25.Copy()
var ans26groupby1 = metric26.Copy()
var GroupBy0Ans = []telegraf.Metric{ans23groupby1, ans24groupby1, ans25groupby1, ans26groupby1}

func setupTestSet2(){
	ans23groupby1.AddField("sumag_value", float64(74.18))
	ans24groupby1.AddField("sumag_value", float64(72))
	ans25groupby1.AddField("sumag_value", float64(163.22))
	ans26groupby1.AddField("sumag_value", float64(163.22))
}
