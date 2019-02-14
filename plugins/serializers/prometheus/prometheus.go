package prometheus

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// SampleID uniquely identifies a Sample
type SampleID string

type serializer struct {
	ExpirationInterval internal.Duration `toml:"expiration_interval"`
	Path               string            `toml:"path"`
	CollectorsExclude  []string          `toml:"collectors_exclude"`
	StringAsLabel      bool              `toml:"string_as_label"`
	ExportTimestamp    bool              `toml:"export_timestamp"`

	sync.Mutex
	// fam is the non-expired MetricFamily by Prometheus metric name.
	fam map[string]*MetricFamily
	// now returns the current time.
	now func() time.Time
}

// Sample represents the current value of a series.
type Sample struct {
	// Labels are the Prometheus labels.
	Labels map[string]string
	// Value is the value in the Prometheus output. Only one of these will populated.
	Value          float64
	HistogramValue map[float64]uint64
	SummaryValue   map[float64]float64
	// Histograms and Summaries need a count and a sum
	Count uint64
	Sum   float64
	// Metric timestamp
	Timestamp time.Time
	// Expiration is the deadline that this Sample is valid until.
	Expiration time.Time
}

// MetricFamily contains the data required to build valid prometheus Metrics.
type MetricFamily struct {
	// Samples are the Sample belonging to this MetricFamily.
	Samples map[SampleID]*Sample
	// Need the telegraf ValueType because there isn't a Prometheus ValueType
	// representing Histogram or Summary
	TelegrafValueType telegraf.ValueType
	// LabelSet is the label counts for all Samples.
	LabelSet map[string]int
}

