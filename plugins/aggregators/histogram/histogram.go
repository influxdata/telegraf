package histogram

import (
	"sort"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

// bucketTag is the tag, which contains right bucket border
const bucketTag = "le"

// bucketInf is the right bucket border for infinite values
const bucketInf = "+Inf"

// HistogramAggregator is aggregator with histogram configs and particular histograms for defined metrics
type HistogramAggregator struct {
	Configs []config `toml:"config"`

	buckets bucketsByMetrics
	cache   map[uint64]metricHistogramCollection
}

// config is the config, which contains name, field of metric and histogram buckets.
type config struct {
	Metric  string   `toml:"measurement_name"`
	Fields  []string `toml:"fields"`
	Buckets buckets  `toml:"buckets"`
}

// bucketsByMetrics contains the buckets grouped by metric and field name
type bucketsByMetrics map[string]bucketsByFields

// bucketsByFields contains the buckets grouped by field name
type bucketsByFields map[string]buckets

// buckets contains the right borders buckets
type buckets []float64

// metricHistogramCollection aggregates the histogram data
type metricHistogramCollection struct {
	histogramCollection map[string]counts
	name                string
	tags                map[string]string
}

// counts is the number of hits in the bucket
type counts []int64

// groupedByCountFields contains grouped fields by their count and fields values
type groupedByCountFields struct {
	name            string
	tags            map[string]string
	fieldsWithCount map[string]int64
}

// NewHistogramAggregator creates new histogram aggregator
func NewHistogramAggregator() telegraf.Aggregator {
	h := &HistogramAggregator{}
	h.buckets = make(bucketsByMetrics)
	h.resetCache()

	return h
}

var sampleConfig = `
  ## The period in which to flush the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Example config that aggregates all fields of the metric.
  # [[aggregators.histogram.config]]
  #   ## The set of buckets.
  #   buckets = [0.0, 15.6, 34.5, 49.1, 71.5, 80.5, 94.5, 100.0]
  #   ## The name of metric.
  #   measurement_name = "cpu"

  ## Example config that aggregates only specific fields of the metric.
  # [[aggregators.histogram.config]]
  #   ## The set of buckets.
  #   buckets = [0.0, 10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0]
  #   ## The name of metric.
  #   measurement_name = "diskio"
  #   ## The concrete fields of metric
  #   fields = ["io_time", "read_time", "write_time"]
`

// SampleConfig returns sample of config
func (h *HistogramAggregator) SampleConfig() string {
	return sampleConfig
}

// Description returns description of aggregator plugin
func (h *HistogramAggregator) Description() string {
	return "Create aggregate histograms."
}

// Add adds new hit to the buckets
func (h *HistogramAggregator) Add(in telegraf.Metric) {
	bucketsByField := make(map[string][]float64)
	for field := range in.Fields() {
		buckets := h.getBuckets(in.Name(), field)
		if buckets != nil {
			bucketsByField[field] = buckets
		}
	}

	if len(bucketsByField) == 0 {
		return
	}

	id := in.HashID()
	agr, ok := h.cache[id]
	if !ok {
		agr = metricHistogramCollection{
			name:                in.Name(),
			tags:                in.Tags(),
			histogramCollection: make(map[string]counts),
		}
	}

	for field, value := range in.Fields() {
		if buckets, ok := bucketsByField[field]; ok {
			if agr.histogramCollection[field] == nil {
				agr.histogramCollection[field] = make(counts, len(buckets)+1)
			}

			if value, ok := convert(value); ok {
				index := sort.SearchFloat64s(buckets, value)
				agr.histogramCollection[field][index]++
			}
		}
	}

	h.cache[id] = agr
}

// Push returns histogram values for metrics
func (h *HistogramAggregator) Push(acc telegraf.Accumulator) {
	metricsWithGroupedFields := []groupedByCountFields{}

	for _, aggregate := range h.cache {
		for field, counts := range aggregate.histogramCollection {
			h.groupFieldsByBuckets(&metricsWithGroupedFields, aggregate.name, field, copyTags(aggregate.tags), counts)
		}
	}

	for _, metric := range metricsWithGroupedFields {
		acc.AddFields(metric.name, makeFieldsWithCount(metric.fieldsWithCount), metric.tags)
	}
}

// groupFieldsByBuckets groups fields by metric buckets which are represented as tags
func (h *HistogramAggregator) groupFieldsByBuckets(
	metricsWithGroupedFields *[]groupedByCountFields,
	name string,
	field string,
	tags map[string]string,
	counts []int64,
) {
	count := int64(0)
	for index, bucket := range h.getBuckets(name, field) {
		count += counts[index]

		tags[bucketTag] = strconv.FormatFloat(bucket, 'f', -1, 64)
		h.groupField(metricsWithGroupedFields, name, field, count, copyTags(tags))
	}

	count += counts[len(counts)-1]
	tags[bucketTag] = bucketInf

	h.groupField(metricsWithGroupedFields, name, field, count, tags)
}

// groupField groups field by count value
func (h *HistogramAggregator) groupField(
	metricsWithGroupedFields *[]groupedByCountFields,
	name string,
	field string,
	count int64,
	tags map[string]string,
) {
	for key, metric := range *metricsWithGroupedFields {
		if name == metric.name && isTagsIdentical(tags, metric.tags) {
			(*metricsWithGroupedFields)[key].fieldsWithCount[field] = count
			return
		}
	}

	fieldsWithCount := map[string]int64{
		field: count,
	}

	*metricsWithGroupedFields = append(
		*metricsWithGroupedFields,
		groupedByCountFields{name: name, tags: tags, fieldsWithCount: fieldsWithCount},
	)
}

// Reset does nothing, because we need to collect counts for a long time, otherwise if config parameter 'reset' has
// small value, we will get a histogram with a small amount of the distribution.
func (h *HistogramAggregator) Reset() {}

// resetCache resets cached counts(hits) in the buckets
func (h *HistogramAggregator) resetCache() {
	h.cache = make(map[uint64]metricHistogramCollection)
}

// getBuckets finds buckets and returns them
func (h *HistogramAggregator) getBuckets(metric string, field string) []float64 {
	if buckets, ok := h.buckets[metric][field]; ok {
		return buckets
	}

	for _, config := range h.Configs {
		if config.Metric == metric {
			if !isBucketExists(field, config) {
				continue
			}

			if _, ok := h.buckets[metric]; !ok {
				h.buckets[metric] = make(bucketsByFields)
			}

			h.buckets[metric][field] = sortBuckets(config.Buckets)
		}
	}

	return h.buckets[metric][field]
}

// isBucketExists checks if buckets exists for the passed field
func isBucketExists(field string, cfg config) bool {
	if len(cfg.Fields) == 0 {
		return true
	}

	for _, fl := range cfg.Fields {
		if fl == field {
			return true
		}
	}

	return false
}

// sortBuckets sorts the buckets if it is needed
func sortBuckets(buckets []float64) []float64 {
	for i, bucket := range buckets {
		if i < len(buckets)-1 && bucket >= buckets[i+1] {
			sort.Float64s(buckets)
			break
		}
	}

	return buckets
}

// convert converts interface to concrete type
func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// copyTags copies tags
func copyTags(tags map[string]string) map[string]string {
	copiedTags := map[string]string{}
	for key, val := range tags {
		copiedTags[key] = val
	}

	return copiedTags
}

// isTagsIdentical checks the identity of two list of tags
func isTagsIdentical(originalTags, checkedTags map[string]string) bool {
	if len(originalTags) != len(checkedTags) {
		return false
	}

	for tagName, tagValue := range originalTags {
		if tagValue != checkedTags[tagName] {
			return false
		}
	}

	return true
}

// makeFieldsWithCount assigns count value to all metric fields
func makeFieldsWithCount(fieldsWithCountIn map[string]int64) map[string]interface{} {
	fieldsWithCountOut := map[string]interface{}{}
	for field, count := range fieldsWithCountIn {
		fieldsWithCountOut[field+"_bucket"] = count
	}

	return fieldsWithCountOut
}

// init initializes histogram aggregator plugin
func init() {
	aggregators.Add("histogram", func() telegraf.Aggregator {
		return NewHistogramAggregator()
	})
}
