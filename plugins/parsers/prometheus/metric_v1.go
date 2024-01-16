package prometheus

import (
	"math"
	"strconv"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func (p *Parser) extractMetricsV1(prommetrics *dto.MetricFamily) []telegraf.Metric {
	now := time.Now()

	// Convert each prometheus metrics to the corresponding telegraf metrics.
	// You will get one telegraf metric with one field per prometheus metric
	// for "simple" types like Gauge and Counter but a telegraf metric with
	// multiple fields for "complex" types like Summary or Histogram.
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

			// Collect the fields
			fields := make(map[string]interface{}, len(summary.Quantile)+2)
			fields["count"] = float64(summary.GetSampleCount())
			fields["sum"] = summary.GetSampleSum()
			for _, q := range summary.Quantile {
				if v := q.GetValue(); !math.IsNaN(v) {
					fname := strconv.FormatFloat(q.GetQuantile(), 'g', -1, 64)
					fields[fname] = v
				}
			}
			metrics = append(metrics, metric.New(metricName, tags, fields, t, telegraf.Summary))
		case dto.MetricType_HISTOGRAM:
			histogram := pm.GetHistogram()

			// Collect the fields
			fields := make(map[string]interface{}, len(histogram.Bucket)+2)
			fields["count"] = float64(pm.GetHistogram().GetSampleCount())
			fields["sum"] = pm.GetHistogram().GetSampleSum()
			for _, b := range histogram.Bucket {
				fname := strconv.FormatFloat(b.GetUpperBound(), 'g', -1, 64)
				fields[fname] = float64(b.GetCumulativeCount())
			}
			metrics = append(metrics, metric.New(metricName, tags, fields, t, telegraf.Histogram))
		default:
			var fname string
			var v float64
			if gauge := pm.GetGauge(); gauge != nil {
				fname = "gauge"
				v = gauge.GetValue()
			} else if counter := pm.GetCounter(); counter != nil {
				fname = "counter"
				v = counter.GetValue()
			} else if untyped := pm.GetUntyped(); untyped != nil {
				fname = "value"
				v = untyped.GetValue()
			}
			if fname != "" && !math.IsNaN(v) {
				fields := map[string]interface{}{fname: v}
				vtype := mapValueType(metricType)
				metrics = append(metrics, metric.New(metricName, tags, fields, t, vtype))
			}
		}
	}

	return metrics
}
