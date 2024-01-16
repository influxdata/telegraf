package prometheus

import (
	"fmt"
	"math"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func (p *Parser) extractMetricsV1(metricFamilies map[string]*dto.MetricFamily) []telegraf.Metric {
	var metrics []telegraf.Metric

	now := time.Now()
	// read metrics
	for metricName, mf := range metricFamilies {
		for _, m := range mf.Metric {
			// reading tags
			tags := GetTagsFromLabels(m, p.DefaultTags)

			// reading fields
			var fields map[string]interface{}
			if mf.GetType() == dto.MetricType_SUMMARY {
				// summary metric
				fields = makeQuantilesV1(m)
				fields["count"] = float64(m.GetSummary().GetSampleCount())
				//nolint:unconvert // Conversion may be needed for float64 https://github.com/mdempsky/unconvert/issues/40
				fields["sum"] = float64(m.GetSummary().GetSampleSum())
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				// histogram metric
				fields = makeBucketsV1(m)
				fields["count"] = float64(m.GetHistogram().GetSampleCount())
				//nolint:unconvert // Conversion may be needed for float64 https://github.com/mdempsky/unconvert/issues/40
				fields["sum"] = float64(m.GetHistogram().GetSampleSum())
			} else {
				// standard metric
				fields = getNameAndValueV1(m)
			}
			// converting to telegraf metric
			if len(fields) > 0 {
				var t time.Time
				if !p.IgnoreTimestamp && m.TimestampMs != nil && *m.TimestampMs > 0 {
					t = time.Unix(0, *m.TimestampMs*1000000)
				} else {
					t = now
				}
				m := metric.New(metricName, tags, fields, t, ValueType(mf.GetType()))
				metrics = append(metrics, m)
			}
		}
	}

	return metrics
}

// Get Quantiles from summary metric
func makeQuantilesV1(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			//nolint:unconvert // Conversion may be needed for float64 https://github.com/mdempsky/unconvert/issues/40
			fields[fmt.Sprint(q.GetQuantile())] = float64(q.GetValue())
		}
	}
	return fields
}

// Get Buckets  from histogram metric
func makeBucketsV1(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, b := range m.GetHistogram().Bucket {
		fields[fmt.Sprint(b.GetUpperBound())] = float64(b.GetCumulativeCount())
	}
	return fields
}

// Get name and value from metric
func getNameAndValueV1(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			//nolint:unconvert // Conversion may be needed for float64 https://github.com/mdempsky/unconvert/issues/40
			fields["gauge"] = float64(m.GetGauge().GetValue())
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			//nolint:unconvert // Conversion may be needed for float64 https://github.com/mdempsky/unconvert/issues/40
			fields["counter"] = float64(m.GetCounter().GetValue())
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			//nolint:unconvert // Conversion may be needed for float64 https://github.com/mdempsky/unconvert/issues/40
			fields["value"] = float64(m.GetUntyped().GetValue())
		}
	}
	return fields
}
