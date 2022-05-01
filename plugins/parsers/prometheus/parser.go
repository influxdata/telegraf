package prometheus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"time"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/prometheus/common"
)

type Parser struct {
	DefaultTags     map[string]string
	Header          http.Header
	IgnoreTimestamp bool
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var parser expfmt.TextParser
	var metrics []telegraf.Metric
	var err error
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// Read raw data
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	// Prepare output
	metricFamilies := make(map[string]*dto.MetricFamily)
	mediatype, params, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
	if err == nil && mediatype == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily" {
		for {
			mf := &dto.MetricFamily{}
			if _, ierr := pbutil.ReadDelimited(reader, mf); ierr != nil {
				if ierr == io.EOF {
					break
				}
				return nil, fmt.Errorf("reading metric family protocol buffer failed: %s", ierr)
			}
			metricFamilies[mf.GetName()] = mf
		}
	} else {
		metricFamilies, err = parser.TextToMetricFamilies(reader)
		if err != nil {
			return nil, fmt.Errorf("reading text format failed: %s", err)
		}
	}

	now := time.Now()

	// read metrics
	for metricName, mf := range metricFamilies {
		for _, m := range mf.Metric {
			// reading tags
			tags := common.MakeLabels(m, p.DefaultTags)
			t := p.GetTimestamp(m, now)

			if mf.GetType() == dto.MetricType_SUMMARY {
				// summary metric
				telegrafMetrics := makeQuantiles(m, tags, metricName, mf.GetType(), t)
				metrics = append(metrics, telegrafMetrics...)
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				// histogram metric
				telegrafMetrics := makeBuckets(m, tags, metricName, mf.GetType(), t)
				metrics = append(metrics, telegrafMetrics...)
			} else {
				// standard metric
				// reading fields
				fields := getNameAndValue(m, metricName)
				// converting to telegraf metric
				if len(fields) > 0 {
					m := metric.New("prometheus", tags, fields, t, common.ValueType(mf.GetType()))
					metrics = append(metrics, m)
				}
			}
		}
	}

	return metrics, err
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("no metrics in line")
	}

	if len(metrics) > 1 {
		return nil, fmt.Errorf("more than one metric in line")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

// Get Quantiles for summary metric & Buckets for histogram
func makeQuantiles(m *dto.Metric, tags map[string]string, metricName string, metricType dto.MetricType, t time.Time) []telegraf.Metric {
	var metrics []telegraf.Metric
	fields := make(map[string]interface{})

	fields[metricName+"_count"] = float64(m.GetSummary().GetSampleCount())
	fields[metricName+"_sum"] = m.GetSummary().GetSampleSum()
	met := metric.New("prometheus", tags, fields, t, common.ValueType(metricType))
	metrics = append(metrics, met)

	for _, q := range m.GetSummary().Quantile {
		newTags := tags
		fields = make(map[string]interface{})

		newTags["quantile"] = fmt.Sprint(q.GetQuantile())
		fields[metricName] = q.GetValue()

		quantileMetric := metric.New("prometheus", newTags, fields, t, common.ValueType(metricType))
		metrics = append(metrics, quantileMetric)
	}
	return metrics
}

// Get Buckets  from histogram metric
func makeBuckets(m *dto.Metric, tags map[string]string, metricName string, metricType dto.MetricType, t time.Time) []telegraf.Metric {
	var metrics []telegraf.Metric
	fields := make(map[string]interface{})

	fields[metricName+"_count"] = float64(m.GetHistogram().GetSampleCount())
	fields[metricName+"_sum"] = m.GetHistogram().GetSampleSum()

	met := metric.New("prometheus", tags, fields, t, common.ValueType(metricType))
	metrics = append(metrics, met)

	for _, b := range m.GetHistogram().Bucket {
		newTags := tags
		fields = make(map[string]interface{})
		newTags["le"] = fmt.Sprint(b.GetUpperBound())
		fields[metricName+"_bucket"] = float64(b.GetCumulativeCount())

		histogramMetric := metric.New("prometheus", newTags, fields, t, common.ValueType(metricType))
		metrics = append(metrics, histogramMetric)
	}
	return metrics
}

// Get name and value from metric
func getNameAndValue(m *dto.Metric, metricName string) map[string]interface{} {
	fields := make(map[string]interface{})
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields[metricName] = m.GetGauge().GetValue()
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			fields[metricName] = m.GetCounter().GetValue()
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			fields[metricName] = m.GetUntyped().GetValue()
		}
	}
	return fields
}

func (p *Parser) GetTimestamp(m *dto.Metric, now time.Time) time.Time {
	var t time.Time
	if !p.IgnoreTimestamp && m.TimestampMs != nil && *m.TimestampMs > 0 {
		t = time.Unix(0, m.GetTimestampMs()*1000000)
	} else {
		t = now
	}
	return t
}
