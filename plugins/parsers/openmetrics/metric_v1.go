package openmetrics

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func (p *Parser) extractMetricsV1(ometrics *MetricFamily) []telegraf.Metric {
	now := time.Now()

	// Convert each prometheus metrics to the corresponding telegraf metrics.
	// You will get one telegraf metric with one field per prometheus metric
	// for "simple" types like Gauge and Counter but a telegraf metric with
	// multiple fields for "complex" types like Summary or Histogram.
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

		// Iterate over the metric points and construct a metric for each
		for _, omp := range om.GetMetricPoints() {
			if omp.Timestamp != nil {
				t = omp.GetTimestamp().AsTime()
			}

			// Construct the metrics
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
				fields := map[string]interface{}{"value": value}
				metrics = append(metrics, metric.New(metricName, tags, fields, t, telegraf.Untyped))
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
				fields := map[string]interface{}{"gauge": value}
				metrics = append(metrics, metric.New(metricName, tags, fields, t, telegraf.Gauge))
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
				fields := map[string]interface{}{"counter": value}
				metrics = append(metrics, metric.New(metricName, tags, fields, t, telegraf.Counter))
			case MetricType_STATE_SET:
				stateset := omp.GetStateSetValue()
				// Collect the fields
				fields := make(map[string]interface{}, len(stateset.States))
				for _, state := range stateset.GetStates() {
					fname := strings.ReplaceAll(state.GetName(), " ", "_")
					fields[fname] = state.GetEnabled()
				}
				metrics = append(metrics, metric.New(metricName, tags, fields, t, telegraf.Untyped))
			case MetricType_INFO:
				info := omp.GetInfoValue().GetInfo()
				fields := map[string]interface{}{"info": uint64(1)}
				mptags := make(map[string]string, len(tags)+len(info))
				for k, v := range tags {
					mptags[k] = v
				}
				for _, itag := range info {
					mptags[itag.Name] = itag.Value
				}
				metrics = append(metrics, metric.New(metricName, mptags, fields, t, telegraf.Untyped))
			case MetricType_HISTOGRAM, MetricType_GAUGE_HISTOGRAM:
				histogram := omp.GetHistogramValue()

				// Collect the fields
				fields := make(map[string]interface{}, len(histogram.Buckets)+3)
				fields["count"] = float64(histogram.GetCount())
				if s := histogram.GetSum(); s != nil {
					switch v := s.(type) {
					case *HistogramValue_DoubleValue:
						fields["sum"] = v.DoubleValue
					case *HistogramValue_IntValue:
						fields["sum"] = float64(v.IntValue)
					}
				}
				if ts := histogram.GetCreated(); ts != nil {
					fields["created"] = float64(ts.Seconds) + float64(ts.Nanos)/float64(time.Nanosecond)
				}
				for _, b := range histogram.Buckets {
					fname := strconv.FormatFloat(b.GetUpperBound(), 'g', -1, 64)
					fields[fname] = float64(b.GetCount())
				}
				metrics = append(metrics, metric.New(metricName, tags, fields, t, telegraf.Histogram))
			case MetricType_SUMMARY:
				summary := omp.GetSummaryValue()

				// Collect the fields
				fields := make(map[string]interface{}, len(summary.Quantile)+2)
				fields["count"] = float64(summary.GetCount())
				if s := summary.GetSum(); s != nil {
					switch v := s.(type) {
					case *SummaryValue_DoubleValue:
						fields["sum"] = v.DoubleValue
					case *SummaryValue_IntValue:
						fields["sum"] = float64(v.IntValue)
					}
				}
				if ts := summary.GetCreated(); ts != nil {
					fields["created"] = float64(ts.Seconds) + float64(ts.Nanos)/float64(time.Second)
				}
				for _, q := range summary.GetQuantile() {
					if v := q.GetValue(); !math.IsNaN(v) {
						fname := strconv.FormatFloat(q.GetQuantile(), 'g', -1, 64)
						fields[fname] = v
					}
				}
				metrics = append(metrics, metric.New(metricName, tags, fields, t, telegraf.Summary))
			}
		}
	}
	return metrics
}
