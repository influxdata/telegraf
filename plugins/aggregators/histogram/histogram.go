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
	Metric  string   `toml:"metric_name"`
	Fields  []string `toml:"metric_fields"`
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

// NewHistogramAggregator creates new histogram aggregator
func NewHistogramAggregator() telegraf.Aggregator {
	h := &HistogramAggregator{}
	h.buckets = make(bucketsByMetrics)
	h.resetCache()

	return h
}

var sampleConfig = `
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## The example of config to aggregate histogram for all fields of specified metric.
  [[aggregators.histogram.config]]
  ## The set of buckets.
  buckets = [0.0, 15.6, 34.5, 49.1, 71.5, 80.5, 94.5, 100.0]
  ## The name of metric.
  metric_name = "cpu"

  ## The example of config to aggregate for specified fields of metric.
  [[aggregators.histogram.config]]
  ## The set of buckets.
  buckets = [0.0, 10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0]
  ## The name of metric.
  metric_name = "diskio"
  ## The concrete fields of metric
  metric_fields = ["io_time", "read_time", "write_time"]
`

// SampleConfig returns sample of config
func (h *HistogramAggregator) SampleConfig() string {
	return sampleConfig
}

// Description returns description of aggregator plugin
func (h *HistogramAggregator) Description() string {
	return "Keep the aggregate histogram of each metric passing through."
}

// Add adds new hit to the buckets
func (h *HistogramAggregator) Add(in telegraf.Metric) {
	var bucketsByField = make(map[string][]float64)
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
	for _, aggregate := range h.cache {
		for field, counts := range aggregate.histogramCollection {

			buckets := h.getBuckets(aggregate.name, field)
			count := int64(0)

			for index, bucket := range buckets {
				count += counts[index]
				addFields(acc, aggregate, field, strconv.FormatFloat(bucket, 'f', 1, 64), count)
			}

			// the adding a value to the infinitive bucket
			count += counts[len(counts)-1]
			addFields(acc, aggregate, field, bucketInf, count)
		}
	}
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

// addFields adds the field with specified tags to accumulator
func addFields(acc telegraf.Accumulator, agr metricHistogramCollection, field string, bucketTagVal string, count int64) {
	fields := map[string]interface{}{field + "_bucket": count}

	tags := map[string]string{}
	for key, val := range agr.tags {
		tags[key] = val
	}
	tags[bucketTag] = bucketTagVal

	acc.AddFields(agr.name, fields, tags)
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

// init initializes histogram aggregator plugin
func init() {
	aggregators.Add("histogram", func() telegraf.Aggregator {
		return NewHistogramAggregator()
	})
}
