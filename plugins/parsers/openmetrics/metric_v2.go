package openmetrics

import (
	"math"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func (p *Parser) extractMetricsV2(ometrics *MetricFamily) []telegraf.Metric {
	now := time.Now()

	// Convert each prometheus metric to a corresponding telegraf metric
	// with one field each. The process will filter NaNs in values and skip
	// the corresponding metrics.
	var metrics []telegraf.Metric
	metricName := ometrics.GetName()
	metricType := ometrics.GetType()
	for _, om := range ometrics.GetMetrics() {
		// Extract the timestamp of the metric if it exists and should
		// not be ignored.
		t := now

		// Convert the labels to tags
		tags := getTagsFromLabels(om, p.DefaultTags)
		if ometrics.Unit != "" {
			tags["unit"] = ometrics.Unit
		}

		// Construct the metrics
		for _, omp := range om.GetMetricPoints() {
			if omp.Timestamp != nil {
				t = omp.GetTimestamp().AsTime()
			}

			switch metricType {
			case MetricType_UNKNOWN:
				x := omp.GetUnknownValue().GetValue()
				if x == nil {
					continue
				}
				var value float64
				switch v := x.(type) {
				case *UnknownValue_DoubleValue:
					value = v.DoubleValue
				case *UnknownValue_IntValue:
					value = float64(v.IntValue)
				}
				if math.IsNaN(value) {
					continue
				}
				fields := map[string]interface{}{metricName: value}
				metrics = append(metrics, metric.New("prometheus", tags, fields, t, telegraf.Untyped))
			case MetricType_GAUGE:
				x := omp.GetGaugeValue().GetValue()
				if x == nil {
					continue
				}
				var value float64
				switch v := x.(type) {
				case *GaugeValue_DoubleValue:
					value = v.DoubleValue
				case *GaugeValue_IntValue:
					value = float64(v.IntValue)
				}
				if math.IsNaN(value) {
					continue
				}
				fields := map[string]interface{}{metricName: value}
				metrics = append(metrics, metric.New("prometheus", tags, fields, t, telegraf.Gauge))
			case MetricType_COUNTER:
				x := omp.GetCounterValue().GetTotal()
				if x == nil {
					continue
				}
				var value float64
				switch v := x.(type) {
				case *CounterValue_DoubleValue:
					value = v.DoubleValue
				case *CounterValue_IntValue:
					value = float64(v.IntValue)
				}
				if math.IsNaN(value) {
					continue
				}
				fields := map[string]interface{}{metricName: value}
				metrics = append(metrics, metric.New("prometheus", tags, fields, t, telegraf.Counter))
			case MetricType_STATE_SET:
				stateset := omp.GetStateSetValue()

				// Add one metric per state
				for _, state := range stateset.GetStates() {
					fields := map[string]interface{}{metricName + "_" + state.Name: state.GetEnabled()}
					metrics = append(metrics, metric.New("prometheus", tags, fields, t, telegraf.Untyped))
				}
			case MetricType_INFO:
				infos := omp.GetInfoValue().GetInfo()
				for _, info := range infos {
					fields := map[string]interface{}{metricName + "_" + info.Name: info.GetValue()}
					metrics = append(metrics, metric.New("prometheus", tags, fields, t, telegraf.Untyped))
				}
			case MetricType_HISTOGRAM, MetricType_GAUGE_HISTOGRAM:
				histogram := omp.GetHistogramValue()

				// Add an overall metric containing the number of samples and and its sum
				histFields := make(map[string]interface{})
				histFields[metricName+"_count"] = float64(histogram.GetCount())
				if s := histogram.GetSum(); s != nil {
					switch v := s.(type) {
					case *HistogramValue_DoubleValue:
						histFields[metricName+"_sum"] = v.DoubleValue
					case *HistogramValue_IntValue:
						histFields[metricName+"_sum"] = float64(v.IntValue)
					}
				}
				if ts := histogram.GetCreated(); ts != nil {
					histFields[metricName+"_created"] = float64(ts.Seconds) + float64(ts.Nanos)/float64(time.Nanosecond)
				}
				metrics = append(metrics, metric.New("prometheus", tags, histFields, t, telegraf.Histogram))

				// Add one metric per histogram bucket
				var infSeen bool
				for _, b := range histogram.GetBuckets() {
					bucketTags := tags
					bucketTags["le"] = strconv.FormatFloat(b.GetUpperBound(), 'g', -1, 64)
					bucketFields := map[string]interface{}{
						metricName + "_bucket": float64(b.GetCount()),
					}
					m := metric.New("prometheus", bucketTags, bucketFields, t, telegraf.Histogram)
					metrics = append(metrics, m)

					// Record if any of the buckets marks an infinite upper bound
					infSeen = infSeen || math.IsInf(b.GetUpperBound(), +1)
				}

				// Infinity bucket is required for proper function of histogram in prometheus
				if !infSeen {
					infTags := tags
					infTags["le"] = "+Inf"
					infFields := map[string]interface{}{
						metricName + "_bucket": float64(histogram.GetCount()),
					}
					m := metric.New("prometheus", infTags, infFields, t, telegraf.Histogram)
					metrics = append(metrics, m)
				}
			case MetricType_SUMMARY:
				summary := omp.GetSummaryValue()

				// Add an overall metric containing the number of samples and and its sum
				summaryFields := make(map[string]interface{})
				summaryFields[metricName+"_count"] = float64(summary.GetCount())

				if s := summary.GetSum(); s != nil {
					switch v := s.(type) {
					case *SummaryValue_DoubleValue:
						summaryFields[metricName+"_sum"] = v.DoubleValue
					case *SummaryValue_IntValue:
						summaryFields[metricName+"_sum"] = float64(v.IntValue)
					}
				}
				if ts := summary.GetCreated(); ts != nil {
					summaryFields[metricName+"_created"] = float64(ts.Seconds) + float64(ts.Nanos)/float64(time.Nanosecond)
				}
				metrics = append(metrics, metric.New("prometheus", tags, summaryFields, t, telegraf.Summary))

				// Add one metric per quantile
				for _, q := range summary.Quantile {
					quantileTags := tags
					quantileTags["quantile"] = strconv.FormatFloat(q.GetQuantile(), 'g', -1, 64)
					quantileFields := map[string]interface{}{
						metricName: q.GetValue(),
					}
					m := metric.New("prometheus", quantileTags, quantileFields, t, telegraf.Summary)
					metrics = append(metrics, m)
				}
			}
		}
	}
	return metrics
}
