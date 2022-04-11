package nginx_plus_api

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type NginxPlusAPI struct {
	Urls            []string        `toml:"urls"`
	APIVersion      int64           `toml:"api_version"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	tls.ClientConfig

	client *http.Client
}

const (
	// Default settings
	defaultAPIVersion = 3

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

func (n *NginxPlusAPI) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Create an HTTP client that is re-used for each
	// collection interval

	if n.APIVersion == 0 {
		n.APIVersion = defaultAPIVersion
	}

	if n.client == nil {
		client, err := n.createHTTPClient()
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

func (n *NginxPlusAPI) createHTTPClient() (*http.Client, error) {
	if n.ResponseTimeout < config.Duration(time.Second) {
		n.ResponseTimeout = config.Duration(time.Second * 5)
	}

	tlsConfig, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(n.ResponseTimeout),
	}

	return client, nil
}

func init() {
	inputs.Add("nginx_plus_api", func() telegraf.Input {
		return &NginxPlusAPI{}
	})
}
