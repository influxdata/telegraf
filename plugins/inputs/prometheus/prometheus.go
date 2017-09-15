package prometheus

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
	"strings"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

type Prometheus struct {
	Urls []string

	// Lookup IP addresses behind URL and gather from them one by one
	DoDnsLookup bool

	// Add a tag named host in the form of 'host:port'
	AddHostTag bool

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

  ## Lookup IP addresses behind URL and gather from them one by one
  do_dns_lookup = false

  ## Add a tag named host in the form of 'host:port'
  add_host_tag = false

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

func (p *Prometheus) SplitHostAndPort(host string) (string, string) {
	hostPort := ""
	hostSplit := strings.Split(host, ":")
	if len(hostSplit) == 2 {
		host = hostSplit[0]
		hostPort = hostSplit[1]
	}
	return host, hostPort
}

func (p *Prometheus) AddressToURL(u *url.URL, address string) string {
	_, port := p.SplitHostAndPort(u.Host)
	reconstructed := u.Scheme + "://" + address + u.Path
	if port != "" {
		reconstructed = u.Scheme + "://" + address + ":" + port + u.Path
	}
	if u.RawQuery != "" {
		reconstructed = reconstructed + "?" + u.RawQuery
	}
	return reconstructed
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

	for _, possibleServ := range p.Urls {
		urls := p.Urls
		if p.DoDnsLookup {
			u, err := url.Parse(possibleServ)
			if err != nil {
				return err
			}
			host, _ := p.SplitHostAndPort(u.Host)
			resolvedAddresses, err := net.LookupHost(host)
			if err != nil {
				log.Printf("prometheus: Could not resolve %s, skipping it. Error: %s", u.Host, err)
				continue
			}
			urls = make([]string, len(resolvedAddresses), len(resolvedAddresses))
			for index, resolved := range resolvedAddresses {
				urls[index] = p.AddressToURL(u, resolved)
			}
		}
		for _, serv := range urls {
			wg.Add(1)
			go func(serviceUrl string) {
				defer wg.Done()
				acc.AddError(p.gatherURL(serviceUrl, acc))
			}(serv)
		}
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

func (p *Prometheus) gatherURL(url string, acc telegraf.Accumulator) error {
	var req, err = http.NewRequest("GET", url, nil)
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
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	metrics, err := Parse(body, resp.Header)
	if err != nil {
		return fmt.Errorf("error reading metrics for %s: %s",
			url, err)
	}
	// Add (or not) collected metrics
	for _, metric := range metrics {
		tags := metric.Tags()
		tags["url"] = url
		if p.AddHostTag {
			tags["host"] = req.Host
		}
		acc.AddFields(metric.Name(), metric.Fields(), tags, metric.Time())
	}

	return nil
}

func init() {
	inputs.Add("prometheus", func() telegraf.Input {
		return &Prometheus{ResponseTimeout: internal.Duration{Duration: time.Second * 3}}
	})
}
