package prometheus

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

type Prometheus struct {
	// An array of urls to scrape metrics from.
	Urls []string

	// An array of Kubernetes services to scrape metrics from.
	KubernetesServices []string

	// Bearer Token authorization file path
	BearerToken string `toml:"bearer_token"`

	ResponseTimeout internal.Duration `toml:"response_timeout"`

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	client *http.Client
}

var sampleConfig = `
  ## An array of urls to scrape metrics from.
  urls = ["http://localhost:9100/metrics"]

  ## An array of Kubernetes services to scrape metrics from.
  # kubernetes_services = ["http://my-service-dns.my-namespace:9100/metrics"]

  ## Use bearer token for authorization
  # bearer_token = /path/to/bearer/token

  ## Specify timeout duration for slower prometheus clients (default is 3s)
  # response_timeout = "3s"

  ## Optional SSL Config
  # ssl_ca = /path/to/cafile
  # ssl_cert = /path/to/certfile
  # ssl_key = /path/to/keyfile
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
`

func (p *Prometheus) SampleConfig() string {
	return sampleConfig
}

func (p *Prometheus) Description() string {
	return "Read metrics from one or many prometheus clients"
}

var ErrProtocolError = errors.New("prometheus protocol error")

func (p *Prometheus) AddressToURL(u *url.URL, address string) string {
	host := address
	if u.Port() != "" {
		host = address + ":" + u.Port()
	}
	reconstructedUrl := url.URL{
		Scheme:     u.Scheme,
		Opaque:     u.Opaque,
		User:       u.User,
		Path:       u.Path,
		RawPath:    u.RawPath,
		ForceQuery: u.ForceQuery,
		RawQuery:   u.RawQuery,
		Fragment:   u.Fragment,
		Host:       host,
	}
	return reconstructedUrl.String()
}

type UrlAndAddress struct {
	OriginalUrl string
	Url         string
	Address     string
}

func (p *Prometheus) GetAllURLs() ([]UrlAndAddress, error) {
	allUrls := make([]UrlAndAddress, 0)
	for _, url := range p.Urls {
		allUrls = append(allUrls, UrlAndAddress{Url: url, OriginalUrl: url})
	}
	for _, service := range p.KubernetesServices {
		u, err := url.Parse(service)
		if err != nil {
			return nil, err
		}
		resolvedAddresses, err := net.LookupHost(u.Hostname())
		if err != nil {
			log.Printf("prometheus: Could not resolve %s, skipping it. Error: %s", u.Host, err)
			continue
		}
		for _, resolved := range resolvedAddresses {
			serviceUrl := p.AddressToURL(u, resolved)
			allUrls = append(allUrls, UrlAndAddress{Url: serviceUrl, Address: resolved, OriginalUrl: service})
		}
	}
	return allUrls, nil
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (p *Prometheus) Gather(acc telegraf.Accumulator) error {
	if p.client == nil {
		client, err := p.createHttpClient()
		if err != nil {
			return err
		}
		p.client = client
	}

	var wg sync.WaitGroup

	allUrls, err := p.GetAllURLs()
	if err != nil {
		return err
	}
	for _, url := range allUrls {
		wg.Add(1)
		go func(serviceUrl UrlAndAddress) {
			defer wg.Done()
			acc.AddError(p.gatherURL(serviceUrl, acc))
		}(url)
	}

	wg.Wait()

	return nil
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

func (p *Prometheus) createHttpClient() (*http.Client, error) {
	tlsCfg, err := internal.GetTLSConfig(
		p.SSLCert, p.SSLKey, p.SSLCA, p.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsCfg,
			DisableKeepAlives: true,
		},
		Timeout: p.ResponseTimeout.Duration,
	}

	return client, nil
}

func (p *Prometheus) gatherURL(url UrlAndAddress, acc telegraf.Accumulator) error {
	var req, err = http.NewRequest("GET", url.Url, nil)
	req.Header.Add("Accept", acceptHeader)
	var token []byte
	var resp *http.Response

	if p.BearerToken != "" {
		token, err = ioutil.ReadFile(p.BearerToken)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+string(token))
	}

	resp, err = p.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url.Url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url.Url, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	metrics, err := Parse(body, resp.Header)
	if err != nil {
		return fmt.Errorf("error reading metrics for %s: %s",
			url.Url, err)
	}
	// Add (or not) collected metrics
	for _, metric := range metrics {
		tags := metric.Tags()
		tags["url"] = url.OriginalUrl
		if url.Address != "" {
			tags["address"] = url.Address
		}

		switch metric.Type() {
		case telegraf.Counter:
			acc.AddCounter(metric.Name(), metric.Fields(), tags, metric.Time())
		case telegraf.Gauge:
			acc.AddGauge(metric.Name(), metric.Fields(), tags, metric.Time())
		case telegraf.Summary:
			acc.AddSummary(metric.Name(), metric.Fields(), tags, metric.Time())
		case telegraf.Histogram:
			acc.AddHistogram(metric.Name(), metric.Fields(), tags, metric.Time())
		default:
			acc.AddFields(metric.Name(), metric.Fields(), tags, metric.Time())
		}
	}

	return nil
}

func init() {
	inputs.Add("prometheus", func() telegraf.Input {
		return &Prometheus{ResponseTimeout: internal.Duration{Duration: time.Second * 3}}
	})
}
