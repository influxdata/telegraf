package v2

import (
	"fmt"
	"math"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	. "github.com/influxdata/telegraf/plugins/parsers/prometheus/common"

	dto "github.com/prometheus/client_model/go"
)

// Parse returns a slice of Metrics from a text representation of a
// metrics
func Parse(metricFamilies map[string]*dto.MetricFamily, defaultTags map[string]string, now time.Time) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric
	var err error

	// read metrics
	for metricName, mf := range metricFamilies {
		for _, m := range mf.Metric {
			// reading tags
			tags := MakeLabels(m, defaultTags)

			if mf.GetType() == dto.MetricType_SUMMARY {
				// summary metric
				telegrafMetrics := makeQuantiles(m, tags, metricName, mf.GetType(), now)
				metrics = append(metrics, telegrafMetrics...)
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				// histogram metric
				telegrafMetrics := makeBuckets(m, tags, metricName, mf.GetType(), now)
				metrics = append(metrics, telegrafMetrics...)
			} else {
				// standard metric
				// reading fields
				fields := make(map[string]interface{})
				fields = getNameAndValue(m, metricName)
				// converting to telegraf metric
				if len(fields) > 0 {
					var t time.Time
					if m.TimestampMs != nil && *m.TimestampMs > 0 {
						t = time.Unix(0, *m.TimestampMs*1000000)
					} else {
						t = now
					}
					metric, err := metric.New("prometheus", tags, fields, t, ValueType(mf.GetType()))
					if err == nil {
						metrics = append(metrics, metric)
					}
				}
			}
		}
	}

	return metrics, err
}

// Get Quantiles for summary metric & Buckets for histogram
func makeQuantiles(m *dto.Metric, tags map[string]string, metricName string, metricType dto.MetricType, now time.Time) []telegraf.Metric {
	var metrics []telegraf.Metric
	fields := make(map[string]interface{})
	var t time.Time
	if m.TimestampMs != nil && *m.TimestampMs > 0 {
		t = time.Unix(0, *m.TimestampMs*1000000)
	} else {
		t = now
	}
	fields[metricName+"_count"] = float64(m.GetSummary().GetSampleCount())
	fields[metricName+"_sum"] = float64(m.GetSummary().GetSampleSum())
	met, err := metric.New("prometheus", tags, fields, t, ValueType(metricType))
	if err == nil {
		metrics = append(metrics, met)
	}

	for _, q := range m.GetSummary().Quantile {
		newTags := tags
		fields = make(map[string]interface{})

		newTags["quantile"] = fmt.Sprint(q.GetQuantile())
		fields[metricName] = float64(q.GetValue())

		quantileMetric, err := metric.New("prometheus", newTags, fields, t, ValueType(metricType))
		if err == nil {
			metrics = append(metrics, quantileMetric)
		}
	}
	return metrics
}

// Get Buckets  from histogram metric
func makeBuckets(m *dto.Metric, tags map[string]string, metricName string, metricType dto.MetricType, now time.Time) []telegraf.Metric {
	var metrics []telegraf.Metric
	fields := make(map[string]interface{})
	var t time.Time
	if m.TimestampMs != nil && *m.TimestampMs > 0 {
		t = time.Unix(0, *m.TimestampMs*1000000)
	} else {
		t = now
	}
	fields[metricName+"_count"] = float64(m.GetHistogram().GetSampleCount())
	fields[metricName+"_sum"] = float64(m.GetHistogram().GetSampleSum())

	met, err := metric.New("prometheus", tags, fields, t, ValueType(metricType))
	if err == nil {
		metrics = append(metrics, met)
	}

	for _, b := range m.GetHistogram().Bucket {
		newTags := tags
		fields = make(map[string]interface{})
		newTags["le"] = fmt.Sprint(b.GetUpperBound())
		fields[metricName+"_bucket"] = float64(b.GetCumulativeCount())

		histogramMetric, err := metric.New("prometheus", newTags, fields, t, ValueType(metricType))
		if err == nil {
			metrics = append(metrics, histogramMetric)
		}
	}
	return metrics
}

// Get name and value from metric
func getNameAndValue(m *dto.Metric, metricName string) map[string]interface{} {
	fields := make(map[string]interface{})
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields[metricName] = float64(m.GetGauge().GetValue())
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			fields[metricName] = float64(m.GetCounter().GetValue())
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			fields[metricName] = float64(m.GetUntyped().GetValue())
		}
	}
	return fields
}
