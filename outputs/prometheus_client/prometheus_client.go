package prometheus_client

import (
	"fmt"
	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/outputs"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

type PrometheusClient struct {
	Listen  string
	server  *http.Server
	metrics map[string]*prometheus.UntypedVec
}

var sampleConfig = `
  # Address to listen on
  # listen = ":9126"
`

func (p *PrometheusClient) Start() error {
	if p.Listen == "" {
		p.Listen = ":9126"
	}
	http.Handle("/metrics", prometheus.Handler())
	server := &http.Server{
		Addr:    p.Listen,
		Handler: prometheus.Handler(),
	}
	p.server = server
	p.metrics = make(map[string]*prometheus.UntypedVec)
	go p.server.ListenAndServe()
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

func (p *PrometheusClient) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}

	for _, point := range points {
		var labels []string
		name := point.Name()
		key := name

		for k, _ := range point.Tags() {
			if len(k) > 0 {
				labels = append(labels, k)
			}
		}

		if _, ok := p.metrics[key]; !ok {
			p.metrics[key] = prometheus.NewUntypedVec(
				prometheus.UntypedOpts{
					Name: key,
					Help: fmt.Sprintf("Telegraf collected point '%s'", name),
				},
				labels,
			)
			prometheus.MustRegister(p.metrics[key])
		}

		l := prometheus.Labels{}
		for tk, tv := range point.Tags() {
			l[tk] = tv
		}

		for _, val := range point.Fields() {
			switch val.(type) {
			case int64:
				ival := val.(int64)
				p.metrics[key].With(l).Set(float64(ival))
			case float64:
				p.metrics[key].With(l).Set(val.(float64))
			}
		}
	}
	return nil
}

func init() {
	outputs.Add("prometheus_client", func() outputs.Output {
		return &PrometheusClient{}
	})
}
