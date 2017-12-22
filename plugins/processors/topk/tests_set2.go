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
var ans23GroupBy1 = metric23.Copy()
var ans24GroupBy1 = metric24.Copy()
var ans25GroupBy1 = metric25.Copy()
var ans26GroupBy1 = metric26.Copy()
var GroupBy1Ans = []telegraf.Metric{ans23GroupBy1, ans24GroupBy1, ans25GroupBy1, ans26GroupBy1}

var ans22GroupBy2 = metric22.Copy()
var ans23GroupBy2 = metric23.Copy()
var ans25GroupBy2 = metric25.Copy()
var ans26GroupBy2 = metric26.Copy()
var GroupBy2Ans = []telegraf.Metric{ans22GroupBy2, ans23GroupBy2, ans25GroupBy2, ans26GroupBy2}

var ans25GroupBy3 = metric25.Copy()
var ans26GroupBy3 = metric26.Copy()
var GroupBy3Ans = []telegraf.Metric{ans25GroupBy3, ans26GroupBy3}


// GroupBy + Field tests
var ans21GroupBy4 = metric21.Copy()
var ans22GroupBy4 = metric22.Copy()
var ans23GroupBy4 = metric23.Copy()
var ans24GroupBy4 = metric24.Copy()
var ans25GroupBy4 = metric25.Copy()
var GroupBy4Ans = []telegraf.Metric{ans21GroupBy4, ans22GroupBy4, ans23GroupBy4, ans24GroupBy4, ans25GroupBy4}

var ans21GroupBy5 = metric21.Copy()
var ans23GroupBy5 = metric23.Copy()
var ans25GroupBy5 = metric25.Copy()
var ans26GroupBy5 = metric26.Copy()
var GroupBy5Ans = []telegraf.Metric{ans21GroupBy5, ans23GroupBy5, ans25GroupBy5, ans26GroupBy5}


// GroupBy Metric
var ans24GroupByMetric1 = metric24.Copy()
var ans25GroupByMetric1 = metric25.Copy()
var ans26GroupByMetric1 = metric26.Copy()
var GroupByMetric1Ans = []telegraf.Metric{ans24GroupByMetric1, ans25GroupByMetric1, ans26GroupByMetric1}

var ans21GroupByMetric2 = metric21.Copy()
var ans22GroupByMetric2 = metric22.Copy()
var ans23GroupByMetric2 = metric23.Copy()
var ans25GroupByMetric2 = metric25.Copy()
var GroupByMetric2Ans = []telegraf.Metric{ans21GroupByMetric2, ans22GroupByMetric2, ans23GroupByMetric2, ans25GroupByMetric2}

// DropNoGroup
var ans21DropNoGroup = metric21.Copy()
var ans22DropNoGroup = metric22.Copy()
var ans24DropNoGroup = metric24.Copy()
var ans25DropNoGroup = metric25.Copy()
var ans26DropNoGroup = metric26.Copy()
var DropNoGroupFalseAns = []telegraf.Metric{ans21DropNoGroup, ans22DropNoGroup, ans24DropNoGroup, ans25DropNoGroup, ans26DropNoGroup}

// DropNonTop=false + PositionField
var ans21DontDropBot = metric21.Copy()
var ans22DontDropBot = metric22.Copy()
var ans23DontDropBot = metric23.Copy()
var ans24DontDropBot = metric24.Copy()
var ans25DontDropBot = metric25.Copy()
var ans26DontDropBot = metric26.Copy()
var DontDropBottomAns = []telegraf.Metric{ans21DontDropBot, ans22DontDropBot, ans23DontDropBot, ans24DontDropBot, ans25DontDropBot, ans26DontDropBot}

// BottomK
var ans21BottomK = metric21.Copy()
var ans22BottomK = metric22.Copy()
var ans24BottomK = metric24.Copy()
var BottomKAns = []telegraf.Metric{ans21BottomK, ans22BottomK, ans24BottomK}

// GroupByKeyTag
var ans21GroupByKeyTag = metric21.Copy()
var ans22GroupByKeyTag = metric22.Copy()
var ans23GroupByKeyTag = metric23.Copy()
var ans24GroupByKeyTag = metric24.Copy()
var ans25GroupByKeyTag = metric25.Copy()
var ans26GroupByKeyTag = metric26.Copy()
var GroupByKeyTagAns = []telegraf.Metric{ans21GroupByKeyTag, ans22GroupByKeyTag, ans23GroupByKeyTag, ans24GroupByKeyTag, ans25GroupByKeyTag, ans26GroupByKeyTag}

