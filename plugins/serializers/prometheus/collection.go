package prometheus

import (
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
)

const helpString = "Telegraf collected metric"

type TimeFunc func() time.Time

type MetricFamily struct {
	Name string
	Type telegraf.ValueType
}

type Metric struct {
	Labels    []LabelPair
	Time      time.Time
	AddTime   time.Time
	Scaler    *Scaler
	Histogram *Histogram
	Summary   *Summary
}

type LabelPair struct {
	Name  string
	Value string
}

type Scaler struct {
	Value float64
}

type Bucket struct {
	Bound float64
	Count uint64
}

type Quantile struct {
	Quantile float64
	Value    float64
}

type Histogram struct {
	Buckets []Bucket
	Count   uint64
	Sum     float64
}

func (h *Histogram) merge(b Bucket) {
	for i := range h.Buckets {
		if h.Buckets[i].Bound == b.Bound {
			h.Buckets[i].Count = b.Count
			return
		}
	}
	h.Buckets = append(h.Buckets, b)
}

type Summary struct {
	Quantiles []Quantile
	Count     uint64
	Sum       float64
}

func (s *Summary) merge(q Quantile) {
	for i := range s.Quantiles {
		if s.Quantiles[i].Quantile == q.Quantile {
			s.Quantiles[i].Value = q.Value
			return
		}
	}
	s.Quantiles = append(s.Quantiles, q)
}

type MetricKey uint64

func MakeMetricKey(labels []LabelPair) MetricKey {
	h := fnv.New64a()
	for _, label := range labels {
		h.Write([]byte(label.Name))  //nolint:revive // from hash.go: "It never returns an error"
		h.Write([]byte("\x00"))      //nolint:revive // from hash.go: "It never returns an error"
		h.Write([]byte(label.Value)) //nolint:revive // from hash.go: "It never returns an error"
		h.Write([]byte("\x00"))      //nolint:revive // from hash.go: "It never returns an error"
	}
	return MetricKey(h.Sum64())
}

type Entry struct {
	Family  MetricFamily
	Metrics map[MetricKey]*Metric
}

type Collection struct {
	Entries map[MetricFamily]Entry
	config  FormatConfig
}

func NewCollection(config FormatConfig) *Collection {
	cache := &Collection{
		Entries: make(map[MetricFamily]Entry),
		config:  config,
	}
	return cache
}

func hasLabel(name string, labels []LabelPair) bool {
	for _, label := range labels {
		if name == label.Name {
			return true
		}
	}
	return false
}

func (c *Collection) createLabels(metric telegraf.Metric) []LabelPair {
	labels := make([]LabelPair, 0, len(metric.TagList()))
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

		name, ok := SanitizeLabelName(tag.Key)
		if !ok {
			continue
		}

		labels = append(labels, LabelPair{Name: name, Value: tag.Value})
	}

	if c.config.StringHandling != StringAsLabel {
		return labels
	}

	addedFieldLabel := false
	for _, field := range metric.FieldList() {
		value, ok := field.Value.(string)
		if !ok {
			continue
		}

		name, ok := SanitizeLabelName(field.Key)
		if !ok {
			continue
		}

		// If there is a tag with the same name as the string field, discard
		// the field and use the tag instead.
		if hasLabel(name, labels) {
			continue
		}

		labels = append(labels, LabelPair{Name: name, Value: value})
		addedFieldLabel = true
	}

	if addedFieldLabel {
		sort.Slice(labels, func(i, j int) bool {
			return labels[i].Name < labels[j].Name
		})
	}

	return labels
}

func (c *Collection) Add(metric telegraf.Metric, now time.Time) {
	labels := c.createLabels(metric)
	for _, field := range metric.FieldList() {
		metricName := MetricName(metric.Name(), field.Key, metric.Type())
		metricName, ok := SanitizeMetricName(metricName)
		if !ok {
			continue
		}

		family := MetricFamily{
			Name: metricName,
			Type: metric.Type(),
		}

		entry, ok := c.Entries[family]
		if !ok {
			entry = Entry{
				Family:  family,
				Metrics: make(map[MetricKey]*Metric),
			}
			c.Entries[family] = entry
		}

		metricKey := MakeMetricKey(labels)

		m, ok := entry.Metrics[metricKey]
		if ok {
			// A batch of metrics can contain multiple values for a single
			// Prometheus sample.  If this metric is older than the existing
			// sample then we can skip over it.
			if metric.Time().Before(m.Time) {
				continue
			}
		}

		switch metric.Type() {
		case telegraf.Counter:
			fallthrough
		case telegraf.Gauge:
			fallthrough
		case telegraf.Untyped:
			value, ok := SampleValue(field.Value)
			if !ok {
				continue
			}

			m = &Metric{
				Labels:  labels,
				Time:    metric.Time(),
				AddTime: now,
				Scaler:  &Scaler{Value: value},
			}

			entry.Metrics[metricKey] = m
		case telegraf.Histogram:
			if m == nil {
				m = &Metric{
					Labels:    labels,
					Time:      metric.Time(),
					AddTime:   now,
					Histogram: &Histogram{},
				}
			} else {
				m.Time = metric.Time()
				m.AddTime = now
			}
			switch {
			case strings.HasSuffix(field.Key, "_bucket"):
				le, ok := metric.GetTag("le")
				if !ok {
					continue
				}
				bound, err := strconv.ParseFloat(le, 64)
				if err != nil {
					continue
				}

				count, ok := SampleCount(field.Value)
				if !ok {
					continue
				}

				m.Histogram.merge(Bucket{
					Bound: bound,
					Count: count,
				})
			case strings.HasSuffix(field.Key, "_sum"):
				sum, ok := SampleSum(field.Value)
				if !ok {
					continue
				}

				m.Histogram.Sum = sum
			case strings.HasSuffix(field.Key, "_count"):
				count, ok := SampleCount(field.Value)
				if !ok {
					continue
				}

				m.Histogram.Count = count
			default:
				continue
			}

			entry.Metrics[metricKey] = m
		case telegraf.Summary:
			if m == nil {
				m = &Metric{
					Labels:  labels,
					Time:    metric.Time(),
					AddTime: now,
					Summary: &Summary{},
				}
			} else {
				m.Time = metric.Time()
				m.AddTime = now
			}
			switch {
			case strings.HasSuffix(field.Key, "_sum"):
				sum, ok := SampleSum(field.Value)
				if !ok {
					continue
				}

				m.Summary.Sum = sum
			case strings.HasSuffix(field.Key, "_count"):
				count, ok := SampleCount(field.Value)
				if !ok {
					continue
				}

				m.Summary.Count = count
			default:
				quantileTag, ok := metric.GetTag("quantile")
				if !ok {
					continue
				}
				quantile, err := strconv.ParseFloat(quantileTag, 64)
				if err != nil {
					continue
				}

				value, ok := SampleValue(field.Value)
				if !ok {
					continue
				}

				m.Summary.merge(Quantile{
					Quantile: quantile,
					Value:    value,
				})
			}

			entry.Metrics[metricKey] = m
		}
	}
}

