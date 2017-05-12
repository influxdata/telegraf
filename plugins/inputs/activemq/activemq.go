package activemq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Activemq struct {
	Urls            []string
	Username        string
	Password        string
	ResponseTimeout internal.Duration
	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool
}

type ActivemqMetrics struct {
	Value Metrics `json:"value"`
}

type Metrics struct {
	TotalConnectionsCount   uint64  `json:"TotalConnectionsCount"`
	TotalProducerCount      uint64  `json:"TotalProducerCount"`
	CurrentConnectionsCount uint64  `json:"CurrentConnectionsCount"`
	TotalDequeueCount       uint64  `json:"TotalDequeueCount"`
	AverageMessageSize      float64 `json:"AverageMessageSize"`
	MinMessageSize          float64 `json:"MinMessageSize"`
	TotalConsumerCount      uint64  `json:"TotalConsumerCount"`
	MaxMessageSize          float64 `json:"MaxMessageSize"`
	TotalMessageCount       uint64  `json:"TotalMessageCount"`
	MemoryPercentUsage      float64 `json:"MemoryPercentUsage"`
	TotalEnqueueCount       uint64  `json:"TotalEnqueueCount"`
}

var sampleConfig = `
  ## An array of Activemq status URI to gather stats.
  ## Default is "http://localhost:8161/api/jolokia/read/org.apache.activemq:type=Broker,brokerName=localhost".
  urls = ["http://localhost:8161/api/jolokia/read/org.apache.activemq:type=Broker,brokerName=localhost"]
	## user credentials for basic HTTP authentication
  username = "myuser"
  password = "mypassword"

  ## Timeout to the complete conection and reponse time in seconds
  response_timeout = "25s" ## default to 5 seconds

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
`

func (n *Activemq) SampleConfig() string {
	return sampleConfig
}

func (n *Activemq) Description() string {
	return "Read Activemq metrics"
}

func (n *Activemq) Gather(acc telegraf.Accumulator) error {
	if len(n.Urls) == 0 {
		n.Urls = []string{"http://localhost:8161/api/jolokia/read/org.apache.activemq:type=Broker,brokerName=localhost"}
	}
	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	var outerr error
	var errch = make(chan error)

	for _, u := range n.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", u, err)
		}

		go func(addr *url.URL) {
			errch <- n.gatherUrl(addr, acc)
		}(addr)
	}

	// Drain channel, waiting for all requests to finish and save last error.
	for range n.Urls {
		if err := <-errch; err != nil {
			outerr = err
		}
	}

	return outerr
}

func (n *Activemq) gatherUrl(addr *url.URL, acc telegraf.Accumulator) error {

	var tr *http.Transport

	if addr.Scheme == "https" {
		tlsCfg, err := internal.GetTLSConfig(
			n.SSLCert, n.SSLKey, n.SSLCA, n.InsecureSkipVerify)
		if err != nil {
			return err
		}
		tr = &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
			TLSClientConfig:       tlsCfg,
		}
	} else {
		tr = &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
		}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   n.ResponseTimeout.Duration,
	}

	req, err := http.NewRequest("GET", addr.String(), nil)
	if err != nil {
		return fmt.Errorf("error on new request to %s : %s\n", addr.String(), err)
	}

	if len(n.Username) != 0 && len(n.Password) != 0 {
		req.SetBasicAuth(n.Username, n.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", addr.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}

	contents, err_contents := ioutil.ReadAll(resp.Body)
	if err_contents != nil {
		return err_contents
	}

	activemqMetrics := ActivemqMetrics{}
	err_json := json.Unmarshal(contents, &activemqMetrics)

	if err_json != nil {
		return err_json
	}

	tags := getTags(addr)

	fields := make(map[string]interface{})

	fields["TotalConnectionsCount"] = activemqMetrics.Value.TotalConnectionsCount
	fields["TotalProducerCount"] = activemqMetrics.Value.TotalProducerCount
	fields["CurrentConnectionsCount"] = activemqMetrics.Value.CurrentConnectionsCount
	fields["TotalDequeueCount"] = activemqMetrics.Value.TotalDequeueCount
	fields["AverageMessageSize"] = activemqMetrics.Value.AverageMessageSize
	fields["MinMessageSize"] = activemqMetrics.Value.MinMessageSize
	fields["TotalConsumerCount"] = activemqMetrics.Value.TotalConsumerCount
	fields["MaxMessageSize"] = activemqMetrics.Value.MaxMessageSize
	fields["TotalMessageCount"] = activemqMetrics.Value.TotalMessageCount
	fields["MemoryPercentUsage"] = activemqMetrics.Value.MemoryPercentUsage
	fields["TotalEnqueueCount"] = activemqMetrics.Value.TotalEnqueueCount

	acc.AddFields("activemq", fields, tags)

	return nil
}

// Get tag(s) for the activemq plugin
func getTags(addr *url.URL) map[string]string {
	h := addr.Host
	host, port, err := net.SplitHostPort(h)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}
	return map[string]string{"server": host, "port": port}
}

func init() {
	inputs.Add("activemq", func() telegraf.Input {
		return &Activemq{}
	})
}
