package v1

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	serializer "github.com/influxdata/telegraf/plugins/serializers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_:]`)
	validNameCharRE   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)
)

// SampleID uniquely identifies a Sample
type SampleID string

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

type Collector struct {
	ExpirationInterval time.Duration
	StringAsLabel      bool
	ExportTimestamp    bool
	Log                telegraf.Logger

	sync.Mutex
	fam map[string]*MetricFamily
}

func NewCollector(expire time.Duration, stringsAsLabel bool, logger telegraf.Logger) *Collector {
	return &Collector{
		ExpirationInterval: expire,
		StringAsLabel:      stringsAsLabel,
		Log:                logger,
		fam:                make(map[string]*MetricFamily),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Lock()
	defer c.Unlock()

	c.Expire(time.Now(), c.ExpirationInterval)

	for name, family := range c.fam {
		// Get list of all labels on MetricFamily
		var labelNames []string
		for k, v := range family.LabelSet {
			if v > 0 {
				labelNames = append(labelNames, k)
			}
		}
		desc := prometheus.NewDesc(name, "Telegraf collected metric", labelNames, nil)

		for _, sample := range family.Samples {
			// Get labels for this sample; unset labels will be set to the
			// empty string
			var labels []string
			for _, label := range labelNames {
				v := sample.Labels[label]
				labels = append(labels, v)
			}

			var metric prometheus.Metric
			var err error
			switch family.TelegrafValueType {
			case telegraf.Summary:
				metric, err = prometheus.NewConstSummary(desc, sample.Count, sample.Sum, sample.SummaryValue, labels...)
			case telegraf.Histogram:
				metric, err = prometheus.NewConstHistogram(desc, sample.Count, sample.Sum, sample.HistogramValue, labels...)
			default:
				metric, err = prometheus.NewConstMetric(desc, getPromValueType(family.TelegrafValueType), sample.Value, labels...)
			}
			if err != nil {
				c.Log.Errorf("Error creating prometheus metric: "+
					"key: %s, labels: %v, err: %v",
					name, labels, err)
				continue
			}

			if c.ExportTimestamp {
				metric = prometheus.NewMetricWithTimestamp(sample.Timestamp, metric)
			}
			ch <- metric
		}
	}
}

func sanitize(value string) string {
	return invalidNameCharRE.ReplaceAllString(value, "_")
}

func isValidTagName(tag string) bool {
	return validNameCharRE.MatchString(tag)
}

func getPromValueType(tt telegraf.ValueType) prometheus.ValueType {
	switch tt {
	case telegraf.Counter:
		return prometheus.CounterValue
	case telegraf.Gauge:
		return prometheus.GaugeValue
	default:
		return prometheus.UntypedValue
	}
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

func addSample(fam *MetricFamily, sample *Sample, sampleID SampleID) {
	for k := range sample.Labels {
		fam.LabelSet[k]++
	}

	fam.Samples[sampleID] = sample
}

func (c *Collector) addMetricFamily(point telegraf.Metric, sample *Sample, mname string, sampleID SampleID) {
	var fam *MetricFamily
	var ok bool
	if fam, ok = c.fam[mname]; !ok {
		fam = &MetricFamily{
			Samples:           make(map[SampleID]*Sample),
			TelegrafValueType: point.Type(),
			LabelSet:          make(map[string]int),
		}
		c.fam[mname] = fam
	}

	addSample(fam, sample, sampleID)
}

// Sorted returns a copy of the metrics in time ascending order.  A copy is
// made to avoid modifying the input metric slice since doing so is not
// allowed.
func sorted(metrics []telegraf.Metric) []telegraf.Metric {
	batch := make([]telegraf.Metric, 0, len(metrics))
	for i := len(metrics) - 1; i >= 0; i-- {
		batch = append(batch, metrics[i])
	}
	sort.Slice(batch, func(i, j int) bool {
		return batch[i].Time().Before(batch[j].Time())
	})
	return batch
}

func (c *Collector) Add(metrics []telegraf.Metric) error {
	c.Lock()
	defer c.Unlock()

	now := time.Now()

	for _, point := range sorted(metrics) {
		tags := point.Tags()
		sampleID := CreateSampleID(tags)

		labels := make(map[string]string)
		for k, v := range tags {
			name, ok := serializer.SanitizeLabelName(k)
			if !ok {
				continue
			}
			labels[name] = v
		}

		// Prometheus doesn't have a string value type, so convert string
		// fields to labels if enabled.
		if c.StringAsLabel {
			for fn, fv := range point.Fields() {
				switch fv := fv.(type) {
				case string:
					name, ok := serializer.SanitizeLabelName(fn)
					if !ok {
						continue
					}
					labels[name] = fv
				}
			}
		}

		switch point.Type() {
		case telegraf.Summary:
			var mname string
			var sum float64
			var count uint64
			summaryvalue := make(map[float64]float64)
			for fn, fv := range point.Fields() {
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
				Timestamp:    point.Time(),
				Expiration:   now.Add(c.ExpirationInterval),
			}
			mname = sanitize(point.Name())

			if !isValidTagName(mname) {
				continue
			}

			c.addMetricFamily(point, sample, mname, sampleID)

		case telegraf.Histogram:
			var mname string
			var sum float64
			var count uint64
			histogramvalue := make(map[float64]uint64)
			for fn, fv := range point.Fields() {
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
				Timestamp:      point.Time(),
				Expiration:     now.Add(c.ExpirationInterval),
			}
			mname = sanitize(point.Name())

			if !isValidTagName(mname) {
				continue
			}

			c.addMetricFamily(point, sample, mname, sampleID)

		default:
			for fn, fv := range point.Fields() {
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
					Timestamp:  point.Time(),
					Expiration: now.Add(c.ExpirationInterval),
				}

				// Special handling of value field; supports passthrough from
				// the prometheus input.
				var mname string
				switch point.Type() {
				case telegraf.Counter:
					if fn == "counter" {
						mname = sanitize(point.Name())
					}
				case telegraf.Gauge:
					if fn == "gauge" {
						mname = sanitize(point.Name())
					}
				}
				if mname == "" {
					if fn == "value" {
						mname = sanitize(point.Name())
					} else {
						mname = sanitize(fmt.Sprintf("%s_%s", point.Name(), fn))
					}
				}
				if !isValidTagName(mname) {
					continue
				}
				c.addMetricFamily(point, sample, mname, sampleID)
			}
		}
	}
	return nil
}

func (c *Collector) Expire(now time.Time, age time.Duration) {
	if age == 0 {
		return
	}

	for name, family := range c.fam {
		for key, sample := range family.Samples {
			if age != 0 && now.After(sample.Expiration) {
				for k := range sample.Labels {
					family.LabelSet[k]--
				}
				delete(family.Samples, key)

				if len(family.Samples) == 0 {
					delete(c.fam, name)
				}
			}
		}
	}
}
