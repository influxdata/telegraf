package v1

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

			// reading fields
			fields := make(map[string]interface{})
			if mf.GetType() == dto.MetricType_SUMMARY {
				// summary metric
				fields = makeQuantiles(m)
				fields["count"] = float64(m.GetSummary().GetSampleCount())
				fields["sum"] = float64(m.GetSummary().GetSampleSum())
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				// histogram metric
				fields = makeBuckets(m)
				fields["count"] = float64(m.GetHistogram().GetSampleCount())
				fields["sum"] = float64(m.GetHistogram().GetSampleSum())

			} else {
				// standard metric
				fields = getNameAndValue(m)
			}
			// converting to telegraf metric
			if len(fields) > 0 {
				var t time.Time
				if m.TimestampMs != nil && *m.TimestampMs > 0 {
					t = time.Unix(0, *m.TimestampMs*1000000)
				} else {
					t = now
				}
				metric, err := metric.New(metricName, tags, fields, t, ValueType(mf.GetType()))
				if err == nil {
					metrics = append(metrics, metric)
				}
			}
		}
	}

	return metrics, err
}

// Get Quantiles from summary metric
func makeQuantiles(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			fields[fmt.Sprint(q.GetQuantile())] = float64(q.GetValue())
		}
	}
	return fields
}

// Get Buckets  from histogram metric
func makeBuckets(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, b := range m.GetHistogram().Bucket {
		fields[fmt.Sprint(b.GetUpperBound())] = float64(b.GetCumulativeCount())
	}
	return fields
}

// Get name and value from metric
func getNameAndValue(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields["gauge"] = float64(m.GetGauge().GetValue())
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			fields["counter"] = float64(m.GetCounter().GetValue())
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			fields["value"] = float64(m.GetUntyped().GetValue())
		}
	}
	return fields
}