func (c *Collection) Expire(now time.Time, age time.Duration) {
	expireTime := now.Add(-age)
	for _, entry := range c.Entries {
		for key, metric := range entry.Metrics {
			if metric.AddTime.Before(expireTime) {
				delete(entry.Metrics, key)
				if len(entry.Metrics) == 0 {
					delete(c.Entries, entry.Family)
				}
			}
		}
	}
}

func (c *Collection) GetEntries(order MetricSortOrder) []Entry {
	entries := make([]Entry, 0, len(c.Entries))
	for _, entry := range c.Entries {
		entries = append(entries, entry)
	}

	if order == SortMetrics {
		sort.Slice(entries, func(i, j int) bool {
			lhs := entries[i].Family
			rhs := entries[j].Family
			if lhs.Name != rhs.Name {
				return lhs.Name < rhs.Name
			}

			return lhs.Type < rhs.Type
		})
	}
	return entries
}

func (c *Collection) GetMetrics(entry Entry, order MetricSortOrder) []*Metric {
	metrics := make([]*Metric, 0, len(entry.Metrics))
	for _, metric := range entry.Metrics {
		metrics = append(metrics, metric)
	}

	if order == SortMetrics {
		sort.Slice(metrics, func(i, j int) bool {
			lhs := metrics[i].Labels
			rhs := metrics[j].Labels
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

	return metrics
}

func (c *Collection) GetProto() []*dto.MetricFamily {
	result := make([]*dto.MetricFamily, 0, len(c.Entries))

	for _, entry := range c.GetEntries(c.config.MetricSortOrder) {
		mf := &dto.MetricFamily{
			Name: proto.String(entry.Family.Name),
			Help: proto.String(helpString),
			Type: MetricType(entry.Family.Type),
		}

		for _, metric := range c.GetMetrics(entry, c.config.MetricSortOrder) {
			l := make([]*dto.LabelPair, 0, len(metric.Labels))
			for _, label := range metric.Labels {
				l = append(l, &dto.LabelPair{
					Name:  proto.String(label.Name),
					Value: proto.String(label.Value),
				})
			}

			m := &dto.Metric{
				Label: l,
			}

			if c.config.TimestampExport == ExportTimestamp {
				m.TimestampMs = proto.Int64(metric.Time.UnixNano() / int64(time.Millisecond))
			}

			switch entry.Family.Type {
			case telegraf.Gauge:
				m.Gauge = &dto.Gauge{Value: proto.Float64(metric.Scaler.Value)}
			case telegraf.Counter:
				m.Counter = &dto.Counter{Value: proto.Float64(metric.Scaler.Value)}
			case telegraf.Untyped:
				m.Untyped = &dto.Untyped{Value: proto.Float64(metric.Scaler.Value)}
			case telegraf.Histogram:
				buckets := make([]*dto.Bucket, 0, len(metric.Histogram.Buckets))
				for _, bucket := range metric.Histogram.Buckets {
					buckets = append(buckets, &dto.Bucket{
						UpperBound:      proto.Float64(bucket.Bound),
						CumulativeCount: proto.Uint64(bucket.Count),
					})
				}

				m.Histogram = &dto.Histogram{
					Bucket:      buckets,
					SampleCount: proto.Uint64(metric.Histogram.Count),
					SampleSum:   proto.Float64(metric.Histogram.Sum),
				}
			case telegraf.Summary:
				quantiles := make([]*dto.Quantile, 0, len(metric.Summary.Quantiles))
				for _, quantile := range metric.Summary.Quantiles {
					quantiles = append(quantiles, &dto.Quantile{
						Quantile: proto.Float64(quantile.Quantile),
						Value:    proto.Float64(quantile.Value),
					})
				}

				m.Summary = &dto.Summary{
					Quantile:    quantiles,
					SampleCount: proto.Uint64(metric.Summary.Count),
					SampleSum:   proto.Float64(metric.Summary.Sum),
				}
			default:
				panic("unknown telegraf.ValueType")
			}

			mf.Metric = append(mf.Metric, m)
		}

		if len(mf.Metric) != 0 {
			result = append(result, mf)
		}
	}

	return result
}
