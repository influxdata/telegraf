package vault

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
type Vault struct {
	URL string `toml:"url"`

	VaultToken       string `toml:"vault_token"`
	VaultTokenString string `toml:"vault_token_string"`

	ResponseTimeout config.Duration `toml:"response_timeout"`

	tls.ClientConfig

	roundTripper http.RoundTripper
}

const timeLayout = "2006-01-02 15:04:05 -0700 MST"

var sampleConfig = `
  ## URL for the Vault agent
  # url = "http://127.0.0.1:8200"

  ## Use vault token for authorization. 
  ## Only one of the options can be set. Leave empty to not use any token.
  # vault_token = "/path/to/auth/token"
  ## OR
  # vault_token_string = "a1234567-40c7-9048-7bae-378687048181"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
`

func init() {
	inputs.Add("vault", func() telegraf.Input {
		return &Vault{
			ResponseTimeout: config.Duration(5 * time.Second),
		}
	})
}

// SampleConfig returns a sample config
func (n *Vault) SampleConfig() string {
	return sampleConfig
}

// Description returns a description of the plugin
func (n *Vault) Description() string {
	return "Read metrics from the Vault API"
}

func (n *Vault) Init() error {
	if n.URL == "" {
		n.URL = "http://127.0.0.1:8200"
	}

	if n.VaultToken != "" && n.VaultTokenString != "" {
		return fmt.Errorf("config error: both auth_token and auth_token_string are set")
	}

	if n.VaultToken != "" {
		token, err := os.ReadFile(n.VaultToken)
		if err != nil {
			return fmt.Errorf("reading file failed: %v", err)
		}
		n.VaultTokenString = strings.TrimSpace(string(token))
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

// Gather, collects metrics from Vault endpoint
func (n *Vault) Gather(acc telegraf.Accumulator) error {
	sysMetrics := &SysMetrics{}
	err := n.loadJSON(n.URL+"/v1/sys/metrics", sysMetrics)
	if err != nil {
		return err
	}

	err = buildVaultMetrics(acc, sysMetrics)
	if err != nil {
		return err
	}

	return nil
}

func (n *Vault) loadJSON(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "X-Vault-Token "+n.VaultTokenString)
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

// buildVaultMetrics, it builds all the metrics and adds them to the accumulator)
func buildVaultMetrics(acc telegraf.Accumulator, sysMetrics *SysMetrics) error {
	t, err := time.Parse(timeLayout, sysMetrics.Timestamp)
	if err != nil {
		return fmt.Errorf("error parsing time: %s", err)
	}

	for _, counters := range sysMetrics.Counters {
		tags := make(map[string]string)
		for key, val := range counters.baseInfo.Labels {
			tags[key] = val.(string)
		}

		fields := map[string]interface{}{
			"count": counters.Count,
			"rate":  counters.Rate,
			"sum":   counters.Sum,
			"min":   counters.Min,
			"max":   counters.Max,
			"mean":  counters.Mean,
		}
		acc.AddCounter(counters.baseInfo.Name, fields, tags, t)
	}

	for _, gauges := range sysMetrics.Gauges {
		tags := make(map[string]string)
		for key, val := range gauges.baseInfo.Labels {
			tags[key] = val.(string)
		}

		fields := map[string]interface{}{
			"value": gauges.Value,
		}

		acc.AddGauge(gauges.Name, fields, tags, t)
	}

	for _, summaries := range sysMetrics.Summaries {
		tags := make(map[string]string)
		for key, val := range summaries.baseInfo.Labels {
			tags[key] = val.(string)
		}

		fields := map[string]interface{}{
			"count":  summaries.Count,
			"rate":   summaries.Rate,
			"sum":    summaries.Sum,
			"stddev": summaries.Stddev,
			"min":    summaries.Min,
			"max":    summaries.Max,
			"mean":   summaries.Mean,
		}
		acc.AddCounter(summaries.Name, fields, tags, t)
	}

	return nil
}
