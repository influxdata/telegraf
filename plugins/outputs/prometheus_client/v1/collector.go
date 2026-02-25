package v1

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/influxdata/telegraf"
	serializers_prometheus "github.com/influxdata/telegraf/plugins/serializers/prometheus"
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
	TypeMapping        serializers_prometheus.MetricTypes
	Log                telegraf.Logger
	NameSanitization   string

	sync.Mutex
	fam          map[string]*MetricFamily
	expireTicker *time.Ticker
}

func NewCollector(
	expire time.Duration,
	stringsAsLabel, exportTimestamp bool,
	typeMapping serializers_prometheus.MetricTypes,
	log telegraf.Logger,
	nameSanitization string,
) *Collector {
	c := &Collector{
		ExpirationInterval: expire,
		StringAsLabel:      stringsAsLabel,
		ExportTimestamp:    exportTimestamp,
		TypeMapping:        typeMapping,
		Log:                log,
		NameSanitization:   nameSanitization,
		fam:                make(map[string]*MetricFamily),
	}

	if c.ExpirationInterval != 0 {
		c.expireTicker = time.NewTicker(c.ExpirationInterval)
		go func() {
			for {
				<-c.expireTicker.C
				c.Expire(time.Now())
			}
		}()
	}

	return c
}

func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Expire metrics, doing this on Collect ensure metrics are removed even if no
	// new metrics are added to the output.
	if c.ExpirationInterval != 0 {
		c.Expire(time.Now())
	}

	c.Lock()
	defer c.Unlock()

	for name, family := range c.fam {
		// Get list of all labels on metricFamily
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
		pointType := c.TypeMapping.DetermineType(mname, point)
		fam = &MetricFamily{
			Samples:           make(map[SampleID]*Sample),
			TelegrafValueType: pointType,
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
	c.addMetrics(metrics)

	// Expire metrics, doing this on Add ensure metrics are removed even if no
	// new metrics are added to the output.
	if c.ExpirationInterval != 0 {
		c.Expire(time.Now())
	}

	return nil
}

func (c *Collector) addMetrics(metrics []telegraf.Metric) {
	c.Lock()
	defer c.Unlock()

	now := time.Now()

	for _, point := range sorted(metrics) {
		tags := point.Tags()
		sampleID := CreateSampleID(tags)

		labels := make(map[string]string)
		for k, v := range tags {
			name, ok := c.sanitizeLabelName(k)
			if !ok {
				continue
			}
			labels[name] = v
		}

		// Prometheus doesn't have a string value type, so convert string
		// fields to labels if enabled.
		if c.StringAsLabel {
			for fn, fv := range point.Fields() {
				sfv, ok := fv.(string)
				if !ok {
					continue
				}

				name, ok := c.sanitizeLabelName(fn)
				if !ok {
					continue
				}
				labels[name] = sfv
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
			mname, ok := c.sanitizeMetricName(point.Name())
			if !ok {
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
			mname, ok := c.sanitizeMetricName(point.Name())
			if !ok {
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
						mname = point.Name()
					}
				case telegraf.Gauge:
					if fn == "gauge" {
						mname = point.Name()
					}
				}
				if mname == "" {
					if fn == "value" {
						mname = point.Name()
					} else {
						mname = fmt.Sprintf("%s_%s", point.Name(), fn)
					}
				}

				mname, ok := c.sanitizeMetricName(mname)
				if !ok {
					continue
				}
				c.addMetricFamily(point, sample, mname, sampleID)
			}
		}
	}
}

func (c *Collector) Expire(now time.Time) {
	c.Lock()
	defer c.Unlock()

	for name, family := range c.fam {
		for key, sample := range family.Samples {
			if now.After(sample.Expiration) {
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

func (c *Collector) sanitizeMetricName(name string) (string, bool) {
	return serializers_prometheus.SanitizeMetricNameByEncoding(name, c.NameSanitization)
}

func (c *Collector) sanitizeLabelName(name string) (string, bool) {
	return serializers_prometheus.SanitizeLabelNameByEncoding(name, c.NameSanitization)
}
