package prometheusremotewrite

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/prometheus"
)

type MetricKey uint64

// MetricSortOrder controls if the output is sorted.
type MetricSortOrder int

const (
	NoSortMetrics MetricSortOrder = iota
	SortMetrics
)

// StringHandling defines how to process string fields.
type StringHandling int

const (
	DiscardStrings StringHandling = iota
	StringAsLabel
)

type FormatConfig struct {
	MetricSortOrder MetricSortOrder
	StringHandling  StringHandling
}

type Serializer struct {
	config FormatConfig
}

func NewSerializer(config FormatConfig) (*Serializer, error) {
	s := &Serializer{config: config}
	return s, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.SerializeBatch([]telegraf.Metric{metric})
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var buf bytes.Buffer
	var entries = make(map[MetricKey]prompb.TimeSeries)
	for _, metric := range metrics {
		commonLabels := s.createLabels(metric)
		var metrickey MetricKey
		var promts prompb.TimeSeries
		for _, field := range metric.FieldList() {
			metricName := prometheus.MetricName(metric.Name(), field.Key, metric.Type())
			metricName, ok := prometheus.SanitizeMetricName(metricName)
			if !ok {
				continue
			}
			switch metric.Type() {
			case telegraf.Counter:
				fallthrough
			case telegraf.Gauge:
				fallthrough
			case telegraf.Untyped:
				value, ok := prometheus.SampleValue(field.Value)
				if !ok {
					continue
				}
				metrickey, promts = getPromTS(metricName, commonLabels, value, metric.Time())
			case telegraf.Histogram:
				switch {
				case strings.HasSuffix(field.Key, "_bucket"):
					// if bucket only, init sum, count, inf
					metrickeysum, promtssum := getPromTS(fmt.Sprintf("%s_sum", metricName), commonLabels, float64(0), metric.Time())
					if _, ok = entries[metrickeysum]; !ok {
						entries[metrickeysum] = promtssum
					}
					metrickeycount, promtscount := getPromTS(fmt.Sprintf("%s_count", metricName), commonLabels, float64(0), metric.Time())
					if _, ok = entries[metrickeycount]; !ok {
						entries[metrickeycount] = promtscount
					}
					labels := make([]prompb.Label, len(commonLabels), len(commonLabels)+1)
					copy(labels, commonLabels)
					labels = append(labels, prompb.Label{
						Name:  "le",
						Value: "+Inf",
					})
					metrickeyinf, promtsinf := getPromTS(fmt.Sprintf("%s_bucket", metricName), labels, float64(0), metric.Time())
					if _, ok = entries[metrickeyinf]; !ok {
						entries[metrickeyinf] = promtsinf
					}

					le, ok := metric.GetTag("le")
					if !ok {
						continue
					}
					bound, err := strconv.ParseFloat(le, 64)
					if err != nil {
						continue
					}
					count, ok := prometheus.SampleCount(field.Value)
					if !ok {
						continue
					}

					labels = make([]prompb.Label, len(commonLabels), len(commonLabels)+1)
					copy(labels, commonLabels)
					labels = append(labels, prompb.Label{
						Name:  "le",
						Value: fmt.Sprint(bound),
					})
					metrickey, promts = getPromTS(fmt.Sprintf("%s_bucket", metricName), labels, float64(count), metric.Time())
				case strings.HasSuffix(field.Key, "_sum"):
					sum, ok := prometheus.SampleSum(field.Value)
					if !ok {
						continue
					}

					metrickey, promts = getPromTS(fmt.Sprintf("%s_sum", metricName), commonLabels, sum, metric.Time())
				case strings.HasSuffix(field.Key, "_count"):
					count, ok := prometheus.SampleCount(field.Value)
					if !ok {
						continue
					}

					// if no bucket generate +Inf entry
					labels := make([]prompb.Label, len(commonLabels), len(commonLabels)+1)
					copy(labels, commonLabels)
					labels = append(labels, prompb.Label{
						Name:  "le",
						Value: "+Inf",
					})
					metrickeyinf, promtsinf := getPromTS(fmt.Sprintf("%s_bucket", metricName), labels, float64(count), metric.Time())
					if minf, ok := entries[metrickeyinf]; !ok || minf.Samples[0].Value == 0 {
						entries[metrickeyinf] = promtsinf
					}

					metrickey, promts = getPromTS(fmt.Sprintf("%s_count", metricName), commonLabels, float64(count), metric.Time())
				default:
					continue
				}
			case telegraf.Summary:
				switch {
				case strings.HasSuffix(field.Key, "_sum"):
					sum, ok := prometheus.SampleSum(field.Value)
					if !ok {
						continue
					}

					metrickey, promts = getPromTS(fmt.Sprintf("%s_sum", metricName), commonLabels, sum, metric.Time())
				case strings.HasSuffix(field.Key, "_count"):
					count, ok := prometheus.SampleCount(field.Value)
					if !ok {
						continue
					}

					metrickey, promts = getPromTS(fmt.Sprintf("%s_count", metricName), commonLabels, float64(count), metric.Time())
				default:
					quantileTag, ok := metric.GetTag("quantile")
					if !ok {
						continue
					}
					quantile, err := strconv.ParseFloat(quantileTag, 64)
					if err != nil {
						continue
					}
					value, ok := prometheus.SampleValue(field.Value)
					if !ok {
						continue
					}

					labels := make([]prompb.Label, len(commonLabels), len(commonLabels)+1)
					copy(labels, commonLabels)
					labels = append(labels, prompb.Label{
						Name:  "quantile",
						Value: fmt.Sprint(quantile),
					})
					metrickey, promts = getPromTS(metricName, labels, value, metric.Time())
				}
			default:
				return nil, fmt.Errorf("unknown type %v", metric.Type())
			}

			// A batch of metrics can contain multiple values for a single
			// Prometheus sample.  If this metric is older than the existing
			// sample then we can skip over it.
			m, ok := entries[metrickey]
			if ok {
				if metric.Time().Before(time.Unix(0, m.Samples[0].Timestamp*1_000_000)) {
					continue
				}
			}
			entries[metrickey] = promts
		}
	}

	var promTS = make([]prompb.TimeSeries, len(entries))
	var i int
	for _, promts := range entries {
		promTS[i] = promts
		i++
	}

	if s.config.MetricSortOrder == SortMetrics {
		sort.Slice(promTS, func(i, j int) bool {
			lhs := promTS[i].Labels
			rhs := promTS[j].Labels
			if len(lhs) != len(rhs) {
				return len(lhs) < len(rhs)
			}

			for index := range lhs {
				l := lhs[index]
				r := rhs[index]

				if l.Name != r.Name {
					return l.Name < r.Name
				}

				if l.Value != r.Value {
					return l.Value < r.Value
				}
			}

			return false
		})
	}
	pb := &prompb.WriteRequest{Timeseries: promTS}
	data, err := pb.Marshal()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal protobuf: %v", err)
	}
	encoded := snappy.Encode(nil, data)
	buf.Write(encoded) //nolint:revive // from buffer.go: "err is always nil"
	return buf.Bytes(), nil
}

