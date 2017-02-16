package natsmonitor

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  ## An array of NATS monitors.
  # urls = ["http://localhost:8222"]
`

type NatsMonitor struct {
	Urls       []string
	httpClient *http.Client
	clients    map[string]*monitorClient
}

func (n *NatsMonitor) SampleConfig() string {
	return sampleConfig
}

func (n *NatsMonitor) Description() string {
	return "Read metrics from one or many NATS monitors"
}

func (n *NatsMonitor) Gather(acc telegraf.Accumulator) error {

	if len(n.clients) == 0 {
		if err := n.setClients(); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	errChan := errchan.New(len(n.clients))

	for _, c := range n.clients {
		wg.Add(1)

		go func(m *monitorClient) {
			defer wg.Done()
			errChan.C <- m.gather(acc)
		}(c)
	}
	wg.Wait()

	return errChan.Error()
}

func (n *NatsMonitor) setClients() error {

	if n.httpClient == nil {
		n.httpClient = &http.Client{
			Timeout: time.Duration(3 * time.Second),
		}
	}

	n.clients = make(map[string]*monitorClient)
	for _, u := range n.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", u, err)
		}
		n.clients[u] = NewMonitorClient(addr, n.httpClient)
	}

	return nil
}

func init() {
	inputs.Add("nats_monitor", func() telegraf.Input {
		return &NatsMonitor{
			Urls: []string{"http://localhost:8222"},
		}
	})
}
