package topk

import (
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var metric21, _ = metric.New(
	"metric1",
	map[string]string{
		"id": "1",
		"tag1": "ONE",
		"tag2": "FIVE",
		"tag3": "SIX",
		"tag4": "EIGHT",
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
		"id": "2",
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "THREE",
		"tag4": "EIGHT",
		
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
		"id": "3",
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "SIX",
		"tag5": "TEN",
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
		"id": "4",
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
		"id": "5",
		"tag1": "TWO",
		"tag2": "FOUR",
		"tag3": "SEVEN",
		"tag4": "NINE",
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
		"id": "6",
		"tag1": "TWO",
		"tag2": "FIVE",
		"tag3": "SEVEN",
		"tag4": "NINE",
	},
	map[string]interface{}{
		"value": float64(75.3),
		"A": float64(29.45),
		"C": float64(4.86),
	},
	time.Now(),
)

var MetricsSet2 = []telegraf.Metric{metric21, metric22, metric23, metric24, metric25, metric26}


// Groupby tests
var ans23groupby1 = metric23.Copy()
var ans24groupby1 = metric24.Copy()
var ans25groupby1 = metric25.Copy()
var ans26groupby1 = metric26.Copy()
var GroupBy1Ans = []telegraf.Metric{ans23groupby1, ans24groupby1, ans25groupby1, ans26groupby1}

var ans22groupby2 = metric22.Copy()
var ans23groupby2 = metric23.Copy()
var ans25groupby2 = metric25.Copy()
var ans26groupby2 = metric26.Copy()
var GroupBy2Ans = []telegraf.Metric{ans22groupby2, ans23groupby2, ans25groupby2, ans26groupby2}

var ans25groupby3 = metric25.Copy()
var ans26groupby3 = metric26.Copy()
var GroupBy3Ans = []telegraf.Metric{ans25groupby3, ans26groupby3}


// Groupby + Field tests
var ans21groupby4 = metric21.Copy()
var ans22groupby4 = metric22.Copy()
var ans23groupby4 = metric23.Copy()
var ans24groupby4 = metric24.Copy()
var ans25groupby4 = metric25.Copy()
var GroupBy4Ans = []telegraf.Metric{ans21groupby4, ans22groupby4, ans23groupby4, ans24groupby4, ans25groupby4}

var ans21groupby5 = metric21.Copy()
var ans23groupby5 = metric23.Copy()
var ans25groupby5 = metric25.Copy()
var ans26groupby5 = metric26.Copy()
var GroupBy5Ans = []telegraf.Metric{ans21groupby5, ans23groupby5, ans25groupby5, ans26groupby5}


// Groupby Metric name tests
var ans24groupbymetric1 = metric24.Copy()
var ans25groupbymetric1 = metric25.Copy()
var ans26groupbymetric1 = metric26.Copy()
var GroupByMetric1Ans = []telegraf.Metric{ans24groupbymetric1, ans25groupbymetric1, ans26groupbymetric1}

var ans21groupbymetric2 = metric24.Copy()
var ans22groupbymetric2 = metric25.Copy()
var ans23groupbymetric2 = metric26.Copy()
var ans25groupbymetric2 = metric25.Copy()
var GroupByMetric2Ans = []telegraf.Metric{ans21groupbymetric2, ans22groupbymetric2, ans23groupbymetric2, ans25groupbymetric2}

// DropNoGroup
var ans21dropnogroup = metric24.Copy()
var ans22dropnogroup = metric25.Copy()
var ans24dropnogroup = metric26.Copy()
var ans25dropnogroup = metric25.Copy()
var ans26dropnogroup = metric25.Copy()
var DropNoGroupFalseAns = []telegraf.Metric{ans21dropnogroup, ans22dropnogroup, ans24dropnogroup, ans25dropnogroup, ans26dropnogroup}

// DropNonTop=false + PositionField
var ans21dontdropbot = metric23.Copy()
var ans22dontdropbot = metric23.Copy()
var ans23dontdropbot = metric23.Copy()
var ans24dontdropbot = metric24.Copy()
var ans25dontdropbot = metric25.Copy()
var ans26dontdropbot = metric26.Copy()
var DontDropBottomAns = []telegraf.Metric{ans21dontdropbot, ans22dontdropbot, ans23dontdropbot, ans24dontdropbot, ans25dontdropbot, ans26dontdropbot}

// Bottomk
var ans23bottomk = metric23.Copy()
var ans24bottomk = metric24.Copy()
var ans25bottomk = metric25.Copy()
var ans26bottomk = metric26.Copy()
var BottomkAns = []telegraf.Metric{ans26bottomk, ans25bottomk, ans24bottomk, ans23bottomk}

// No drops
var ans21nodrops1 = metric22.Copy()
var ans22nodrops1 = metric22.Copy()
var ans23nodrops1 = metric23.Copy()
var ans24nodrops1 = metric22.Copy()
var ans25nodrops1 = metric25.Copy()
var ans26nodrops1 = metric26.Copy()
var NoDropsAns1 = []telegraf.Metric{ans22nodrops1, ans23nodrops1, ans25nodrops1, ans26nodrops1}

func setupTestSet2(){
	ans23groupby1.AddField("sumag_value", float64(74.18))
	ans24groupby1.AddField("sumag_value", float64(72))
	ans25groupby1.AddField("sumag_value", float64(163.22))
	ans26groupby1.AddField("sumag_value", float64(163.22))

	ans22groupby2.AddField("avg_value", float64(74.20750000000001))
	ans23groupby2.AddField("avg_value", float64(74.20750000000001))
	ans25groupby2.AddField("avg_value", float64(74.20750000000001))
	ans26groupby2.AddField("avg_value", float64(74.20750000000001))

	ans25groupby3.AddField("minaggfield_value", float64(75.3))
	ans26groupby3.AddField("minaggfield_value", float64(75.3))

	ans21groupby4.AddField("avg_A", float64(95.36))
	ans22groupby4.AddField("avg_A", float64(39.01))
	ans23groupby4.AddField("avg_A", float64(39.01))

	ans21groupby5.AddField("sum_C", float64(72.41))
	ans23groupby5.AddField("sum_B", float64(60.96))
	ans25groupby5.AddField("sum_B", float64(81.55))
	ans25groupby5.AddField("sum_C", float64(49.96))
	ans26groupby5.AddField("sum_C", float64(49.96))

	ans24groupbymetric1.AddField("sigma_value", float64(235.22))
	ans25groupbymetric1.AddField("sigma_value", float64(235.22))
	ans26groupbymetric1.AddField("sigma_value", float64(235.22))

	ans21groupbymetric2.AddField("SUM_value", float64(31.31))
	ans21groupbymetric2.AddField("SUM_A", float64(95.36))
	ans22groupbymetric2.AddField("SUM_value", float64(133.61))
	ans22groupbymetric2.AddField("SUM_A", float64(78.02))
	ans23groupbymetric2.AddField("SUM_value", float64(133.61))
	ans23groupbymetric2.AddField("SUM_A", float64(78.02))
	ans25groupbymetric2.AddField("SUM_value", float64(87.92))

	ans21dontdropbot.AddField("sumag_value", float64(31.31))
	ans22dontdropbot.AddField("sumag_value", float64(59.43))
	ans23dontdropbot.AddField("sumag_value", float64(74.18))
	ans23dontdropbot.AddField("aggpos_value", 2)
	ans24dontdropbot.AddField("sumag_value", float64(72))
	ans24dontdropbot.AddField("aggpos_value", 3)
	ans25dontdropbot.AddField("sumag_value", float64(163.22))
	ans25dontdropbot.AddField("aggpos_value", 1)
	ans26dontdropbot.AddField("sumag_value", float64(163.22))
	ans26dontdropbot.AddField("aggpos_value", 1)

	ans23nodrops1.AddField("sumag_value", float64(74.18))
	ans23nodrops1.AddField("aggpos_value", 2)
	ans24nodrops1.AddField("sumag_value", float64(72))
	ans24nodrops1.AddField("aggpos_value", 3)
	ans25nodrops1.AddField("sumag_value", float64(163.22))
	ans25nodrops1.AddField("aggpos_value", 1)
	ans26nodrops1.AddField("sumag_value", float64(163.22))
	ans26nodrops1.AddField("aggpos_value", 1)
}