func hasLabel(name string, labels []prompb.Label) bool {
	for _, label := range labels {
		if name == label.Name {
			return true
		}
	}
	return false
}

func (s *Serializer) createLabels(metric telegraf.Metric) []prompb.Label {
	labels := make([]prompb.Label, 0, len(metric.TagList()))
	for _, tag := range metric.TagList() {
		// Ignore special tags for histogram and summary types.
		switch metric.Type() {
		case telegraf.Histogram:
			if tag.Key == "le" {
				continue
			}
		case telegraf.Summary:
			if tag.Key == "quantile" {
				continue
			}
		}

		name, ok := prometheus.SanitizeLabelName(tag.Key)
		if !ok {
			continue
		}

		// remove tags with empty values
		if tag.Value == "" {
			continue
		}

		labels = append(labels, prompb.Label{Name: name, Value: tag.Value})
	}

	if s.config.StringHandling != StringAsLabel {
		return labels
	}

	addedFieldLabel := false
	for _, field := range metric.FieldList() {
		value, ok := field.Value.(string)
		if !ok {
			continue
		}

		name, ok := prometheus.SanitizeLabelName(field.Key)
		if !ok {
			continue
		}

		// If there is a tag with the same name as the string field, discard
		// the field and use the tag instead.
		if hasLabel(name, labels) {
			continue
		}

		labels = append(labels, prompb.Label{Name: name, Value: value})
		addedFieldLabel = true
	}

	if addedFieldLabel {
		sort.Slice(labels, func(i, j int) bool {
			return labels[i].Name < labels[j].Name
		})
	}

	return labels
}

func MakeMetricKey(labels []prompb.Label) MetricKey {
	h := fnv.New64a()
	for _, label := range labels {
		h.Write([]byte(label.Name))  //nolint:revive // from hash.go: "It never returns an error"
		h.Write([]byte("\x00"))      //nolint:revive // from hash.go: "It never returns an error"
		h.Write([]byte(label.Value)) //nolint:revive // from hash.go: "It never returns an error"
		h.Write([]byte("\x00"))      //nolint:revive // from hash.go: "It never returns an error"
	}
	return MetricKey(h.Sum64())
}

func getPromTS(name string, labels []prompb.Label, value float64, ts time.Time) (MetricKey, prompb.TimeSeries) {
	sample := []prompb.Sample{{
		// Timestamp is int milliseconds for remote write.
		Timestamp: ts.UnixNano() / int64(time.Millisecond),
		Value:     value,
	}}
	labelscopy := make([]prompb.Label, len(labels), len(labels)+1)
	copy(labelscopy, labels)
	labels = append(labelscopy, prompb.Label{
		Name:  "__name__",
		Value: name,
	})
	return MakeMetricKey(labels), prompb.TimeSeries{Labels: labels, Samples: sample}
}
