package prometheus

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

type Prometheus struct {
	Urls []string

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
}

var sampleConfig = `
  ## An array of urls to scrape metrics from.
  urls = ["http://localhost:9100/metrics"]

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

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (p *Prometheus) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	var outerr error

	for _, serv := range p.Urls {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			outerr = p.gatherURL(serv, acc)
		}(serv)
	}

	wg.Wait()

	return outerr
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

func (p *Prometheus) gatherURL(url string, acc telegraf.Accumulator) error {
	collectDate := time.Now()
	var req, err = http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", acceptHeader)
	var token []byte
	var resp *http.Response

	tlsCfg, err := internal.GetTLSConfig(
		p.SSLCert, p.SSLKey, p.SSLCA, p.InsecureSkipVerify)
	if err != nil {
		return err
	}

	var rt http.RoundTripper = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   5 * time.Second,
		TLSClientConfig:       tlsCfg,
		ResponseHeaderTimeout: p.ResponseTimeout.Duration,
		DisableKeepAlives:     true,
	}

	if p.BearerToken != "" {
		token, err = ioutil.ReadFile(p.BearerToken)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+string(token))
	}

	resp, err = rt.RoundTrip(req)
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
		acc.AddFields(metric.Name(), metric.Fields(), tags, collectDate)
	}

	return nil
}

func init() {
	inputs.Add("prometheus", func() telegraf.Input {
		return &Prometheus{ResponseTimeout: internal.Duration{Duration: time.Second * 3}}
	})
}
