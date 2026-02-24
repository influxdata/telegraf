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

type metricFamily struct {
	name string
	typ  telegraf.ValueType
}

type metric struct {
	labels    []labelPair
	time      time.Time
	addTime   time.Time
	scaler    *scaler
	histogram *histogram
	summary   *summary
}

type labelPair struct {
	name  string
	value string
}

type scaler struct {
	value float64
}

type bucket struct {
	bound float64
	count uint64
}

type quantile struct {
	quantile float64
	value    float64
}

type histogram struct {
	buckets []bucket
	count   uint64
	sum     float64
}

func (h *histogram) merge(b bucket) {
	for i := range h.buckets {
		if h.buckets[i].bound == b.bound {
			h.buckets[i].count = b.count
			return
		}
	}
	h.buckets = append(h.buckets, b)
}

type summary struct {
	quantiles []quantile
	count     uint64
	sum       float64
}

func (s *summary) merge(q quantile) {
	for i := range s.quantiles {
		if s.quantiles[i].quantile == q.quantile {
			s.quantiles[i].value = q.value
			return
		}
	}
	s.quantiles = append(s.quantiles, q)
}

type metricKey uint64

func makeMetricKey(labels []labelPair) metricKey {
	h := fnv.New64a()
	for _, label := range labels {
		h.Write([]byte(label.name))
		h.Write([]byte("\x00"))
		h.Write([]byte(label.value))
		h.Write([]byte("\x00"))
	}
	return metricKey(h.Sum64())
}

type entry struct {
	family  metricFamily
	metrics map[metricKey]*metric
}

// Collection is a cache of metrics that are being processed.
type Collection struct {
	entries map[metricFamily]entry
	config  FormatConfig
}

// NewCollection creates a new Collection instance.
func NewCollection(config FormatConfig) *Collection {
	cache := &Collection{
		entries: make(map[metricFamily]entry),
		config:  config,
	}
	return cache
}

func (c *Collection) sanitizeMetricName(name string) (string, bool) {
	return SanitizeMetricNameByEncoding(name, c.config.NameSanitization)
}

func (c *Collection) sanitizeLabelName(name string) (string, bool) {
	return SanitizeLabelNameByEncoding(name, c.config.NameSanitization)
}

func hasLabel(name string, labels []labelPair) bool {
	for _, label := range labels {
		if name == label.name {
			return true
		}
	}
	return false
}

func (c *Collection) createLabels(metric telegraf.Metric) []labelPair {
	labels := make([]labelPair, 0, len(metric.TagList()))
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

		name, ok := c.sanitizeLabelName(tag.Key)
		if !ok {
			continue
		}

		labels = append(labels, labelPair{name: name, value: tag.Value})
	}

	if !c.config.StringAsLabel {
		return labels
	}

	addedFieldLabel := false
	for _, field := range metric.FieldList() {
		value, ok := field.Value.(string)
		if !ok {
			continue
		}

		name, ok := c.sanitizeLabelName(field.Key)
		if !ok {
			continue
		}

		// If there is a tag with the same name as the string field, discard
		// the field and use the tag instead.
		if hasLabel(name, labels) {
			continue
		}

		labels = append(labels, labelPair{name: name, value: value})
		addedFieldLabel = true
	}

	if addedFieldLabel {
		sort.Slice(labels, func(i, j int) bool {
			return labels[i].name < labels[j].name
		})
	}

	return labels
}

// Add adds a metric to the collection. It will create a new entry if the metric is not already present.
func (c *Collection) Add(m telegraf.Metric, now time.Time) {
	labels := c.createLabels(m)
	for _, field := range m.FieldList() {
		metricName := MetricName(m.Name(), field.Key, m.Type())
		metricName, ok := c.sanitizeMetricName(metricName)
		if !ok {
			continue
		}
		metricType := c.config.TypeMappings.DetermineType(metricName, m)

		family := metricFamily{
			name: metricName,
			typ:  metricType,
		}

		singleEntry, ok := c.entries[family]
		if !ok {
			singleEntry = entry{
				family:  family,
				metrics: make(map[metricKey]*metric),
			}
			c.entries[family] = singleEntry
		}

		metricKey := makeMetricKey(labels)

		existingMetric, ok := singleEntry.metrics[metricKey]
		if ok {
			// A batch of metrics can contain multiple values for a single
			// Prometheus sample.  If this metric is older than the existing
			// sample then we can skip over it.
			if m.Time().Before(existingMetric.time) {
				continue
			}
		}

		switch m.Type() {
		case telegraf.Counter:
			fallthrough
		case telegraf.Gauge:
			fallthrough
		case telegraf.Untyped:
			value, ok := SampleValue(field.Value)
			if !ok {
				continue
			}

			existingMetric = &metric{
				labels:  labels,
				time:    m.Time(),
				addTime: now,
				scaler:  &scaler{value: value},
			}

			singleEntry.metrics[metricKey] = existingMetric
		case telegraf.Histogram:
			if existingMetric == nil {
				existingMetric = &metric{
					labels:    labels,
					time:      m.Time(),
					addTime:   now,
					histogram: &histogram{},
				}
			} else {
				existingMetric.time = m.Time()
				existingMetric.addTime = now
			}
			switch {
			case strings.HasSuffix(field.Key, "_bucket"):
				le, ok := m.GetTag("le")
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

				existingMetric.histogram.merge(bucket{
					bound: bound,
					count: count,
				})
			case strings.HasSuffix(field.Key, "_sum"):
				sum, ok := SampleSum(field.Value)
				if !ok {
					continue
				}

				existingMetric.histogram.sum = sum
			case strings.HasSuffix(field.Key, "_count"):
				count, ok := SampleCount(field.Value)
				if !ok {
					continue
				}

				existingMetric.histogram.count = count
			default:
				continue
			}

			singleEntry.metrics[metricKey] = existingMetric
		case telegraf.Summary:
			if existingMetric == nil {
				existingMetric = &metric{
					labels:  labels,
					time:    m.Time(),
					addTime: now,
					summary: &summary{},
				}
			} else {
				existingMetric.time = m.Time()
				existingMetric.addTime = now
			}
			switch {
			case strings.HasSuffix(field.Key, "_sum"):
				sum, ok := SampleSum(field.Value)
				if !ok {
					continue
				}

				existingMetric.summary.sum = sum
			case strings.HasSuffix(field.Key, "_count"):
				count, ok := SampleCount(field.Value)
				if !ok {
					continue
				}

				existingMetric.summary.count = count
			default:
				quantileTag, ok := m.GetTag("quantile")
				if !ok {
					continue
				}
				singleQuantile, err := strconv.ParseFloat(quantileTag, 64)
				if err != nil {
					continue
				}

				value, ok := SampleValue(field.Value)
				if !ok {
					continue
				}

				existingMetric.summary.merge(quantile{
					quantile: singleQuantile,
					value:    value,
				})
			}

			singleEntry.metrics[metricKey] = existingMetric
		}
	}
}

