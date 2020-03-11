package prometheus_remote_write

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/outputs/prometheus_client"
)

func init() {
	outputs.Add("prometheus_remote_write", func() telegraf.Output {
		return &PrometheusRemoteWrite{}
	})
}

type PrometheusRemoteWrite struct {
	URL                  string `toml:"url"`
	BearerToken          string `toml:"bearer_token"`
	BasicUsername        string `toml:"basic_username"`
	BasicPassword        string `toml:"basic_password"`
	RetryForClientErrors bool   `toml:"retry_for_client_errors"`
	tls.ClientConfig

	client http.Client
}

var sampleConfig = `
  ## URL to send Prometheus remote write requests to.
  url = "http://localhost/push"

  ## Optional HTTP asic auth credentials.
  # basic_username = "username"
  # basic_password = "pa55w0rd"

  ## Optional TLS Config for use on HTTP connections.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
	## Optional Bearer token
  # bearer_token = "bearer_token"
  ## Disable retry for 4XX http status codes
  # retry_for_client_errors = false
`

func (p *PrometheusRemoteWrite) Connect() error {
	tlsConfig, err := p.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	p.client = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	return nil
}

func (p *PrometheusRemoteWrite) Close() error {
	return nil
}

func (p *PrometheusRemoteWrite) Description() string {
	return "Configuration for the Prometheus remote write client to spawn"
}

func (p *PrometheusRemoteWrite) SampleConfig() string {
	return sampleConfig
}

func (p *PrometheusRemoteWrite) Write(metrics []telegraf.Metric) error {
	var req prompb.WriteRequest

	for _, metric := range prometheus_client.Sorted(metrics) {
		tags := metric.TagList()
		commonLabels := make([]prompb.Label, 0, len(tags))
		for _, tag := range tags {
			commonLabels = append(commonLabels, prompb.Label{
				Name:  prometheus_client.Sanitize(tag.Key),
				Value: tag.Value,
			})
		}

		// Prometheus doesn't have a string value type, so convert string
		// fields to labels if enabled.
		for fn, fv := range metric.Fields() {
			switch fv := fv.(type) {
			case string:
				tName := prometheus_client.Sanitize(fn)
				if !prometheus_client.IsValidTagName(tName) {
					continue
				}
				commonLabels = append(commonLabels, prompb.Label{
					Name:  tName,
					Value: fv,
				})
			}
		}

		for _, field := range metric.FieldList() {
			var metricName string
			if (metric.Type() == telegraf.Histogram || metric.Type() == telegraf.Summary) && (field.Key != "sum" && field.Key != "count") {
				metricName = prometheus_client.Sanitize(metric.Name())
			} else {
				metricName = getSanitizedMetricName(metric.Name(), prometheus_client.Sanitize(field.Key))
			}
			labels := make([]prompb.Label, len(commonLabels), len(commonLabels)+1)
			copy(labels, commonLabels)
			labels = append(labels, prompb.Label{
				Name:  "__name__",
				Value: metricName,
			})
			sort.Sort(byName(labels))

			// Ignore string and bool fields.
			var value float64
			switch fv := field.Value.(type) {
			case int64:
				value = float64(fv)
			case uint64:
				value = float64(fv)
			case float64:
				value = fv
			default:
				continue
			}

			// Send keys as label values for histograms and summaries
			switch metric.Type() {
			case telegraf.Histogram:
				if field.Key != "sum" && field.Key != "count" {
					labels = append(labels, prompb.Label{
						Name:  "le",
						Value: field.Key,
					})
				}
			case telegraf.Summary:
				if field.Key != "sum" && field.Key != "count" {
					labels = append(labels, prompb.Label{
						Name:  "quantile",
						Value: field.Key,
					})
				}
			}
			ts := metric.Time().UnixNano() / int64(time.Millisecond)
			req.Timeseries = append(req.Timeseries, prompb.TimeSeries{
				Labels: labels,
				Samples: []prompb.Sample{{
					Timestamp: ts,
					Value:     value,
				}},
			})

		}
	}

	buf, err := proto.Marshal(&req)
	if err != nil {
		return err
	}

	compressed := snappy.Encode(nil, buf)
	httpReq, err := http.NewRequest("POST", p.URL, bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	httpReq.Header.Add("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	httpReq.Header.Set("User-Agent", "Telegraf/"+internal.Version())
	if p.BearerToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.BearerToken)
	}

	if p.BasicUsername != "" || p.BasicPassword != "" {
		httpReq.SetBasicAuth(p.BasicUsername, p.BasicPassword)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 && p.retryClientErrors(resp.StatusCode) {
		return fmt.Errorf("server returned HTTP status %s (%d)", resp.Status, resp.StatusCode)
	}
	return nil
}

func (p *PrometheusRemoteWrite) retryClientErrors(statusCode int) bool {
	retryFlag := true
	if p.RetryForClientErrors == false {
		retryFlag = false
	}
	if retryFlag == false && (statusCode == http.StatusTooManyRequests || statusCode == http.StatusBadRequest) {
		log.Printf("E! [outputs.prometheus_remote_write] dropped metrics because of bad request or rate limit.\n")
		return false
	}
	return true
}

func getSanitizedMetricName(name, field string) string {
	return prometheus_client.Sanitize(fmt.Sprintf("%s_%s", name, field))
}

type byName []prompb.Label

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
