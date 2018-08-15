// +build !freebsd

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

	client *http.Client
}

var sampleConfig = `
  ## The address of the monitoring endpoint of the NATS server
  server = "http://localhost:8222"

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

	if n.client == nil {
		n.client = n.createHTTPClient()
	}
	resp, err := n.client.Get(url.String())
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

	acc.AddFields("nats",
		map[string]interface{}{
			"in_msgs":           stats.InMsgs,
			"out_msgs":          stats.OutMsgs,
			"in_bytes":          stats.InBytes,
			"out_bytes":         stats.OutBytes,
			"uptime":            stats.Now.Sub(stats.Start).Nanoseconds(),
			"cores":             stats.Cores,
			"cpu":               stats.CPU,
			"mem":               stats.Mem,
			"connections":       stats.Connections,
			"total_connections": stats.TotalConnections,
			"subscriptions":     stats.Subscriptions,
			"slow_consumers":    stats.SlowConsumers,
			"routes":            stats.Routes,
			"remotes":           stats.Remotes,
		},
		map[string]string{"server": n.Server},
		time.Now())

	return nil
}

func (n *Nats) createHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	timeout := n.ResponseTimeout.Duration
	if timeout == time.Duration(0) {
		timeout = 5 * time.Second
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

func init() {
	inputs.Add("nats", func() telegraf.Input {
		return &Nats{
			Server: "http://localhost:8222",
		}
	})
}
