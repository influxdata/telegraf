package icinga2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Icinga2 represents the config object for the plugin
type Icinga2 struct {
	URL string

	// Bearer Token authorization file path
	BearerToken string `toml:"bearer_token"`

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	// HTTP Timeout specified as a string - 3s, 1m, 1h
	ResponseTimeout internal.Duration

	Username string
	Password string

	RoundTripper http.RoundTripper
}

var sampleConfig = `
  ## URL for the icinga2 
  url = "https://hostname:5665"

  ## Use bearer token for authorization
  # bearer_token = /path/to/bearer/token

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional SSL Config
  # ssl_ca = /path/to/cafile
  # ssl_cert = /path/to/certfile
  # ssl_key = /path/to/keyfile
  ## Use SSL but skip chain & host verification
  insecure_skip_verify = true

  ## Credentials for basic HTTP authentication.
  username = "root"
  password = "root"
`

const (
	summaryEndpoint = `%s/v1/status`
)

func init() {
	inputs.Add("icinga2", func() telegraf.Input {
		return &Icinga2{}
	})
}

//SampleConfig returns a sample config
func (i *Icinga2) SampleConfig() string {
	return sampleConfig
}

//Description returns the description of this plugin
func (i *Icinga2) Description() string {
	return "Read metrics from icinga2 v1/status"
}

//Gather collects icinga2 metrics from a given URL
func (i *Icinga2) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(i *Icinga2) {
		defer wg.Done()
		acc.AddError(i.gatherSummary(i.URL, acc))
	}(i)
	wg.Wait()
	return nil
}

func buildURL(endpoint string, base string) (*url.URL, error) {
	u := fmt.Sprintf(endpoint, base)
	addr, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse address '%s': %s", u, err)
	}
	return addr, nil
}

func (i *Icinga2) gatherSummary(baseURL string, acc telegraf.Accumulator) error {
	url := fmt.Sprintf("%s/v1/status", baseURL)
	var req, err = http.NewRequest("GET", url, nil)
	var token []byte
	var resp *http.Response

	//DEBUG fmt.Println(i.Username, i.Password)
	req.SetBasicAuth(i.Username, i.Password)

	tlsCfg, err := internal.GetTLSConfig(i.SSLCert, i.SSLKey, i.SSLCA, i.InsecureSkipVerify)
	if err != nil {
		return err
	}

	if i.RoundTripper == nil {
		// Set default values
		if i.ResponseTimeout.Duration < time.Second {
			i.ResponseTimeout.Duration = time.Second * 5
		}
		i.RoundTripper = &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			TLSClientConfig:       tlsCfg,
			ResponseHeaderTimeout: i.ResponseTimeout.Duration,
		}
	}

	if i.BearerToken != "" {
		token, err = ioutil.ReadFile(i.BearerToken)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+string(token))
	}

	resp, err = i.RoundTripper.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	/*DEBUG s := string(bodyText)
	  fmt.Println(s) */

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	summaryMetrics := &SummaryMetrics{}
	//err = json.NewDecoder(resp.Body).Decode(summaryMetrics)
	err = (json.Unmarshal(bodyText, summaryMetrics))
	if err != nil {
		fmt.Printf(`Error parsing response: %s`, err)
	}

	//DEBUG fmt.Println(len(summaryMetrics.RawResult))

	buildCIBStatusMetrics(summaryMetrics, acc)
	//buildNodeMetrics(summaryMetrics, acc)
	//buildPodMetrics(summaryMetrics, acc)
	return nil
}