// No drops
var ans21NoDrops1 = metric21.Copy()
var ans22NoDrops1 = metric22.Copy()
var ans23NoDrops1 = metric23.Copy()
var ans24NoDrops1 = metric24.Copy()
var ans25NoDrops1 = metric25.Copy()
var ans26NoDrops1 = metric26.Copy()
var NoDropsAns1 = []telegraf.Metric{ans21NoDrops1, ans22NoDrops1, ans23NoDrops1, ans24NoDrops1, ans25NoDrops1, ans26NoDrops1}

// Simple topk
var ans23SimpleTopK = metric23.Copy()
var ans25SimpleTopK = metric25.Copy()
var ans26SimpleTopK = metric26.Copy()
var SimpleTopKAns = []telegraf.Metric{ans23SimpleTopK, ans25SimpleTopK, ans26SimpleTopK}

func setupTestSet2(){
	ans23GroupBy1.AddField("sumag_value", float64(74.18))
	ans24GroupBy1.AddField("sumag_value", float64(72))
	ans25GroupBy1.AddField("sumag_value", float64(163.22))
	ans26GroupBy1.AddField("sumag_value", float64(163.22))

	ans22GroupBy2.AddField("avg_value", float64(74.20750000000001))
	ans23GroupBy2.AddField("avg_value", float64(74.20750000000001))
	ans25GroupBy2.AddField("avg_value", float64(74.20750000000001))
	ans26GroupBy2.AddField("avg_value", float64(74.20750000000001))

	ans25GroupBy3.AddField("minaggfield_value", float64(75.3))
	ans26GroupBy3.AddField("minaggfield_value", float64(75.3))

	ans21GroupBy4.AddField("avg_A", float64(95.36))
	ans22GroupBy4.AddField("avg_A", float64(39.01))
	ans23GroupBy4.AddField("avg_A", float64(39.01))

	ans21GroupBy5.AddField("sum_C", float64(72.41))
	ans23GroupBy5.AddField("sum_B", float64(60.96))
	ans25GroupBy5.AddField("sum_B", float64(81.55))
	ans25GroupBy5.AddField("sum_C", float64(49.96))
	ans26GroupBy5.AddField("sum_C", float64(49.96))

	ans24GroupByMetric1.AddField("sigma_value", float64(235.22000000000003))
	ans25GroupByMetric1.AddField("sigma_value", float64(235.22000000000003))
	ans26GroupByMetric1.AddField("sigma_value", float64(235.22000000000003))

	ans21GroupByMetric2.AddField("SUM_A", float64(95.36))
	ans22GroupByMetric2.AddField("SUM_A", float64(78.02))
	ans22GroupByMetric2.AddField("SUM_value", float64(133.61))
	ans23GroupByMetric2.AddField("SUM_A", float64(78.02))
	ans23GroupByMetric2.AddField("SUM_value", float64(133.61))
	ans25GroupByMetric2.AddField("SUM_value", float64(87.92))

	ans23DontDropBot.AddField("sumag_value", float64(74.18))
	ans23DontDropBot.AddField("aggpos_value", 2)
	ans24DontDropBot.AddField("sumag_value", float64(72))
	ans24DontDropBot.AddField("aggpos_value", 3)
	ans25DontDropBot.AddField("sumag_value", float64(163.22))
	ans25DontDropBot.AddField("aggpos_value", 1)
	ans26DontDropBot.AddField("sumag_value", float64(163.22))
	ans26DontDropBot.AddField("aggpos_value", 1)

	ans23GroupByKeyTag.AddTag("gbt", "tag1=TWO&tag3=SIX&")
	ans24GroupByKeyTag.AddTag("gbt", "tag1=ONE&tag3=THREE&")
	ans25GroupByKeyTag.AddTag("gbt", "tag1=TWO&tag3=SEVEN&")
	ans26GroupByKeyTag.AddTag("gbt", "tag1=TWO&tag3=SEVEN&")

	ans23NoDrops1.AddField("sumag_value", float64(74.18))
	ans23NoDrops1.AddField("aggpos_value", 2)
	ans24NoDrops1.AddField("sumag_value", float64(72))
	ans24NoDrops1.AddField("aggpos_value", 3)
	ans25NoDrops1.AddField("sumag_value", float64(163.22))
	ans25NoDrops1.AddField("aggpos_value", 1)
	ans26NoDrops1.AddField("sumag_value", float64(163.22))
	ans26NoDrops1.AddField("aggpos_value", 1)

	ans23SimpleTopK.HashID()
	ans25SimpleTopK.HashID()
	ans26SimpleTopK.HashID()
}
