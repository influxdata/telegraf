package kibana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// mask for masking username/password from error messages
var mask = regexp.MustCompile(`https?:\/\/\S+:\S+@`)

const statusPath = "/api/status"

type kibanaStatus struct {
	Name    string  `json:"name"`
	UUID    string  `json:"uuid"`
	Version version `json:"version"`
	Status  status  `json:"status"`
	Metrics metrics `json:"metrics"`
}

type version struct {
	Number        string `json:"number"`
	BuildHash     string `json:"build_hash"`
	BuildNumber   int    `json:"build_number"`
	BuildSnapshot bool   `json:"build_snapshot"`
}

type status struct {
	Overall  overallStatus `json:"overall"`
	Statuses interface{}   `json:"statuses"`
}

type overallStatus struct {
	State string `json:"state"`
}

type metrics struct {
	UptimeInMillis        int64         `json:"uptime_in_millis"`
	ConcurrentConnections int64         `json:"concurrent_connections"`
	ResponseTimes         responseTimes `json:"response_times"`
	Process               process       `json:"process"`
}

type responseTimes struct {
	AvgInMillis int64 `json:"avg_in_millis"`
	MaxInMillis int64 `json:"max_in_millis"`
}

type process struct {
	Mem mem `json:"mem"`
}

type mem struct {
	HeapMaxInBytes  int64 `json:"heap_max_in_bytes"`
	HeapUsedInBytes int64 `json:"heap_used_in_bytes"`
}

const sampleConfig = `
  ## specify a list of one or more Kibana servers
  # you can add username and password to your url to use basic authentication:
  # servers = ["http://user:pass@localhost:5601"]
  servers = ["http://localhost:5601"]

  ## Timeout for HTTP requests
  http_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

type Kibana struct {
	Local       bool
	Servers     []string
	HttpTimeout internal.Duration
	tls.ClientConfig

	client                  *http.Client
	catMasterResponseTokens []string
	isMaster                bool
}

func NewKibana() *Kibana {
	return &Kibana{
		HttpTimeout: internal.Duration{Duration: time.Second * 5},
	}
}

// perform status mapping
func mapHealthStatusToCode(s string) int {
	switch strings.ToLower(s) {
	case "green":
		return 1
	case "yellow":
		return 2
	case "red":
		return 3
	}
	return 0
}

// SampleConfig returns sample configuration for this plugin.
func (k *Kibana) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (k *Kibana) Description() string {
	return "Read status information from one or more Kibana servers"
}

func (k *Kibana) Gather(acc telegraf.Accumulator) error {
	if k.client == nil {
		client, err := k.createHttpClient()

		if err != nil {
			return err
		}
		k.client = client
	}

	var wg sync.WaitGroup
	wg.Add(len(k.Servers))

	for _, serv := range k.Servers {
		go func(baseUrl string, acc telegraf.Accumulator) {
			defer wg.Done()
			if err := k.gatherKibanaStatus(baseUrl, acc); err != nil {
				acc.AddError(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
				return
			}
		}(serv, acc)
	}

	wg.Wait()
	return nil
}

func (k *Kibana) createHttpClient() (*http.Client, error) {
	tlsCfg, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: k.HttpTimeout.Duration,
		TLSClientConfig:       tlsCfg,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   k.HttpTimeout.Duration,
	}

	return client, nil
}

func (k *Kibana) gatherKibanaStatus(baseUrl string, acc telegraf.Accumulator) error {

	kibanaStatus := &kibanaStatus{}
	url := baseUrl + statusPath

	host, err := k.gatherJsonData(url, kibanaStatus)
	if err != nil {
		return err
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	tags["name"] = kibanaStatus.Name
	tags["uuid"] = kibanaStatus.UUID
	tags["server"] = host
	tags["version"] = kibanaStatus.Version.Number

	fields["status"] = kibanaStatus.Status.Overall.State
	fields["status_code"] = mapHealthStatusToCode(kibanaStatus.Status.Overall.State)

	fields["uptime_ms"] = kibanaStatus.Metrics.UptimeInMillis
	fields["concurrent_connections"] = kibanaStatus.Metrics.ConcurrentConnections
	fields["heap_max_bytes"] = kibanaStatus.Metrics.Process.Mem.HeapMaxInBytes
	fields["heap_used_bytes"] = kibanaStatus.Metrics.Process.Mem.HeapUsedInBytes
	fields["response_time_avg_ms"] = kibanaStatus.Metrics.ResponseTimes.AvgInMillis
	fields["response_time_max_ms"] = kibanaStatus.Metrics.ResponseTimes.MaxInMillis

	acc.AddFields("kibana", fields, tags)

	return nil
}

func (k *Kibana) gatherJsonData(url string, v interface{}) (host string, err error) {
	r, err := k.client.Get(url)
	if err != nil {
		return "", err
	}

	host = r.Request.Host

	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		return "", fmt.Errorf("kibana: API responded with status-code %d, expected %d",
			r.StatusCode, http.StatusOK)
	}

	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return host, err
	}

	return host, nil
}

func init() {
	inputs.Add("kibana", func() telegraf.Input {
		return NewKibana()
	})
}
