package prometheus_histogram

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
)

// PrometheusHistogramAggregator aggregates metrics into Prometheus histograms
type PrometheusHistogramAggregator struct {
	Configs  []*Config `toml:"config"`
	Registry *prometheus.Registry
}

// Config is a histogram configuration for a given measurement.
type Config struct {
	MeasurementName string    `toml:"measurement_name"`
	Unit            string    `toml:"measurement_unit"`
	Buckets         []float64 `toml:"buckets"`
}

// NewPrometheusHistogramAggregator creates a new Prometheus histogram aggregator
func NewPrometheusHistogramAggregator() telegraf.Aggregator {
	return &PrometheusHistogramAggregator{
		Registry: prometheus.NewRegistry(),
	}
}

const sampleConfig = `
  ## The period in which to flush the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Example config that aggregates a metric.
  # [[aggregators.prometheus_histogram.config]]
  #   ## The set of buckets.
  #   buckets = [0.0, 15.6, 34.5, 49.1, 71.5, 80.5, 94.5, 100.0]
  #   ## The name of metric.
  #   measurement_name = "cpu"
  #   ## Unit of the measurement
  #   measurement_unit = "seconds"
`

// SampleConfig returns a sample configuration for this plugin
func (h *PrometheusHistogramAggregator) SampleConfig() string {
	return sampleConfig
}

// Description returns the description of this plugin
func (h *PrometheusHistogramAggregator) Description() string {
	return "Aggregates metrics into Prometheus histograms."
}

// Add adds new metric to the buckets
func (h *PrometheusHistogramAggregator) Add(metric telegraf.Metric) {
	metricConfig, ok := h.metricConfig(metric.Name())
	if !ok {
		return
	}

	value, ok := metric.GetField("gauge")
	if !ok {
		return
	}

	typedValue, ok := toFloat(value)
	if !ok {
		return
	}

	histogramOpts := &prometheus.HistogramOpts{
		Name:        fmt.Sprintf("%s_%s", metricConfig.MeasurementName, metricConfig.Unit),
		Help:        "Telegraf collected metric",
		ConstLabels: metric.Tags(),
		Buckets:     metricConfig.Buckets,
	}

	histogram, err := h.getOrRegisterPrometheusHistogram(histogramOpts)
	if err != nil {
		return
	}

	histogram.Observe(typedValue)
}

func (h *PrometheusHistogramAggregator) metricConfig(metricName string) (*Config, bool) {
	for _, config := range h.Configs {
		if config.MeasurementName == metricName {
			return config, true
		}
	}

	return nil, false
}

func (h *PrometheusHistogramAggregator) getOrRegisterPrometheusHistogram(histogramOpts *prometheus.HistogramOpts) (prometheus.Histogram, error) {
	histogram := prometheus.NewHistogram(*histogramOpts)
	err := h.Registry.Register(histogram)
	if err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if h, ok := are.ExistingCollector.(prometheus.Histogram); ok {
				return h, nil
			}

			return nil, errors.New("already registered a non histogram collector with name " + histogramOpts.Name)
		}

		return nil, errors.New("unable to register histogram for " + histogramOpts.Name)
	}

	return histogram, nil
}

// Push returns histogram values for metrics
func (h *PrometheusHistogramAggregator) Push(acc telegraf.Accumulator) {
	metricFamilies, err := h.Registry.Gather()
	if err != nil {
		return
	}

	for _, metricFamily := range metricFamilies {
		if *metricFamily.Type == io_prometheus_client.MetricType_HISTOGRAM {
			for _, metric := range metricFamily.Metric {
				acc.AddHistogram(*metricFamily.Name, convertFields(metric.Histogram), copyTags(metric.Label))
			}
		}
	}
}

func convertFields(h *io_prometheus_client.Histogram) map[string]interface{} {
	fields := makeBuckets(h)
	fields["count"] = float64(h.GetSampleCount())
	fields["sum"] = float64(h.GetSampleSum())
	return fields
}

func makeBuckets(h *io_prometheus_client.Histogram) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, b := range h.Bucket {
		fields[fmt.Sprint(b.GetUpperBound())] = float64(b.GetCumulativeCount())
	}
	return fields
}

// Reset does nothing, because we need to collect counts for a long time, otherwise if config parameter 'reset' has
// small value, we will get a histogram with a small amount of the distribution.
func (h *PrometheusHistogramAggregator) Reset() {}

func toFloat(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func copyTags(labelPairs []*io_prometheus_client.LabelPair) map[string]string {
	copiedTags := map[string]string{}
	for _, labelPair := range labelPairs {
		copiedTags[*labelPair.Name] = *labelPair.Value
	}
	return copiedTags
}

func init() {
	aggregators.Add("prometheus_histogram", func() telegraf.Aggregator {
		return NewPrometheusHistogramAggregator()
	})
}
