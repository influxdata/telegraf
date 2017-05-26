package prometheus_client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/prometheus/client_golang/prometheus"
)

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

type MetricWithExpiration struct {
	Metric     prometheus.Metric
	Expiration time.Time
}

type PrometheusClient struct {
	Listen             string
	ExpirationInterval internal.Duration `toml:"expiration_interval"`
	server             *http.Server

	metrics map[string]*MetricWithExpiration

	sync.Mutex
}

var sampleConfig = `
  ## Address to listen on
  # listen = ":9126"

  ## Interval to expire metrics and not deliver to prometheus, 0 == no expiration
  # expiration_interval = "60s"
`

func (p *PrometheusClient) Start() error {
	p.metrics = make(map[string]*MetricWithExpiration)
	prometheus.Register(p)

	if p.Listen == "" {
		p.Listen = "localhost:9126"
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", prometheus.Handler())

	p.server = &http.Server{
		Addr:    p.Listen,
		Handler: mux,
	}

	go p.server.ListenAndServe()
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
	return p.server.Shutdown(ctx)
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

// Implements prometheus.Collector
func (p *PrometheusClient) Collect(ch chan<- prometheus.Metric) {
	p.Lock()
	defer p.Unlock()

	for key, m := range p.metrics {
		if p.ExpirationInterval.Duration != 0 && time.Now().After(m.Expiration) {
			delete(p.metrics, key)
		} else {
			ch <- m.Metric
		}
	}
}

func sanitize(value string) string {
	return invalidNameCharRE.ReplaceAllString(value, "_")
}

func (p *PrometheusClient) Write(metrics []telegraf.Metric) error {
	p.Lock()
	defer p.Unlock()

	if len(metrics) == 0 {
		return nil
	}

	// Since each metric must have a consistent set of labels, we collect the
	// tag keys for each measurement so that we can ensure they are all set.
	tagsByName := make(map[string][]string)
	for _, point := range metrics {
		tags := []string{}
		for tagKey, _ := range point.Tags() {
			if len(tagKey) == 0 {
				continue
			}
			tags = append(tags, tagKey)
		}
		name := point.Name()
		tagsByName[name] = tags
	}

	for _, point := range metrics {
		name := point.Name()

		// Convert tags into prometheus labels
		l := prometheus.Labels{}
		tagSet := point.Tags()
		tagKeys := tagsByName[name]
		for _, tagKey := range tagKeys {
			tagValue, ok := tagSet[tagKey]
			if !ok {
				// Tags missing from this measurement are labeled with a empty
				// string value.
				tagValue = ""
			}
			l[sanitize(tagKey)] = tagValue
		}

		// Get a type if it's available, defaulting to Untyped
		var mType prometheus.ValueType
		switch point.Type() {
		case telegraf.Counter:
			mType = prometheus.CounterValue
		case telegraf.Gauge:
			mType = prometheus.GaugeValue
		default:
			mType = prometheus.UntypedValue
		}

		for fieldKey, val := range point.Fields() {
			// Ignore string and bool fields.
			switch val.(type) {
			case string:
				continue
			case bool:
				continue
			}

			// sanitize the field key
			fieldKey = sanitize(fieldKey)
			mname := sanitize(name)
			if fieldKey != "value" {
				mname = fmt.Sprintf("%s_%s", mname, fieldKey)
			}

			desc := prometheus.NewDesc(mname, "Telegraf collected metric", nil, l)
			var metric prometheus.Metric
			var err error

			// switch for field type
			switch val := val.(type) {
			case int64:
				metric, err = prometheus.NewConstMetric(desc, mType, float64(val))
			case float64:
				metric, err = prometheus.NewConstMetric(desc, mType, val)
			default:
				continue
			}
			if err != nil {
				log.Printf("E! Error creating prometheus metric, "+
					"key: %s, labels: %v,\nerr: %s\n",
					mname, l, err.Error())
			}

			p.metrics[desc.String()] = &MetricWithExpiration{
				Metric:     metric,
				Expiration: time.Now().Add(p.ExpirationInterval.Duration),
			}
		}
	}
	return nil
}

func init() {
	outputs.Add("prometheus_client", func() telegraf.Output {
		return &PrometheusClient{
			ExpirationInterval: internal.Duration{Duration: time.Second * 60},
		}
	})
}
