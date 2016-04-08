package prometheus

// Parser inspired from
// https://github.com/prometheus/prom2json/blob/master/main.go

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"mime"

	"github.com/influxdata/telegraf"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// PrometheusParser is an object for Parsing incoming metrics.
type PrometheusParser struct {
	// PromFormat
	PromFormat map[string]string
	// DefaultTags will be added to every parsed metric
	//	DefaultTags map[string]string
}

// Parse returns a slice of Metrics from a text representation of a
// metrics
func (p *PrometheusParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric
	var parser expfmt.TextParser
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// Read raw data
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	// Get format
	mediatype, params, err := mime.ParseMediaType(p.PromFormat["Content-Type"])
	// Prepare output
	metricFamilies := make(map[string]*dto.MetricFamily)
	if err == nil && mediatype == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily" {
		for {
			metricFamily := &dto.MetricFamily{}
			if _, err = pbutil.ReadDelimited(reader, metricFamily); err != nil {
				if err == io.EOF {
					break
				}
				return nil, fmt.Errorf("reading metric family protocol buffer failed: %s", err)
			}
			metricFamilies[metricFamily.GetName()] = metricFamily
		}
	} else {
		metricFamilies, err = parser.TextToMetricFamilies(reader)
		if err != nil {
			return nil, fmt.Errorf("reading text format failed: %s", err)
		}
		// read metrics
		for metricName, mf := range metricFamilies {
			for _, m := range mf.Metric {
				// reading tags
				tags := makeLabels(m)
				/*
					for key, value := range p.DefaultTags {
						tags[key] = value
					}
				*/
				// reading fields
				fields := make(map[string]interface{})
				if mf.GetType() == dto.MetricType_SUMMARY {
					// summary metric
					fields = makeQuantiles(m)
					fields["count"] = float64(m.GetHistogram().GetSampleCount())
					fields["sum"] = float64(m.GetSummary().GetSampleSum())
				} else if mf.GetType() == dto.MetricType_HISTOGRAM {
					// historgram metric
					fields = makeBuckets(m)
					fields["count"] = float64(m.GetHistogram().GetSampleCount())
					fields["sum"] = float64(m.GetSummary().GetSampleSum())

				} else {
					// standard metric
					fields = getNameAndValue(m)
				}
				// converting to telegraf metric
				if len(fields) > 0 {
					metric, err := telegraf.NewMetric(metricName, tags, fields)
					if err == nil {
						metrics = append(metrics, metric)
					}
				}
			}
		}
	}
	return metrics, err
}

// Parse one line
func (p *PrometheusParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line + "\n"))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf(
			"Can not parse the line: %s, for data format: prometheus", line)
	}

	return metrics[0], nil
}

/*
// Set default tags
func (p *PrometheusParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}
*/

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

// Get labels from metric
func makeLabels(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}
	return result
}

// Get name and value from metric
func getNameAndValue(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields["gauge"] = float64(m.GetGauge().GetValue())
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields["counter"] = float64(m.GetCounter().GetValue())
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields["value"] = float64(m.GetUntyped().GetValue())
		}
	}
	return fields
}
