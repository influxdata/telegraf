package prometheus_client

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

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
	// Expiration is the deadline that this Sample is valid until.
	Expiration time.Time
}

// MetricFamily contains the data required to build valid prometheus Metrics.
type MetricFamily struct {
	// Samples are the Sample belonging to this MetricFamily.
	Samples map[SampleID]*Sample
	// Need the telegraf ValueType because there isn't a Prometheus ValueType
	// representing Histogram or Summary
	ValueType telegraf.ValueType
	// LabelSet is the label counts for all Samples.
	LabelSet map[string]int
	// This is the description for the MetricFamily
	Description string
}

type PrometheusClient struct {
	Listen             string
	TLSCert            string            `toml:"tls_cert"`
	TLSKey             string            `toml:"tls_key"`
	BasicUsername      string            `toml:"basic_username"`
	BasicPassword      string            `toml:"basic_password"`
	ExpirationInterval internal.Duration `toml:"expiration_interval"`
	Path               string            `toml:"path"`
	CollectorsExclude  []string          `toml:"collectors_exclude"`
	StringAsLabel      bool              `toml:"string_as_label"`

	server *http.Server

	sync.Mutex
	// fam is the non-expired MetricFamily by Prometheus metric name.
	fam map[string]*MetricFamily
	// now returns the current time.
	now func() time.Time
}

var sampleConfig = `
  ## Address to listen on
  # listen = ":9273"

  ## Use TLS
  #tls_cert = "/etc/ssl/telegraf.crt"
  #tls_key = "/etc/ssl/telegraf.key"

  ## Use http basic authentication
  #basic_username = "Foo"
  #basic_password = "Bar"

  ## Interval to expire metrics and not deliver to prometheus, 0 == no expiration
  # expiration_interval = "60s"

  ## Collectors to enable, valid entries are "gocollector" and "process".
  ## If unset, both are enabled.
  collectors_exclude = ["gocollector", "process"]

  # Send string metrics as Prometheus labels.
  # Unless set to false all string metrics will be sent as labels.
  string_as_label = true
`

func (p *PrometheusClient) basicAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.BasicUsername != "" && p.BasicPassword != "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

			username, password, ok := r.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(username), []byte(p.BasicUsername)) != 1 ||
				subtle.ConstantTimeCompare([]byte(password), []byte(p.BasicPassword)) != 1 {
				http.Error(w, "Not authorized", 401)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}

func (p *PrometheusClient) Start() error {
	defaultCollectors := map[string]bool{
		"gocollector": true,
		"process":     true,
	}
	for _, collector := range p.CollectorsExclude {
		delete(defaultCollectors, collector)
	}

	registry := prometheus.NewRegistry()
	for collector, _ := range defaultCollectors {
		switch collector {
		case "gocollector":
			registry.Register(prometheus.NewGoCollector())
		case "process":
			registry.Register(prometheus.NewProcessCollector(os.Getpid(), ""))
		default:
			return fmt.Errorf("unrecognized collector %s", collector)
		}
	}

	registry.Register(p)

	if p.Listen == "" {
		p.Listen = "localhost:9273"
	}

	if p.Path == "" {
		p.Path = "/metrics"
	}

	mux := http.NewServeMux()
	mux.Handle(p.Path, p.basicAuth(promhttp.HandlerFor(
		registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError})))

	p.server = &http.Server{
		Addr:    p.Listen,
		Handler: mux,
	}

	go func() {
		var err error
		if p.TLSCert != "" && p.TLSKey != "" {
			err = p.server.ListenAndServeTLS(p.TLSCert, p.TLSKey)
		} else {
			err = p.server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Printf("E! Error creating prometheus metric endpoint, err: %s\n",
				err.Error())
		}
	}()

	return nil
}

func (p *PrometheusClient) Stop() {
	// plugin gets cleaned up in Close() already.
}

func (p *PrometheusClient) Connect() error {
	// This service output does not need to make any further connections
	return nil
}

func (p *PrometheusClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := p.server.Shutdown(ctx)
	prometheus.Unregister(p)
	return err
}

func (p *PrometheusClient) SampleConfig() string {
	return sampleConfig
}

