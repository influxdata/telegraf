package jolokia2

import (
	"time"

	"github.com/influxdata/telegraf"
)

type JolokiaProxy struct {
	DefaultFieldPrefix    string
	DefaultFieldSeparator string
	DefaultTagPrefix      string

	URL                   string `toml:"url"`
	DefaultTargetPassword string
	DefaultTargetUsername string
	Targets               []JolokiaProxyTargetConfig `toml:"target"`

	Username           string
	Password           string
	SSLCA              string `toml:"ssl_ca"`
	SSLCert            string `toml:"ssl_cert"`
	SSLKey             string `toml:"ssl_key"`
	InsecureSkipVerify bool
	ResponseTimeout    time.Duration `toml:"response_timeout"`

	Metrics  []MetricConfig `toml:"metric"`
	client   *Client
	gatherer *Gatherer
}

type JolokiaProxyTargetConfig struct {
	URL      string `toml:"url"`
	Username string
	Password string
}

func (jp *JolokiaProxy) SampleConfig() string {
	return `
  # default_tag_prefix      = ""
  # default_field_prefix    = ""
  # default_field_separator = "."

  ## Proxy agent
  url = "http://localhost:8080/jolokia"
  # username = ""
  # password = ""
  # response_timeout = "5s"

  ## Optional SSL config
  # ssl_ca   = "/var/private/ca.pem"
  # ssl_cert = "/var/private/client.pem"
  # ssl_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ## Add proxy targets to query
  # default_target_username = ""
  # default_target_password = ""
  [[inputs.jolokia_proxy.target]]
    url = "service:jmx:rmi:///jndi/rmi://targethost:9999/jmxrmi"
	# username = ""
	# password = ""

  ## Add metrics to read
  [[inputs.jolokia_proxy.metric]]
    name  = "java_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]
`
}

func (jp *JolokiaProxy) Description() string {
	return "Read JMX metrics from a Jolokia REST proxy endpoint"
}

func (jp *JolokiaProxy) Gather(acc telegraf.Accumulator) error {
	if jp.gatherer == nil {
		jp.gatherer = NewGatherer(jp.createMetrics())
	}

	if jp.client == nil {
		client, err := jp.createClient()

		if err != nil {
			return err
		}

		jp.client = client
	}

	return jp.gatherer.Gather(jp.client, acc)
}

func (jp *JolokiaProxy) createMetrics() []Metric {
	var metrics []Metric

	for _, config := range jp.Metrics {
		metrics = append(metrics, NewMetric(config,
			jp.DefaultFieldPrefix, jp.DefaultFieldSeparator, jp.DefaultTagPrefix))
	}

	return metrics
}

func (jp *JolokiaProxy) createClient() (*Client, error) {
	proxyConfig := &ProxyConfig{
		DefaultTargetUsername: jp.DefaultTargetUsername,
		DefaultTargetPassword: jp.DefaultTargetPassword,
	}

	for _, target := range jp.Targets {
		proxyConfig.Targets = append(proxyConfig.Targets, ProxyTargetConfig{
			URL:      target.URL,
			Username: target.Username,
			Password: target.Password,
		})
	}

	return NewClient(jp.URL, &ClientConfig{
		Username:           jp.Username,
		Password:           jp.Password,
		ResponseTimeout:    jp.ResponseTimeout,
		SSLCA:              jp.SSLCA,
		SSLCert:            jp.SSLCert,
		SSLKey:             jp.SSLKey,
		InsecureSkipVerify: jp.InsecureSkipVerify,
		ProxyConfig:        proxyConfig,
	})
}
