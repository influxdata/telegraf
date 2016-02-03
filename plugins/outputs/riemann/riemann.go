package riemann

import (
	"errors"
	"fmt"
	"os"

	"github.com/amir/raidman"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Riemann struct {
	URL       string
	Transport string

	client *raidman.Client
}

var sampleConfig = `
  ### URL of server
  url = "localhost:5555"
  ### transport protocol to use either tcp or udp
  transport = "tcp"
`

func (r *Riemann) Connect() error {
	c, err := raidman.Dial(r.Transport, r.URL)

	if err != nil {
		return err
	}

	r.client = c
	return nil
}

func (r *Riemann) Close() error {
	r.client.Close()
	return nil
}

func (r *Riemann) SampleConfig() string {
	return sampleConfig
}

func (r *Riemann) Description() string {
	return "Configuration for the Riemann server to send metrics to"
}

func (r *Riemann) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	var events []*raidman.Event
	for _, p := range metrics {
		evs := buildEvents(p)
		for _, ev := range evs {
			events = append(events, ev)
		}
	}

	var senderr = r.client.SendMulti(events)
	if senderr != nil {
		return errors.New(fmt.Sprintf("FAILED to send riemann message: %s\n",
			senderr))
	}

	return nil
}

func buildEvents(p telegraf.Metric) []*raidman.Event {
	events := []*raidman.Event{}
	for fieldName, value := range p.Fields() {
		host, ok := p.Tags()["host"]
		if !ok {
			hostname, err := os.Hostname()
			if err != nil {
				host = "unknown"
			} else {
				host = hostname
			}
		}

		event := &raidman.Event{
			Host:    host,
			Service: p.Name() + "_" + fieldName,
			Metric:  value,
		}
		events = append(events, event)
	}

	return events
}

func init() {
	outputs.Add("riemann", func() telegraf.Output {
		return &Riemann{}
	})
}
