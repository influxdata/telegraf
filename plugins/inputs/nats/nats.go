package nats

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"encoding/json"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"

	gnatsd "github.com/nats-io/gnatsd/server"
)

type Nats struct {
	Server          string
	ResponseTimeout internal.Duration
}

var sampleConfig = `
  ## The address of the monitoring endpoint of the NATS server
  server = "http://localhost:1337"

  ## Maximum time to receive response
  # response_timeout = "5s"
`

func (n *Nats) SampleConfig() string {
	return sampleConfig
}

func (n *Nats) Description() string {
	return "Provides metrics about the state of a NATS server"
}

func (n *Nats) Gather(acc telegraf.Accumulator) error {
	url, err := url.Parse(n.Server)
	if err != nil {
		return err
	}
	url.Path = path.Join(url.Path, "varz")

	client := n.createHTTPClient()
	resp, err := client.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	stats := new(gnatsd.Varz)
	err = json.Unmarshal([]byte(bytes), &stats)
	if err != nil {
		return err
	}

	acc.AddFields("nats_varz",
		map[string]interface{}{
			"in_msgs":           stats.InMsgs,
			"out_msgs":          stats.OutMsgs,
			"uptime":            time.Since(stats.Start).Seconds(),
			"connections":       stats.Connections,
			"total_connections": stats.TotalConnections,
			"in_bytes":          stats.InBytes,
			"cpu_usage":         stats.CPU,
			"out_bytes":         stats.OutBytes,
			"mem":               stats.Mem,
			"subscriptions":     stats.Subscriptions,
		}, nil, time.Now())

	return nil
}

func (n *Nats) createHTTPClient() *http.Client {
	timeout := n.ResponseTimeout.Duration
	if timeout == time.Duration(0) {
		timeout = 5 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
	}
}

func init() {
	inputs.Add("nats", func() telegraf.Input {
		return &Nats{
			Server: "http://localhost:8222",
		}
	})
}
