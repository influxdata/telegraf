package nginx_plus_api

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type NginxPlusApi struct {
	Urls            []string          `toml:"urls"`
	ApiVersion      int64             `toml:"api_version"`
	ResponseTimeout internal.Duration `toml:"response_timeout"`
	tls.ClientConfig

	client *http.Client
}

const (
	// Default settings
	defaultApiVersion = 3

	// Paths
	processesPath   = "processes"
	connectionsPath = "connections"
	sslPath         = "ssl"

	httpRequestsPath      = "http/requests"
	httpServerZonesPath   = "http/server_zones"
	httpLocationZonesPath = "http/location_zones"
	httpUpstreamsPath     = "http/upstreams"
	httpCachesPath        = "http/caches"

	resolverZonesPath = "resolvers"

	streamServerZonesPath = "stream/server_zones"
	streamUpstreamsPath   = "stream/upstreams"
)

var sampleConfig = `
  ## An array of API URI to gather stats.
  urls = ["http://localhost/api"]

  # Nginx API version, default: 3
  # api_version = 3

  # HTTP response timeout (default: 5s)
  response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (n *NginxPlusApi) SampleConfig() string {
	return sampleConfig
}

func (n *NginxPlusApi) Description() string {
	return "Read Nginx Plus Api documentation"
}

func (n *NginxPlusApi) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Create an HTTP client that is re-used for each
	// collection interval

	if n.ApiVersion == 0 {
		n.ApiVersion = defaultApiVersion
	}

	if n.client == nil {
		client, err := n.createHttpClient()
		if err != nil {
			return err
		}
		n.client = client
	}

	for _, u := range n.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			n.gatherMetrics(addr, acc)
		}(addr)
	}

	wg.Wait()
	return nil
}

func (n *NginxPlusApi) createHttpClient() (*http.Client, error) {
	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	tlsConfig, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: n.ResponseTimeout.Duration,
	}

	return client, nil
}

func init() {
	inputs.Add("nginx_plus_api", func() telegraf.Input {
		return &NginxPlusApi{}
	})
}