// Expire removes metrics that are older than the specified age.
func (c *Collection) Expire(now time.Time, age time.Duration) {
	expireTime := now.Add(-age)
	for _, entry := range c.entries {
		for key, metric := range entry.metrics {
			if metric.addTime.Before(expireTime) {
				delete(entry.metrics, key)
				if len(entry.metrics) == 0 {
					delete(c.entries, entry.family)
				}
			}
		}
	}
}

// GetEntries returns a slice of all entries in the collection.
func (c *Collection) GetEntries() []entry {
	entries := make([]entry, 0, len(c.entries))
	for _, entry := range c.entries {
		entries = append(entries, entry)
	}

	if c.config.SortMetrics {
		sort.Slice(entries, func(i, j int) bool {
			lhs := entries[i].family
			rhs := entries[j].family
			if lhs.name != rhs.name {
				return lhs.name < rhs.name
			}

			return lhs.typ < rhs.typ
		})
	}
	return entries
}

// GetMetrics returns a slice of all metrics in the entry.
func (c *Collection) GetMetrics(entry entry) []*metric {
	metrics := make([]*metric, 0, len(entry.metrics))
	for _, metric := range entry.metrics {
		metrics = append(metrics, metric)
	}

	if c.config.SortMetrics {
		sort.Slice(metrics, func(i, j int) bool {
			lhs := metrics[i].labels
			rhs := metrics[j].labels
			if len(lhs) != len(rhs) {
				return len(lhs) < len(rhs)
			}

			for index := range lhs {
				l := lhs[index]
				r := rhs[index]

				if l.name != r.name {
					return l.name < r.name
				}

				if l.value != r.value {
					return l.value < r.value
				}
			}

			return false
		})
	}

	return metrics
}

// GetProto returns a slice of all metrics in the collection as protobuf messages.
func (c *Collection) GetProto() []*dto.MetricFamily {
	result := make([]*dto.MetricFamily, 0, len(c.entries))

	for _, entry := range c.GetEntries() {
		mf := &dto.MetricFamily{
			Name: proto.String(entry.family.name),
			Type: metricType(entry.family.typ),
		}

		if !c.config.CompactEncoding {
			mf.Help = proto.String(helpString)
		}

		for _, metric := range c.GetMetrics(entry) {
			l := make([]*dto.LabelPair, 0, len(metric.labels))
			for _, label := range metric.labels {
				l = append(l, &dto.LabelPair{
					Name:  proto.String(label.name),
					Value: proto.String(label.value),
				})
			}

			m := &dto.Metric{
				Label: l,
			}

			if c.config.ExportTimestamp {
				m.TimestampMs = proto.Int64(metric.time.UnixNano() / int64(time.Millisecond))
			}

			switch entry.family.typ {
			case telegraf.Gauge:
				m.Gauge = &dto.Gauge{Value: proto.Float64(metric.scaler.value)}
			case telegraf.Counter:
				m.Counter = &dto.Counter{Value: proto.Float64(metric.scaler.value)}
			case telegraf.Untyped:
				m.Untyped = &dto.Untyped{Value: proto.Float64(metric.scaler.value)}
			case telegraf.Histogram:
				buckets := make([]*dto.Bucket, 0, len(metric.histogram.buckets))
				for _, bucket := range metric.histogram.buckets {
					buckets = append(buckets, &dto.Bucket{
						UpperBound:      proto.Float64(bucket.bound),
						CumulativeCount: proto.Uint64(bucket.count),
					})
				}

				m.Histogram = &dto.Histogram{
					Bucket:      buckets,
					SampleCount: proto.Uint64(metric.histogram.count),
					SampleSum:   proto.Float64(metric.histogram.sum),
				}
			case telegraf.Summary:
				quantiles := make([]*dto.Quantile, 0, len(metric.summary.quantiles))
				for _, quantile := range metric.summary.quantiles {
					quantiles = append(quantiles, &dto.Quantile{
						Quantile: proto.Float64(quantile.quantile),
						Value:    proto.Float64(quantile.value),
					})
				}

				m.Summary = &dto.Summary{
					Quantile:    quantiles,
					SampleCount: proto.Uint64(metric.summary.count),
					SampleSum:   proto.Float64(metric.summary.sum),
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
