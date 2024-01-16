package prometheus

import (
	"math"
	"strconv"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func (p *Parser) extractMetricsV2(prommetrics *dto.MetricFamily) []telegraf.Metric {
	now := time.Now()

	// Convert each prometheus metric to a corresponding telegraf metric
	// with one field each. The process will filter NaNs in values and skip
	// the corresponding metrics.
	var metrics []telegraf.Metric
	metricName := prommetrics.GetName()
	metricType := prommetrics.GetType()
	for _, pm := range prommetrics.Metric {
		// Extract the timestamp of the metric if it exists and should
		// not be ignored.
		t := now
		if ts := pm.GetTimestampMs(); !p.IgnoreTimestamp && ts > 0 {
			t = time.UnixMilli(ts)
		}

		// Convert the labels to tags
		tags := getTagsFromLabels(pm, p.DefaultTags)

		// Construct the metrics
		switch metricType {
		case dto.MetricType_SUMMARY:
			summary := pm.GetSummary()

			// Add an overall metric containing the number of samples and and its sum
			summaryFields := make(map[string]interface{})
			summaryFields[metricName+"_count"] = float64(summary.GetSampleCount())
			summaryFields[metricName+"_sum"] = summary.GetSampleSum()
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
		case dto.MetricType_HISTOGRAM:
			histogram := pm.GetHistogram()

			// Add an overall metric containing the number of samples and and its sum
			histFields := make(map[string]interface{})
			histFields[metricName+"_count"] = float64(histogram.GetSampleCount())
			histFields[metricName+"_sum"] = histogram.GetSampleSum()
			metrics = append(metrics, metric.New("prometheus", tags, histFields, t, telegraf.Histogram))

			// Add one metric per histogram bucket
			var infSeen bool
			for _, b := range histogram.Bucket {
				bucketTags := tags
				bucketTags["le"] = strconv.FormatFloat(b.GetUpperBound(), 'g', -1, 64)
				bucketFields := map[string]interface{}{
					metricName + "_bucket": float64(b.GetCumulativeCount()),
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
					metricName + "_bucket": float64(histogram.GetSampleCount()),
				}
				m := metric.New("prometheus", infTags, infFields, t, telegraf.Histogram)
				metrics = append(metrics, m)
			}
		default:
			v := math.Inf(1)
			if gauge := pm.GetGauge(); gauge != nil {
				v = gauge.GetValue()
			} else if counter := pm.GetCounter(); counter != nil {
				v = counter.GetValue()
			} else if untyped := pm.GetUntyped(); untyped != nil {
				v = untyped.GetValue()
			}
			if !math.IsNaN(v) {
				fields := map[string]interface{}{metricName: v}
				vtype := mapValueType(metricType)
				metrics = append(metrics, metric.New("prometheus", tags, fields, t, vtype))
			}
		}
	}

	return metrics
}
