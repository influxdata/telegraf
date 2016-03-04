package prometheus_client

import (
	"fmt"
	"log"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/prometheus/client_golang/prometheus"
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
		var labels []string
		key := point.Name()

		for k, _ := range point.Tags() {
			if len(k) > 0 {
				labels = append(labels, k)
			}
		}

		l := prometheus.Labels{}
		for tk, tv := range point.Tags() {
			l[tk] = tv
		}

		for n, val := range point.Fields() {
			mname := fmt.Sprintf("%s_%s", key, n)
			if _, ok := p.metrics[mname]; !ok {
				p.metrics[mname] = prometheus.NewUntypedVec(
					prometheus.UntypedOpts{
						Name: mname,
						Help: fmt.Sprintf("Telegraf collected point '%s'", mname),
					},
					labels,
				)
				prometheus.MustRegister(p.metrics[mname])
			}

			switch val := val.(type) {
			default:
				log.Printf("Prometheus output, unsupported type. key: %s, type: %T\n",
					mname, val)
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