func (p *PrometheusClient) Description() string {
	return "Configuration for the Prometheus client to spawn"
}

// Implements prometheus.Collector
func (p *PrometheusClient) Describe(ch chan<- *prometheus.Desc) {
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

// Expire removes Samples that have expired.
func (p *PrometheusClient) Expire() {
	now := p.now()
	for name, family := range p.fam {
		for key, sample := range family.Samples {
			if p.ExpirationInterval.Duration != 0 && now.After(sample.Expiration) {
				for k, _ := range sample.Labels {
					family.LabelSet[k]--
				}
				delete(family.Samples, key)

				if len(family.Samples) == 0 {
					delete(p.fam, name)
				}
			}
		}
	}
}

// Collect implements prometheus.Collector
func (p *PrometheusClient) Collect(ch chan<- prometheus.Metric) {
	p.Lock()
	defer p.Unlock()

	p.Expire()

	for name, family := range p.fam {
		// Get list of all labels on MetricFamily
		var labelNames []string
		for k, v := range family.LabelSet {
			if v > 0 {
				labelNames = append(labelNames, k)
			}
		}
		desc := prometheus.NewDesc(name, family.Description, labelNames, nil)

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
			switch family.ValueType {
			case telegraf.Summary:
				metric, err = prometheus.NewConstSummary(desc, sample.Count, sample.Sum, sample.SummaryValue, labels...)
			case telegraf.Histogram:
				metric, err = prometheus.NewConstHistogram(desc, sample.Count, sample.Sum, sample.HistogramValue, labels...)
			default:
				metric, err = prometheus.NewConstMetric(desc, getPromValueType(family.ValueType), sample.Value, labels...)
			}
			if err != nil {
				log.Printf("E! Error creating prometheus metric, "+
					"key: %s, labels: %v,\nerr: %s\n",
					name, labels, err.Error())
			}

			ch <- metric
		}
	}
}

