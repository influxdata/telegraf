package prometheus_client

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	sanitizedChars = strings.NewReplacer("/", "_", "@", "_", " ", "_", "-", "_", ".", "_")

	// Prometheus metric names must match this regex
	// see https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	metricName = regexp.MustCompile("^[a-zA-Z_:][a-zA-Z0-9_:]*$")

	// Prometheus labels must match this regex
	// see https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	labelName = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")
)

type PrometheusClient struct {
	Listen  string
	metrics map[string]*prometheus.UntypedVec
}

var sampleConfig = `
  ## Address to listen on
  # listen = ":9126"
`

func (p *PrometheusClient) Start() error {
	if p.Listen == "" {
		p.Listen = "localhost:9126"
	}

	http.Handle("/metrics", prometheus.Handler())
	server := &http.Server{
		Addr: p.Listen,
	}

	p.metrics = make(map[string]*prometheus.UntypedVec)
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

func (p *PrometheusClient) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, point := range metrics {
		key := point.Name()
		key = sanitizedChars.Replace(key)

		var labels []string
		l := prometheus.Labels{}
		for k, v := range point.Tags() {
			k = sanitizedChars.Replace(k)
			if len(k) == 0 {
				continue
			}
			if !labelName.MatchString(k) {
				continue
			}
			labels = append(labels, k)
			l[k] = v
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
			n = sanitizedChars.Replace(n)
			var mname string
			if n == "value" {
				mname = key
			} else {
				mname = fmt.Sprintf("%s_%s", key, n)
			}

			// verify that it is a valid measurement name
			if !metricName.MatchString(mname) {
				continue
			}

			// Create a new metric if it hasn't been created yet.
			if _, ok := p.metrics[mname]; !ok {
				p.metrics[mname] = prometheus.NewUntypedVec(
					prometheus.UntypedOpts{
						Name: mname,
						Help: "Telegraf collected metric",
					},
					labels,
				)
				if err := prometheus.Register(p.metrics[mname]); err != nil {
					log.Printf("prometheus_client: Metric failed to register with prometheus, %s", err)
					continue
				}
			}

			switch val := val.(type) {
			case int64:
				m, err := p.metrics[mname].GetMetricWith(l)
				if err != nil {
					log.Printf("ERROR Getting metric in Prometheus output, "+
						"key: %s, labels: %v,\nerr: %s\n",
						mname, l, err.Error())
					continue
				}
				m.Set(float64(val))
			case float64:
				m, err := p.metrics[mname].GetMetricWith(l)
				if err != nil {
					log.Printf("ERROR Getting metric in Prometheus output, "+
						"key: %s, labels: %v,\nerr: %s\n",
						mname, l, err.Error())
					continue
				}
				m.Set(val)
			default:
				continue
			}
		}
	}
	return nil
}

func init() {
	outputs.Add("prometheus_client", func() telegraf.Output {
		return &PrometheusClient{}
	})
}