func NewSerializer() (*serializer, error) {
	s := &serializer{
		ExpirationInterval: internal.Duration{Duration: time.Second * 60},
		StringAsLabel:      true,
		ExportTimestamp:    true,
		fam:                make(map[string]*MetricFamily),
		now:                time.Now,
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.createObject(metric)
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var batch bytes.Buffer
	for _, metric := range metrics {
		b, err := s.createObject(metric)
		if err != nil {
			continue
		}
		batch.Write(b)
	}
	return batch.Bytes(), nil
}

func (s *serializer) createObject(metric telegraf.Metric) ([]byte, error) {
	s.Lock()
	defer s.Unlock()

	now := s.now()
	tags := metric.Tags()
	sampleID := CreateSampleID(tags)

	labels := make(map[string]string)
	for k, v := range tags {
		labels[sanitize(k)] = v
	}

	// Prometheus doesn't have a string value type, so convert string
	// fields to labels if enabled.
	if s.StringAsLabel {
		for fn, fv := range metric.Fields() {
			switch fv := fv.(type) {
			case string:
				labels[sanitize(fn)] = fv
			}
		}
	}

	switch metric.Type() {
	case telegraf.Summary:
		var mname string
		var sum float64
		var count uint64
		summaryvalue := make(map[float64]float64)
		for fn, fv := range metric.Fields() {
			var value float64
			switch fv := fv.(type) {
			case int64:
				value = float64(fv)
			case uint64:
				value = float64(fv)
			case float64:
				value = fv
			default:
				continue
			}

			switch fn {
			case "sum":
				sum = value
			case "count":
				count = uint64(value)
			default:
				limit, err := strconv.ParseFloat(fn, 64)
				if err == nil {
					summaryvalue[limit] = value
				}
			}
		}
		sample := &Sample{
			Labels:       labels,
			SummaryValue: summaryvalue,
			Count:        count,
			Sum:          sum,
			Timestamp:    metric.Time(),
			Expiration:   now.Add(s.ExpirationInterval.Duration),
		}
		mname = sanitize(metric.Name())

		s.addMetricFamily(metric, sample, mname, sampleID)

	case telegraf.Histogram:
		var mname string
		var sum float64
		var count uint64
		histogramvalue := make(map[float64]uint64)
		for fn, fv := range metric.Fields() {
			var value float64
			switch fv := fv.(type) {
			case int64:
				value = float64(fv)
			case uint64:
				value = float64(fv)
			case float64:
				value = fv
			default:
				continue
			}

			switch fn {
			case "sum":
				sum = value
			case "count":
				count = uint64(value)
			default:
				limit, err := strconv.ParseFloat(fn, 64)
				if err == nil {
					histogramvalue[limit] = uint64(value)
				}
			}
		}
		sample := &Sample{
			Labels:         labels,
			HistogramValue: histogramvalue,
			Count:          count,
			Sum:            sum,
			Timestamp:      metric.Time(),
			Expiration:     now.Add(s.ExpirationInterval.Duration),
		}
		mname = sanitize(metric.Name())

		s.addMetricFamily(metric, sample, mname, sampleID)

	default:
		for fn, fv := range metric.Fields() {
			// Ignore string and bool fields.
			var value float64
			switch fv := fv.(type) {
			case int64:
				value = float64(fv)
			case uint64:
				value = float64(fv)
			case float64:
				value = fv
			default:
				continue
			}

			sample := &Sample{
				Labels:     labels,
				Value:      value,
				Timestamp:  metric.Time(),
				Expiration: now.Add(s.ExpirationInterval.Duration),
			}

			// Special handling of value field; supports passthrough from
			// the prometheus input.
			var mname string
			switch metric.Type() {
			case telegraf.Counter:
				if fn == "counter" {
					mname = sanitize(metric.Name())
				}
			case telegraf.Gauge:
				if fn == "gauge" {
					mname = sanitize(metric.Name())
				}
			}
			if mname == "" {
				if fn == "value" {
					mname = sanitize(metric.Name())
				} else {
					mname = sanitize(fmt.Sprintf("%s_%s", metric.Name(), fn))
				}
			}

			s.addMetricFamily(metric, sample, mname, sampleID)
		}
	}

	return s.metricHandler()
}

func (s *serializer) Expire() {
	now := s.now()
	for name, family := range s.fam {
		for key, sample := range family.Samples {
			if s.ExpirationInterval.Duration != 0 && now.After(sample.Expiration) {
				for k := range sample.Labels {
					family.LabelSet[k]--
				}
				delete(family.Samples, key)

				if len(family.Samples) == 0 {
					delete(s.fam, name)
				}
			}
		}
	}
}

func (s *serializer) metricHandler() ([]byte, error) {
	s.Expire()
	var out []byte

	for name, family := range s.fam {
		if len(family.Samples) == 0 {
			log.Printf("W! There is no metric in metric family.")
			continue
		}

		dtofamily := s.convertToMetricFamily(family, name)
		o := &bytes.Buffer{}
		if _, err := expfmt.MetricFamilyToText(o, dtofamily); err != nil {
			log.Printf("E! Metric convert to text format failed: %s\n.", err.Error())
		}

		out = append(out, o.Bytes()...)
	}

	return out, nil
}

func (s *serializer) convertToMetricFamily(fam *MetricFamily, name string) *dto.MetricFamily {
	switch fam.TelegrafValueType {
	case telegraf.Counter:
		in := &dto.MetricFamily{
			Name:   proto.String(name),
			Help:   proto.String("Telegraf collected metric"),
			Type:   dto.MetricType_COUNTER.Enum(),
			Metric: getdtoMetric(fam.Samples, telegraf.Counter),
		}
		return in
	case telegraf.Gauge:
		in := &dto.MetricFamily{
			Name:   proto.String(name),
			Help:   proto.String("Telegraf collected metric"),
			Type:   dto.MetricType_GAUGE.Enum(),
			Metric: getdtoMetric(fam.Samples, telegraf.Gauge),
		}
		return in
	case telegraf.Untyped:
		in := &dto.MetricFamily{
			Name:   proto.String(name),
			Help:   proto.String("Telegraf collected metric"),
			Type:   dto.MetricType_UNTYPED.Enum(),
			Metric: getdtoMetric(fam.Samples, telegraf.Untyped),
		}
		return in
	case telegraf.Summary:
		in := &dto.MetricFamily{
			Name:   proto.String(name),
			Help:   proto.String("Telegraf collected metric"),
			Type:   dto.MetricType_SUMMARY.Enum(),
			Metric: getdtoMetric(fam.Samples, telegraf.Summary),
		}
		return in
	case telegraf.Histogram:
		in := &dto.MetricFamily{
			Name:   proto.String(name),
			Help:   proto.String("Telegraf collected metric"),
			Type:   dto.MetricType_HISTOGRAM.Enum(),
			Metric: getdtoMetric(fam.Samples, telegraf.Histogram),
		}
		return in
	}

	return nil
}

func addSample(fam *MetricFamily, sample *Sample, sampleID SampleID) {

	for k := range sample.Labels {
		fam.LabelSet[k]++
	}

	fam.Samples[sampleID] = sample
}

func (s *serializer) addMetricFamily(point telegraf.Metric, sample *Sample, mname string, sampleID SampleID) {
	var fam *MetricFamily
	var ok bool
	if fam, ok = s.fam[mname]; !ok {
		fam = &MetricFamily{
			Samples:           make(map[SampleID]*Sample),
			TelegrafValueType: point.Type(),
			LabelSet:          make(map[string]int),
		}
		s.fam[mname] = fam
	}

	addSample(fam, sample, sampleID)
	//log.Printf("family %s has samples len %d", mname, len(s.fam[mname].Samples))
}

func getdtoMetric(samples map[SampleID]*Sample, tt telegraf.ValueType) []*dto.Metric {
	var metrics []*dto.Metric
	switch tt {
	case telegraf.Counter:
		for _, sample := range samples {
			metric := &dto.Metric{
				Label: getLabels(sample.Labels),
				Counter: &dto.Counter{
					Value: proto.Float64(sample.Value),
				},
				//TimestampMs: proto.Int64(sample.Timestamp.Unix()),
			}
			metrics = append(metrics, metric)
		}

	case telegraf.Gauge:
		for _, sample := range samples {
			metric := &dto.Metric{
				Label: getLabels(sample.Labels),
				Gauge: &dto.Gauge{
					Value: proto.Float64(sample.Value),
				},
			}
			metrics = append(metrics, metric)
		}

	case telegraf.Untyped:
		for _, sample := range samples {
			metric := &dto.Metric{
				Label: getLabels(sample.Labels),
				Untyped: &dto.Untyped{
					Value: proto.Float64(sample.Value),
				},
			}
			metrics = append(metrics, metric)
		}

	case telegraf.Summary:
		for _, sample := range samples {
			metric := &dto.Metric{
				Summary: &dto.Summary{
					SampleCount: proto.Uint64(sample.Count),
					SampleSum:   proto.Float64(sample.Sum),
					Quantile:    getSummaryQuantile(sample.SummaryValue),
				},
			}
			metrics = append(metrics, metric)
		}
	case telegraf.Histogram:
		for _, sample := range samples {
			metric := &dto.Metric{
				Histogram: &dto.Histogram{
					SampleCount: proto.Uint64(sample.Count),
					SampleSum:   proto.Float64(sample.Sum),
					Bucket:      getHistogramBucket(sample.HistogramValue),
				},
			}
			metrics = append(metrics, metric)
		}
	}

	return metrics
}

func getHistogramBucket(histogramValue map[float64]uint64) []*dto.Bucket {
	var la []*dto.Bucket
	for q, v := range histogramValue {
		qu := &dto.Bucket{
			UpperBound:      proto.Float64(q),
			CumulativeCount: proto.Uint64(v),
		}
		la = append(la, qu)
	}

	return la
}

func getSummaryQuantile(summaryValue map[float64]float64) []*dto.Quantile {
	var la []*dto.Quantile
	for q, v := range summaryValue {
		qu := &dto.Quantile{
			Quantile: proto.Float64(q),
			Value:    proto.Float64(v),
		}
		la = append(la, qu)
	}

	return la
}

func getLabels(labels map[string]string) []*dto.LabelPair {
	var la []*dto.LabelPair
	for name, value := range labels {
		label := &dto.LabelPair{
			Name:  proto.String(name),
			Value: proto.String(value),
		}
		la = append(la, label)
	}

	return la
}

func sanitize(value string) string {
	return invalidNameCharRE.ReplaceAllString(value, "_")
}

// CreateSampleID creates a SampleID based on the tags of a telegraf.Metric.
func CreateSampleID(tags map[string]string) SampleID {
	pairs := make([]string, 0, len(tags))
	for k, v := range tags {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return SampleID(strings.Join(pairs, ","))
}
