package marathon

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  ## Specify Marathon servers via an address:port list
  ## e.g.
  ##   localhost:8080
  ##
  ## If host is specified but port is not specified, then 8080 is used as the port
  ## If no servers are specified, then localhost:8080 is used as the server
  servers = ["localhost:8080"]

  ## Timeout for HTTP requests to the Marathon servers
  http_timeout = "5s"
`

// Marathon is a plugin to read metrics from one or many Marathon servers
type Marathon struct {
	Servers     []string
	HttpTimeout internal.Duration
	client      *http.Client
}

// NewMarathon returns a new instance of Marathon
func NewMarathon() *Marathon {
	return &Marathon{
		HttpTimeout: internal.Duration{Duration: time.Second * 5},
	}
}

// SampleConfig returns a sample configuration block
func (m *Marathon) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the Marathon plugin
func (m *Marathon) Description() string {
	return "Telegraf plugin for gathering metrics from Marathon hosts"
}

// Gather() metrics from given list of Marathon servers
func (m *Marathon) Gather(acc telegraf.Accumulator) error {

	if m.client == nil {
		client, err := m.createHttpClient()

		if err != nil {
			return err
		}

		m.client = client
	}

	if len(m.Servers) == 0 {
		m.Servers = append(m.Servers, "127.0.0.1:8080")
	}

	var wg sync.WaitGroup

	for _, v := range m.Servers {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			acc.AddError(m.gatherMetrics(c, ":8080", acc))
		}(v)
	}

	wg.Wait()

	return nil
}

func (m *Marathon) createHttpClient() (*http.Client, error) {
	tr := &http.Transport{
		ResponseHeaderTimeout: m.HttpTimeout.Duration,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   m.HttpTimeout.Duration,
	}
	return client, nil
}

func contains(strSlice []string, searchStr string) bool {
	for _, value := range strSlice {
		if value == searchStr {
			return true
		}
	}
	return false
}

func (m *Marathon) gatherMetrics(addr string, defaultPort string, acc telegraf.Accumulator) error {
	var jsonOut map[string]map[string]map[string]interface{}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		addr = addr + defaultPort
	}

	tags := map[string]string{
		"server": host,
	}

	resp, err := m.client.Get("http://" + addr + "/metrics")
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(data), &jsonOut); err != nil {
		return errors.New("Error decoding JSON response")
	}

	for _, metrics := range jsonOut {
		for metric, value := range metrics {
			acc.AddFields(strings.Replace(metric, ".", "_", -1), value, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("marathon", func() telegraf.Input {
		return NewMarathon()
	})
}