func buildCIBStatusMetrics(summaryMetrics *SummaryMetrics, acc telegraf.Accumulator) {

	summaryMetrics.Cib = &CIB{}
	err := json.Unmarshal(summaryMetrics.RawResult[1], summaryMetrics.Cib)
	if err != nil {
		fmt.Printf(`Error parsing response: %s`, err)
	}

	if err != nil {
		fmt.Printf(`Error parsing response: %s`, err)
	} else {
		tags := map[string]string{
			"node_name": "bbicinga",
		}
		fields := make(map[string]interface{})
		fields["ActiveHostChecks"] = summaryMetrics.Cib.Status.ActiveHostChecks
		fields["ActiveHostChecks"] = summaryMetrics.Cib.Status.ActiveHostChecks
		fields["ActiveHostChecks15Min"] = summaryMetrics.Cib.Status.ActiveHostChecks15Min
		fields["ActiveHostChecks1Min"] = summaryMetrics.Cib.Status.ActiveHostChecks1Min
		fields["ActiveHostChecks5Min"] = summaryMetrics.Cib.Status.ActiveHostChecks5Min
		fields["ActiveServiceChecks"] = summaryMetrics.Cib.Status.ActiveServiceChecks
		fields["ActiveServiceChecks15Min"] = summaryMetrics.Cib.Status.ActiveServiceChecks15Min
		fields["ActiveServiceChecks1Min"] = summaryMetrics.Cib.Status.ActiveServiceChecks1Min
		fields["ActiveServiceChecks5Min"] = summaryMetrics.Cib.Status.ActiveServiceChecks5Min
		fields["AvgExecutionTime"] = summaryMetrics.Cib.Status.AvgExecutionTime
		fields["AvgLatency"] = summaryMetrics.Cib.Status.AvgLatency
		fields["MaxExecutionTime"] = summaryMetrics.Cib.Status.MaxExecutionTime
		fields["MaxLatency"] = summaryMetrics.Cib.Status.MaxLatency
		fields["MinExecutionTime"] = summaryMetrics.Cib.Status.MinExecutionTime
		fields["MinLatency"] = summaryMetrics.Cib.Status.MinLatency
		fields["NumHostsAcknowledged"] = summaryMetrics.Cib.Status.NumHostsAcknowledged
		fields["NumHostsDown"] = summaryMetrics.Cib.Status.NumHostsDown
		fields["NumHostsFlapping"] = summaryMetrics.Cib.Status.NumHostsFlapping
		fields["NumHostsInDowntime"] = summaryMetrics.Cib.Status.NumHostsInDowntime
		fields["NumHostsPending"] = summaryMetrics.Cib.Status.NumHostsPending
		fields["NumHostsUnreachable"] = summaryMetrics.Cib.Status.NumHostsUnreachable
		fields["NumHostsUp"] = summaryMetrics.Cib.Status.NumHostsUp
		fields["NumServicesAcknowledged"] = summaryMetrics.Cib.Status.NumServicesAcknowledged
		fields["NumServicesCritical"] = summaryMetrics.Cib.Status.NumServicesCritical
		fields["NumServicesFlapping"] = summaryMetrics.Cib.Status.NumServicesFlapping
		fields["NumServicesInDowntime"] = summaryMetrics.Cib.Status.NumServicesInDowntime
		fields["NumServicesOk"] = summaryMetrics.Cib.Status.NumServicesOk
		fields["NumServicesPending"] = summaryMetrics.Cib.Status.NumServicesPending
		fields["NumServicesUnknown"] = summaryMetrics.Cib.Status.NumServicesUnknown
		fields["NumServicesUnreachable"] = summaryMetrics.Cib.Status.NumServicesUnreachable
		fields["NumServicesWarning"] = summaryMetrics.Cib.Status.NumServicesWarning
		fields["PassiveHostChecks"] = summaryMetrics.Cib.Status.PassiveHostChecks
		fields["PassiveHostChecks15Min"] = summaryMetrics.Cib.Status.PassiveHostChecks15Min
		fields["PassiveHostChecks1Min"] = summaryMetrics.Cib.Status.PassiveHostChecks1Min
		fields["PassiveHostChecks5Min"] = summaryMetrics.Cib.Status.PassiveHostChecks5Min
		fields["PassiveServiceChecks"] = summaryMetrics.Cib.Status.PassiveServiceChecks
		fields["PassiveServiceChecks15Min"] = summaryMetrics.Cib.Status.PassiveServiceChecks15Min
		fields["PassiveServiceChecks1Min"] = summaryMetrics.Cib.Status.PassiveServiceChecks1Min
		fields["PassiveServiceChecks5Min"] = summaryMetrics.Cib.Status.PassiveServiceChecks5Min
		fields["Uptime"] = summaryMetrics.Cib.Status.Uptime

		//DEBUG fmt.Println(fields["ActiveHostChecks"])

		acc.AddFields("icinga2", fields, tags)
	}
}
