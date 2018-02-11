package flume

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net/url"
	"sync"
	"time"
)

//Flume config
type Flume struct {
	// Flume merics servers
	Servers []string
	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to client cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool
	// HTTP client
	client *http.Client
	// Response timeout
	ResponseTimeout internal.Duration
}

func (f *Flume) Description() string {
	return "Read metrics from one server"
}

func (f *Flume) SampleConfig() string {
	return `
  # specify servers via a url matching:
  #
  servers = ["http://localhost:6666/metrics"]

  # TLS/SSL configuration
  ssl_ca = "/etc/telegraf/ca.pem"
  ssl_cert = "/etc/telegraf/cert.cer"
  ssl_key = "/etc/telegraf/key.key"
  insecure_skip_verify = false
  # HTTP response timeout (default: 5s)
  response_timeout = "5s"
`
}

func (f *Flume) createHTTPClient() (*http.Client, error) {

	tlsCfg, err := internal.GetTLSConfig(
		f.SSLCert, f.SSLKey, f.SSLCA, f.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	if f.ResponseTimeout.Duration < time.Second {
		f.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: f.ResponseTimeout.Duration,
	}

	return client, nil
}

func (f *Flume) gatherURL(addr *url.URL, acc telegraf.Accumulator) error {
	resp, err := f.client.Get(addr.String())
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}

	var metrics map[string]json.RawMessage

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return fmt.Errorf("%s returned unmarshalable body %s", addr.String(), err)
	}

	for c, metricJSON := range metrics {

		tags := map[string]string{"component": c, "server": addr.String()}

		metric := map[string]interface{}{}

		err := json.Unmarshal([]byte(metricJSON), &metric)
		if err != nil {
			return err
		}

		fields := map[string]interface{}{}
		for m, value := range metric {
			fields[m] = value
		}
		acc.AddFields("flume"+"_"+fields["Type"].(string), fields, tags)

	}

	return nil
}

func (f *Flume) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Create an HTTP client that is re-used for each
	// collection interval
	if f.client == nil {
		client, err := f.createHTTPClient()
		if err != nil {
			return err
		}
		f.client = client
	}

	for _, u := range f.Servers {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			acc.AddError(f.gatherURL(addr, acc))
		}(addr)
	}

	wg.Wait()
	return nil

}

func init() {
	inputs.Add("flume", func() telegraf.Input { return &Flume{} })
}
