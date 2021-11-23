package nomad

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Nomad configuration object
type Nomad struct {
	URL string `toml:"url"`

	AuthToken       string `toml:"auth_token"`
	AuthTokenString string `toml:"auth_token_string"`

	ResponseTimeout config.Duration `toml:"response_timeout"`

	tls.ClientConfig

	roundTripper http.RoundTripper
}

const timeLayout = "2006-01-02 15:04:05 -0700 MST"

var sampleConfig = `
  ## URL for the Nomad agent
  # url = "http://127.0.0.1:4646"

  ## Use auth token for authorization. 
  ## Only one of the options can be set. Leave empty to not use any token.
  # auth_token = "/path/to/auth/token"
  ## OR
  # auth_token_string = "a1234567-40c7-9048-7bae-378687048181"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
`

func init() {
	inputs.Add("nomad", func() telegraf.Input {
		return &Nomad{
			ResponseTimeout: config.Duration(5 * time.Second),
		}
	})
}

// SampleConfig returns a sample config
func (n *Nomad) SampleConfig() string {
	return sampleConfig
}

// Description returns a description of the plugin
func (n *Nomad) Description() string {
	return "Read metrics from the Nomad API"
}

func (n *Nomad) Init() error {
	if n.URL == "" {
		n.URL = "http://127.0.0.1:4646"
	}

	if n.AuthToken != "" && n.AuthTokenString != "" {
		return fmt.Errorf("config error: both auth_token and auth_token_string are set")
	}

	if n.AuthToken != "" {
		token, err := os.ReadFile(n.AuthToken)
		if err != nil {
			return fmt.Errorf("reading file failed: %v", err)
		}
		n.AuthTokenString = strings.TrimSpace(string(token))
	}

	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("setting up TLS configuration failed: %v", err)
	}

	n.roundTripper = &http.Transport{
		TLSHandshakeTimeout:   5 * time.Second,
		TLSClientConfig:       tlsCfg,
		ResponseHeaderTimeout: time.Duration(n.ResponseTimeout),
	}

	return nil
}

// Gather, collects metrics from Nomad endpoint
func (n *Nomad) Gather(acc telegraf.Accumulator) error {
	summaryMetrics := &MetricsSummary{}
	err := n.loadJSON(n.URL+"/v1/metrics", summaryMetrics)
	if err != nil {
		return err
	}

	err = buildNomadMetrics(acc, summaryMetrics)
	if err != nil {
		return err
	}

	return nil
}

func (n *Nomad) loadJSON(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "X-Nomad-Token "+n.AuthTokenString)
	req.Header.Add("Accept", "application/json")

	resp, err := n.roundTripper.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("error parsing json response: %s", err)
	}

	return nil
}

// buildNomadMetrics, it builds all the metrics and adds them to the accumulator)
func buildNomadMetrics(acc telegraf.Accumulator, summaryMetrics *MetricsSummary) error {
	t, err := time.Parse(timeLayout, summaryMetrics.Timestamp)
	if err != nil {
		return fmt.Errorf("error parsing time: %s", err)
	}

	for _, counters := range summaryMetrics.Counters {
		tags := counters.DisplayLabels

		fields := map[string]interface{}{
			"count": counters.Count,
			"rate":  counters.Rate,
			"sum":   counters.Sum,
			"sumsq": counters.SumSq,
			"min":   counters.Min,
			"max":   counters.Max,
			"mean":  counters.Mean,
		}
		acc.AddCounter(counters.Name, fields, tags, t)
	}

	for _, gauges := range summaryMetrics.Gauges {
		tags := gauges.DisplayLabels

		fields := map[string]interface{}{
			"value": gauges.Value,
		}

		acc.AddGauge(gauges.Name, fields, tags, t)

	}

	for _, points := range summaryMetrics.Points {
		tags := make(map[string]string)

		fields := map[string]interface{}{
			"value": points.Points,
		}

		acc.AddFields(points.Name, fields, tags, t)
	}

	for _, samples := range summaryMetrics.Samples {
		tags := samples.DisplayLabels

		fields := map[string]interface{}{
			"count":  samples.Count,
			"rate":   samples.Rate,
			"sum":    samples.Sum,
			"stddev": samples.Stddev,
			"sumsq":  samples.SumSq,
			"min":    samples.Min,
			"max":    samples.Max,
			"mean":   samples.Mean,
		}
		acc.AddCounter(samples.Name, fields, tags, t)
	}

	return nil
}
