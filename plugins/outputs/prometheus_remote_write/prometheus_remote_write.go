package prometheus_remote_write

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

func init() {
	outputs.Add("prometheus_remote_write", func() telegraf.Output {
		return &PrometheusRemoteWrite{}
	})
}

var (
	invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_:]`)
	validNameCharRE   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)
)

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

	for _, metric := range sorted(metrics) {
		tags := metric.TagList()
		commonLabels := make([]prompb.Label, 0, len(tags))
		for _, tag := range tags {
			commonLabels = append(commonLabels, prompb.Label{
				Name:  sanitize(tag.Key),
				Value: tag.Value,
			})
		}

		for _, field := range metric.FieldList() {
			metricName := getSanitizedMetricName(metric.Name(), field.Key)
			labels := make([]prompb.Label, len(commonLabels), len(commonLabels)+1)
			copy(labels, commonLabels)
			labels = append(labels, prompb.Label{
				Name:  "__name__",
				Value: metricName,
			})
			sort.Sort(byName(labels))

			// Ignore histograms and summaries.
			switch metric.Type() {
			case telegraf.Histogram, telegraf.Summary:
				continue
			}

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

func getSanitizedMetricName(name, field string) string {
	return sanitize(fmt.Sprintf("%s_%s", name, field))
}

func sanitize(value string) string {
	return invalidNameCharRE.ReplaceAllString(value, "_")
}

type byName []prompb.Label

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
