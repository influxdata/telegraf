package prometheus

// Parser inspired from
// https://github.com/prometheus/prom2json/blob/master/main.go

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// Parse returns a slice of Metrics from a text representation of a
// metrics
func Parse(buf []byte, header http.Header) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// Read raw data
	reader := bytes.NewReader(buf)

	format := expfmt.ResponseFormat(header)
	decoder := expfmt.NewDecoder(reader, format)

	// read metrics
	family := &dto.MetricFamily{}
	for {
		err := decoder.Decode(family)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decoding prometheus exposition format failed: %s", err)
		}

		for _, m := range family.Metric {
			// reading tags
			tags := makeLabels(m)
			tags["prometheus_type"] = family.GetType().String()
			tags["prometheus_help"] = family.GetHelp()
			// reading fields
			fields := make(map[string]interface{})
			switch family.GetType() {
			case dto.MetricType_SUMMARY:
				// summary metric
				fields = makeQuantiles(m.GetSummary())
				fields["count"] = float64(m.GetSummary().GetSampleCount())
				fields["sum"] = float64(m.GetSummary().GetSampleSum())
			case dto.MetricType_HISTOGRAM:
				// histogram metric
				fields = makeBuckets(m.GetHistogram())
				fields["count"] = float64(m.GetHistogram().GetSampleCount())
				fields["sum"] = float64(m.GetHistogram().GetSampleSum())
			case dto.MetricType_COUNTER:
				// counter metric
				// counter is allways at least 0
				fields["counter"] = float64(m.GetCounter().GetValue())
			case dto.MetricType_GAUGE:
				// gauge metric
				// gauge can be unset, returning NaN
				if !math.IsNaN(m.GetGauge().GetValue()) {
					fields["gauge"] = float64(m.GetGauge().GetValue())
				}
			default:
				// untyped metric
				if !math.IsNaN(m.GetUntyped().GetValue()) {
					fields["value"] = float64(m.GetUntyped().GetValue())
				}
			}
			// converting to telegraf metric
			if len(fields) > 0 {
				var t time.Time
				if m.TimestampMs != nil && *m.TimestampMs > 0 {
					t = time.Unix(0, *m.TimestampMs*1000000)
				} else {
					t = time.Now()
				}
				metric, err := metric.New(family.GetName(), tags, fields, t, valueType(family.GetType()))
				if err == nil {
					metrics = append(metrics, metric)
				}
			}
		}
	}

	return metrics, nil
}

func valueType(mt dto.MetricType) telegraf.ValueType {
	switch mt {
	case dto.MetricType_COUNTER:
		return telegraf.Counter
	case dto.MetricType_GAUGE:
		return telegraf.Gauge
	case dto.MetricType_SUMMARY:
		return telegraf.Summary
	case dto.MetricType_HISTOGRAM:
		return telegraf.Histogram
	default:
		return telegraf.Untyped
	}
}

// Get Quantiles from summary metric
func makeQuantiles(s *dto.Summary) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, q := range s.Quantile {
		// with no events to process summary returns NaN
		if !math.IsNaN(q.GetValue()) {
			fields[fmt.Sprint(q.GetQuantile())] = float64(q.GetValue())
		}
	}
	return fields
}

// Get Buckets  from histogram metric
func makeBuckets(h *dto.Histogram) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, b := range h.Bucket {
		fields[fmt.Sprint(b.GetUpperBound())] = float64(b.GetCumulativeCount())
	}
	return fields
}

// Get labels from metric
func makeLabels(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}
	return result
}
