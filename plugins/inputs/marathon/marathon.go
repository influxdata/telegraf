package marathon

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

type Marathon struct {
	Servers     []string
	MetricTypes []string `toml:"metric_types"`
}

var allMetricTypes = []string{
	"gauges", "counters", "histograms", "meters", "timers",
}

var sampleConfig = `
  ## A list of Marathon servers.
  servers = ["localhost:8080"]
  ## Metric types to be collected, by default all enabled.
  metric_types = [
    "gauges",
    "counters",
    "histograms",
    "meters",
    "timers",
  ]
`

// SampleConfig returns a sample configuration block
func (m *Marathon) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the Marathon plugin
func (m *Marathon) Description() string {
	return "Telegraf plugin for gathering metrics from Marathon hosts"
}

func (m *Marathon) SetDefaults() {
	if len(m.MetricTypes) == 0 {
		m.MetricTypes = allMetricTypes
	}
}

// Gather() metrics from given list of Marathon servers
func (m *Marathon) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	m.SetDefaults()

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

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

func contains(strSlice []string, searchStr string) bool {
	for _, value := range strSlice {
		if value == searchStr {
			return true
		}
	}
	return false
}

func (m *Marathon) filterMetrics(metrics map[string]interface{}) {
	for k, _ := range metrics {
		if contains(m.MetricTypes, k) == false {
			delete(metrics, k)
		}
	}
}

func (m *Marathon) gatherMetrics(addr string, defaultPort string, acc telegraf.Accumulator) error {
	var jsonOut map[string]interface{}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		addr = addr + defaultPort
	}

	tags := map[string]string{
		"server": host,
	}

	resp, err := client.Get("http://" + addr + "/metrics")
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

	m.filterMetrics(jsonOut)

	jf := jsonparser.JSONFlattener{}
	err = jf.FlattenJSON("", jsonOut)

	if err != nil {
		return err
	}

	acc.AddFields("marathon", jf.Fields, tags)

	return nil
}

func init() {
	inputs.Add("marathon", func() telegraf.Input {
		return &Marathon{}
	})
}