func sanitize(value string) string {
	return invalidNameCharRE.ReplaceAllString(value, "_")
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

// Prometheus values are float64 and strings are unsupported
func getPromValue(value interface{}) (float64, bool) {
	switch value := value.(type) {
	case int64:
		return float64(value), true
	case uint64:
		return float64(value), true
	case float64:
		return value, true
	default:
		return 0, false
	}
}

func (p *PrometheusClient) addMetricFamily(metricName string, sampleID SampleID, sample *Sample, description string, valueType telegraf.ValueType) {
	fam, ok := p.fam[metricName]
	if !ok {
		fam = &MetricFamily{
			Samples:     make(map[SampleID]*Sample),
			ValueType:   valueType,
			LabelSet:    make(map[string]int),
			Description: description,
		}
		p.fam[metricName] = fam
	}

	for k := range sample.Labels {
		fam.LabelSet[k]++
	}

	fam.Samples[sampleID] = sample
}

func (p *PrometheusClient) addSummaryFamily(point telegraf.Metric, description string) {
	var sum float64
	var count uint64
	sumValue := make(map[float64]float64)

	for fn, fv := range point.Fields() {
		value, ok := getPromValue(fv)
		if !ok {
			continue
		}

		switch fn {
		case "sum":
			sum = value
		case "count":
			count = uint64(value)
		default:
			quantile, err := strconv.ParseFloat(fn, 64)
			if err == nil {
				sumValue[quantile] = value
			}
		}
	}
	metricName := sanitize(point.Name())
	sampleID := CreateSampleID(point.Tags())
	sample := &Sample{
		Labels:       makeLabels(point, p.StringAsLabel),
		SummaryValue: sumValue,
		Count:        count,
		Sum:          sum,
		Expiration:   p.now().Add(p.ExpirationInterval.Duration),
	}

	p.addMetricFamily(metricName, sampleID, sample, description, telegraf.Summary)
}

func (p *PrometheusClient) addHistogramFamily(point telegraf.Metric, description string) {
	var sum float64
	var count uint64
	histValue := make(map[float64]uint64)

	for fn, fv := range point.Fields() {
		value, ok := getPromValue(fv)
		if !ok {
			continue
		}

		switch fn {
		case "sum":
			sum = value
		case "count":
			count = uint64(value)
		default:
			bucket, err := strconv.ParseFloat(fn, 64)
			if err == nil {
				histValue[bucket] = uint64(value)
			}
		}
	}
	metricName := sanitize(point.Name())
	sampleID := CreateSampleID(point.Tags())
	sample := &Sample{
		Labels:         makeLabels(point, p.StringAsLabel),
		HistogramValue: histValue,
		Count:          count,
		Sum:            sum,
		Expiration:     p.now().Add(p.ExpirationInterval.Duration),
	}

	p.addMetricFamily(metricName, sampleID, sample, description, telegraf.Histogram)
}

func (p *PrometheusClient) addValueFamily(point telegraf.Metric, description string, valueType telegraf.ValueType) {
	var metricName string
	sampleID := CreateSampleID(point.Tags())
	for fn, fv := range point.Fields() {
		// Ignore string and bool fields.
		value, ok := getPromValue(fv)
		if !ok {
			continue
		}

		sample := &Sample{
			Labels:     makeLabels(point, p.StringAsLabel),
			Value:      value,
			Expiration: p.now().Add(p.ExpirationInterval.Duration),
		}

		// In case of non-prometheus generated metric with type append the fn to the metric
		metricName = sanitize(fmt.Sprintf("%s_%s", point.Name(), fn))
		switch valueType {
		case telegraf.Counter:
			// Do not append generic fn counter to the metric name
			if fn == "counter" || fn == "value" {
				metricName = sanitize(point.Name())
			}
		case telegraf.Gauge:
			// Do not append generic fn gauge to the metric name
			if fn == "gauge" || fn == "value" {
				metricName = sanitize(point.Name())
			}
		default:
			// Do not append generic fn value to the metric name
			if fn == "value" {
				metricName = sanitize(point.Name())
			}
		}

		p.addMetricFamily(metricName, sampleID, sample, description, valueType)
	}
}

func makeLabels(point telegraf.Metric, sal bool) map[string]string {
	labels := make(map[string]string)
	for k, v := range point.Tags() {
		// These tags are used only internally
		if k != "prometheus_type" && k != "prometheus_help" {
			labels[sanitize(k)] = v
		}
	}

	// Prometheus doesn't have a string value type, so convert string
	// fields to labels if enabled.
	if sal {
		for fn, fv := range point.Fields() {
			switch fv := fv.(type) {
			case string:
				labels[sanitize(fn)] = fv
			}
		}
	}

	return labels
}

func (p *PrometheusClient) Write(metrics []telegraf.Metric) error {
	p.Lock()
	defer p.Unlock()

	for _, point := range metrics {

		desc, ok := point.GetTag("prometheus_help")
		if !ok {
			desc = "Telegraf collected metric"
		}

		tag, ok := point.GetTag("prometheus_type")
		if ok {
			switch tag {
			case "COUNTER":
				p.addValueFamily(point, desc, telegraf.Counter)
			case "GAUGE":
				p.addValueFamily(point, desc, telegraf.Gauge)
			case "HISTOGRAM":
				p.addHistogramFamily(point, desc)
			case "SUMMARY":
				p.addSummaryFamily(point, desc)
			default:
				p.addValueFamily(point, desc, telegraf.Untyped)
			}
		} else {
			switch point.Type() {
			case telegraf.Counter:
				p.addValueFamily(point, desc, telegraf.Counter)
			case telegraf.Gauge:
				p.addValueFamily(point, desc, telegraf.Gauge)
			case telegraf.Histogram:
				p.addHistogramFamily(point, desc)
			case telegraf.Summary:
				p.addSummaryFamily(point, desc)
			default:
				p.addValueFamily(point, desc, telegraf.Untyped)
			}
		}
	}
	return nil
}

func init() {
	outputs.Add("prometheus_client", func() telegraf.Output {
		return &PrometheusClient{
			ExpirationInterval: internal.Duration{Duration: time.Second * 60},
			StringAsLabel:      true,
			fam:                make(map[string]*MetricFamily),
			now:                time.Now,
		}
	})
}
