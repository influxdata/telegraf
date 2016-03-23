package prometheus_client

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
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
	Listen       string
	SmartStrings bool
	metrics      map[string]*prometheus.UntypedVec
}

var sampleConfig = `
  ## Address to listen on
  # listen = ":9126"
  # smart_strings = true
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
			n = sanitizedChars.Replace(n)
			var mname string
			if n == "value" {
				mname = key
			} else {
				mname = fmt.Sprintf("%s_%s", key, n)
			}

			if !metricName.MatchString(mname) {
				continue
			}

			if _, ok := p.metrics[mname]; !ok {
				p.metrics[mname] = prometheus.NewUntypedVec(
					prometheus.UntypedOpts{
						Name: mname,
						Help: fmt.Sprintf("Telegraf collected point '%s'", mname),
					},
					labels,
				)
				if err := prometheus.Register(p.metrics[mname]); err != nil {
					log.Printf("prometheus_client: Metric failed to register with prometheus, %s", err)
					continue
				}
			}

			switch val := val.(type) {
			default:
				log.Printf("Prometheus output, unsupported type. key: %s, type: %T\n",
					mname, val)
			case string:
				if !p.SmartStrings {
					log.Printf("Prometheus output, unsupported type. key: %s, label: %s, type: %T\n",
						mname, l, val)
				}
				// Get metric value
				m, err := p.metrics[mname].GetMetricWith(l)
				if err != nil {
					log.Printf("ERROR Getting metric in Prometheus output, "+
						"key: %s, labels: %v,\nerr: %s\n",
						mname, l, err.Error())
					continue
				}

				// If has dot in val - parse as float, else int
				if strings.Contains(val, ".") {
					// Float
					tval, err := strconv.ParseFloat(val, 64)
					if err != nil {
						log.Printf("Prometheus output, can't convert string to float. key: %s, label: %s, val: %s\n",
							mname, l, val)
						continue
					}
					m.Set(tval)
				} else {
					// Int
					tval, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						log.Printf("Prometheus output, can't convert string to int. key: %s, label: %s, val: %s\n",
							mname, l, val)
						continue
					}
					m.Set(float64(tval))
				}
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
