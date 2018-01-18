package jolokia2

import (
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

type JolokiaAgent struct {
	DefaultFieldPrefix    string
	DefaultFieldSeparator string
	DefaultTagPrefix      string

	URLs            []string `toml:"urls"`
	Username        string
	Password        string
	ResponseTimeout time.Duration `toml:"response_timeout"`

	SSLCA              string `toml:"ssl_ca"`
	SSLCert            string `toml:"ssl_cert"`
	SSLKey             string `toml:"ssl_key"`
	InsecureSkipVerify bool

	Metrics  []MetricConfig `toml:"metric"`
	gatherer *Gatherer
}

func (ja *JolokiaAgent) SampleConfig() string {
	return `
  # default_tag_prefix      = ""
  # default_field_prefix    = ""
  # default_field_separator = "."

  # Add agents URLs to query
  urls = ["http://localhost:8080/jolokia"]
  # username = ""
  # password = ""
  # response_timeout = "5s"

  ## Optional SSL config
  # ssl_ca   = "/var/private/ca.pem"
  # ssl_cert = "/var/private/client.pem"
  # ssl_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ## Add metrics to read
  [[inputs.jolokia2.metric]]
    name  = "java_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]
`
}

func (ja *JolokiaAgent) Description() string {
	return "Read JMX metrics from a Jolokia REST agent endpoint"
}

func (ja *JolokiaAgent) Gather(acc telegraf.Accumulator) error {
	if ja.gatherer == nil {
		ja.gatherer = NewGatherer(ja.createMetrics())
	}

	var wg sync.WaitGroup

	for _, url := range ja.URLs {
		client, err := ja.createClient(url)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to create client for %s: %v", url, err))
			continue
		}

		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()

			err = ja.gatherer.Gather(client, acc)
			if err != nil {
				acc.AddError(fmt.Errorf("Unable to gather metrics for %s: %v", client.URL, err))
			}

		}(client)
	}

	wg.Wait()

	return nil
}

func (ja *JolokiaAgent) createMetrics() []Metric {
	var metrics []Metric

	for _, config := range ja.Metrics {
		metrics = append(metrics, NewMetric(config,
			ja.DefaultFieldPrefix, ja.DefaultFieldSeparator, ja.DefaultTagPrefix))
	}

	return metrics
}

func (ja *JolokiaAgent) createClient(url string) (*Client, error) {
	return NewClient(url, &ClientConfig{
		Username:           ja.Username,
		Password:           ja.Password,
		ResponseTimeout:    ja.ResponseTimeout,
		SSLCA:              ja.SSLCA,
		SSLCert:            ja.SSLCert,
		SSLKey:             ja.SSLKey,
		InsecureSkipVerify: ja.InsecureSkipVerify,
	})
}
