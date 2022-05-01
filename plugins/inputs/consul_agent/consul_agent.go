package consul_agent

import (
	"encoding/json"
	"errors"
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

// consul_agent configuration object
type ConsulAgent struct {
	URL string `toml:"url"`

	TokenFile string `toml:"token_file"`
	Token     string `toml:"token"`

	ResponseTimeout config.Duration `toml:"timeout"`

	tls.ClientConfig

	roundTripper http.RoundTripper
}

const timeLayout = "2006-01-02 15:04:05 -0700 MST"

func init() {
	inputs.Add("consul_agent", func() telegraf.Input {
		return &ConsulAgent{
			ResponseTimeout: config.Duration(5 * time.Second),
		}
	})
}

func (n *ConsulAgent) Init() error {
	if n.URL == "" {
		n.URL = "http://127.0.0.1:8500"
	}

	if n.TokenFile != "" && n.Token != "" {
		return errors.New("config error: both token_file and token are set")
	}

	if n.TokenFile != "" {
		token, err := os.ReadFile(n.TokenFile)
		if err != nil {
			return fmt.Errorf("reading file failed: %v", err)
		}
		n.Token = strings.TrimSpace(string(token))
	}

	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("setting up TLS configuration failed: %v", err)
	}

	n.roundTripper = &http.Transport{
		TLSHandshakeTimeout:   time.Duration(n.ResponseTimeout),
		TLSClientConfig:       tlsCfg,
		ResponseHeaderTimeout: time.Duration(n.ResponseTimeout),
	}

	return nil
}

// Gather, collects metrics from Consul endpoint
func (n *ConsulAgent) Gather(acc telegraf.Accumulator) error {
	summaryMetrics, err := n.loadJSON(n.URL + "/v1/agent/metrics")
	if err != nil {
		return err
	}

	return buildConsulAgent(acc, summaryMetrics)
}

func (n *ConsulAgent) loadJSON(url string) (*AgentInfo, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Consul-Token", n.Token)
	req.Header.Add("Accept", "application/json")

	resp, err := n.roundTripper.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	var metrics AgentInfo
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	if err != nil {
		return nil, fmt.Errorf("error parsing json response: %s", err)
	}

	return &metrics, nil
}

// buildConsulAgent, it builds all the metrics and adds them to the accumulator)
func buildConsulAgent(acc telegraf.Accumulator, agentInfo *AgentInfo) error {
	t, err := time.Parse(timeLayout, agentInfo.Timestamp)
	if err != nil {
		return fmt.Errorf("error parsing time: %s", err)
	}

	for _, counters := range agentInfo.Counters {
		fields := map[string]interface{}{
			"count":  counters.Count,
			"sum":    counters.Sum,
			"max":    counters.Max,
			"mean":   counters.Mean,
			"min":    counters.Min,
			"rate":   counters.Rate,
			"stddev": counters.Stddev,
		}
		tags := counters.Labels

		acc.AddCounter(counters.Name, fields, tags, t)
	}

	for _, gauges := range agentInfo.Gauges {
		fields := map[string]interface{}{
			"value": gauges.Value,
		}
		tags := gauges.Labels

		acc.AddGauge(gauges.Name, fields, tags, t)
	}

	for _, points := range agentInfo.Points {
		fields := map[string]interface{}{
			"value": points.Points,
		}
		tags := make(map[string]string)

		acc.AddFields(points.Name, fields, tags, t)
	}

	for _, samples := range agentInfo.Samples {
		fields := map[string]interface{}{
			"count":  samples.Count,
			"sum":    samples.Sum,
			"max":    samples.Max,
			"mean":   samples.Mean,
			"min":    samples.Min,
			"rate":   samples.Rate,
			"stddev": samples.Stddev,
		}
		tags := samples.Labels

		acc.AddCounter(samples.Name, fields, tags, t)
	}

	return nil
}
