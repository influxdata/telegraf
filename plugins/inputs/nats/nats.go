package nats

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"encoding/json"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	gnatsd "github.com/nats-io/gnatsd/server"
)

type Nats struct {
	Server string
}

var sampleConfig = `
  ## The address of the monitoring end-point of the NATS server
  server = "http://localhost:1337"
`

func (n *Nats) SampleConfig() string {
	return sampleConfig
}

func (n *Nats) Description() string {
	return "Provides metrics about the state of a NATS server"
}

func (n *Nats) Gather(acc telegraf.Accumulator) error {
	theServer := fmt.Sprintf("%s/varz", n.Server)

	/* download the page we are intereted in */
	resp, err := http.Get(theServer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var stats = new(gnatsd.Varz)

	err = json.Unmarshal([]byte(bytes), &stats)
	if err != nil {
		return err
	}

	acc.AddFields("nats",
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

func init() {
	inputs.Add("nats", func() telegraf.Input {
		return &Nats{
			Server: "http://localhost:8222",
		}
	})
}
