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
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Nomad struct {
	URL string

	AuthToken       string `toml:"auth_token"`
	AuthTokenString string `toml:"auth_token_string"`

	LabelInclude []string `toml:"label_include"`
	LabelExclude []string `toml:"label_exclude"`

	labelFilter filter.Filter

	ResponseTimeout config.Duration

	tls.ClientConfig

	RoundTripper http.RoundTripper
}

const timeLayout = "2006-01-02 15:04:05 -0700 MST"

var sampleConfig = `
  ## URL for the Nomad agent
  url = "http://127.0.0.1:4646"

  ## Use auth token for authorization. ('auth_token' takes priority)
  ## If both of these are empty, no token will be used.
  # auth_token = "/path/to/auth/token"
  ## OR
  # auth_token_string = "a1234567-40c7-9048-7bae-378687048181"

  ## Labels to be added as tags. An empty array for both include and
  ## exclude will include all labels.
  # label_include = []
  # label_exclude = ["*"]

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
			LabelInclude: []string{},
			LabelExclude: []string{"*"},
		}
	})
}

func (n *Nomad) SampleConfig() string {
	return sampleConfig
}

func (n *Nomad) Description() string {
	return "Read metrics from the Nomad api"
}

func (n *Nomad) Init() error {
	if n.AuthToken == "" && n.AuthTokenString == "" {
		n.AuthToken = ""
	}

	if n.AuthToken != "" {
		token, err := os.ReadFile(n.AuthToken)
		if err != nil {
			return err
		}
		n.AuthTokenString = strings.TrimSpace(string(token))
	}

	labelFilter, err := filter.NewIncludeExcludeFilter(n.LabelInclude, n.LabelExclude)
	if err != nil {
		return err
	}
	n.labelFilter = labelFilter

	return nil
}

func (n *Nomad) Gather(acc telegraf.Accumulator) error {
	acc.AddError(n.gatherSummary(n.URL, acc))
	return nil
}

func (n *Nomad) gatherSummary(baseURL string, acc telegraf.Accumulator) error {
	summaryMetrics := &MetricsSummary{}
	err := n.LoadJSON(fmt.Sprintf("%s/v1/metrics", baseURL), summaryMetrics)
	if err != nil {
		return err
	}

	buildNomadMetrics(summaryMetrics, acc)

	return nil
}

func (n *Nomad) LoadJSON(url string, v interface{}) error {
	var req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	var resp *http.Response
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	if n.RoundTripper == nil {
		if n.ResponseTimeout < config.Duration(time.Second) {
			n.ResponseTimeout = config.Duration(time.Second * 5)
		}
		n.RoundTripper = &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			TLSClientConfig:       tlsCfg,
			ResponseHeaderTimeout: time.Duration(n.ResponseTimeout),
		}
	}
	req.Header.Set("Authorization", "X-Nomad-Token "+n.AuthTokenString)
	req.Header.Add("Accept", "application/json")
	resp, err = n.RoundTripper.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf(`error parsing response: %s`, err)
	}

	return nil
}

func buildNomadMetrics(summaryMetrics *MetricsSummary, acc telegraf.Accumulator) {

	t, err := time.Parse(timeLayout, summaryMetrics.Timestamp)
	if err != nil {
		panic(err)
	}
	sampledValueFields := make(map[string]interface{})

	for c, counters := range summaryMetrics.Counters {
		tags := summaryMetrics.Counters[c].DisplayLabels

		sampledValueFields["count"] = counters.Count
		sampledValueFields["rate"] = counters.Rate
		sampledValueFields["sum"] = counters.Sum
		sampledValueFields["sumsq"] = counters.SumSq
		sampledValueFields["min"] = counters.Min
		sampledValueFields["max"] = counters.Max
		sampledValueFields["mean"] = counters.Mean

		acc.AddCounter(counters.Name, sampledValueFields, tags, t)
	}

	for c, gauges := range summaryMetrics.Gauges {
		tags := summaryMetrics.Gauges[c].DisplayLabels

		fields := make(map[string]interface{})
		fields["value"] = gauges.Value

		t, err := time.Parse(timeLayout, summaryMetrics.Timestamp)
		if err != nil {
			panic(err)
		}
		acc.AddGauge(gauges.Name, fields, tags, t)
	}

	for _, points := range summaryMetrics.Points {
		tags := make(map[string]string)

		fields := make(map[string]interface{})
		fields["value"] = points.Points

		acc.AddFields(points.Name, fields, tags, t)
	}

	for c, samples := range summaryMetrics.Samples {
		tags := summaryMetrics.Samples[c].DisplayLabels

		sampledValueFields := make(map[string]interface{})
		sampledValueFields["count"] = samples.Count
		sampledValueFields["rate"] = samples.Rate
		sampledValueFields["sum"] = samples.Sum
		sampledValueFields["sumsq"] = samples.SumSq
		sampledValueFields["min"] = samples.Min
		sampledValueFields["max"] = samples.Max
		sampledValueFields["mean"] = samples.Mean

		acc.AddCounter(samples.Name, sampledValueFields, tags, t)
	}
}
