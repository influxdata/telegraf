package prometheus_client

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/prometheus/client_golang/prometheus"
)

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

type PrometheusClient struct {
	Listen string

	metrics map[string]prometheus.Metric

	sync.Mutex
}

var sampleConfig = `
  ## Address to listen on
  # listen = ":9126"
`

func (p *PrometheusClient) Start() error {
	prometheus.MustRegister(p)
	defer func() {
		if r := recover(); r != nil {
			// recovering from panic here because there is no way to stop a
			// running http go server except by a kill signal. Since the server
			// does not stop on SIGHUP, Start() will panic when the process
			// is reloaded.
		}
	}()
	if p.Listen == "" {
		p.Listen = "localhost:9126"
	}

	http.Handle("/metrics", prometheus.Handler())
	server := &http.Server{
		Addr: p.Listen,
	}

	go server.ListenAndServe()
	return nil
}

func (p *PrometheusClient) Stop() {
	// TODO: Use a listener for http.Server that counts active connections
	//       that can be stopped and closed gracefully
}

func (p *PrometheusClient) Connect() error {
	// This service output does not need to make any further connections
	return nil
}

func (p *PrometheusClient) Close() error {
	// This service output does not need to close any of its connections
	return nil
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

	for _, m := range p.metrics {
		ch <- m
	}
}

func (p *PrometheusClient) Write(metrics []telegraf.Metric) error {
	p.Lock()
	defer p.Unlock()

	p.metrics = make(map[string]prometheus.Metric)

	if len(metrics) == 0 {
		return nil
	}

	for _, point := range metrics {
		key := point.Name()
		key = invalidNameCharRE.ReplaceAllString(key, "_")

		// convert tags into prometheus labels
		var labels []string
		l := prometheus.Labels{}
		for k, v := range point.Tags() {
			k = invalidNameCharRE.ReplaceAllString(k, "_")
			if len(k) == 0 {
				continue
			}
			labels = append(labels, k)
			l[k] = v
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

		for n, val := range point.Fields() {
			// Ignore string and bool fields.
			switch val.(type) {
			case string:
				continue
			case bool:
				continue
			}

			// sanitize the measurement name
			n = invalidNameCharRE.ReplaceAllString(n, "_")
			var mname string
			if n == "value" {
				mname = key
			} else {
				mname = fmt.Sprintf("%s_%s", key, n)
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
				log.Printf("ERROR creating prometheus metric, "+
					"key: %s, labels: %v,\nerr: %s\n",
					mname, l, err.Error())
			}
			p.metrics[desc.String()] = metric
		}
	}
	return nil
}

func init() {
	outputs.Add("prometheus_client", func() telegraf.Output {
		return &PrometheusClient{}
	})
}
