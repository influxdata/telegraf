package riemann

import (
	"errors"
	"fmt"
	"os"

	"github.com/amir/raidman"
	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/outputs"
)

type Riemann struct {
	URL       string
	Transport string

	client raidman.Client
}

var sampleConfig = `
  # URL of server
  url = "localhost:5555"
  # transport protocol to use either tcp or udp
  transport = "tcp"
`

func (r *Riemann) Connect() error {
	c, err := raidman.Dial(r.Transport, r.URL)

	if err != nil {
		return err
	}

	r.client = *c
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

func (r *Riemann) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}

	var events []*raidman.Event
	for _, p := range points {
		ev := buildEvent(p)
		events = append(events, &ev)
	}

	var senderr = r.client.SendMulti(events)
	if senderr != nil {
		return errors.New(fmt.Sprintf("FAILED to send riemann message: %s\n",
			senderr))
	}

	return nil
}

func buildEvent(p *client.Point) raidman.Event {
	host := p.Tags()["host"]

	if len(host) == 0 {
		hostname, err := os.Hostname()
		if err != nil {
			host = "unknown"
		} else {
			host = hostname
		}
	}

	var event = &raidman.Event{
		Host:    host,
		Service: p.Name(),
		Metric:  p.Fields()["value"],
	}

	return *event
}

func init() {
	outputs.Add("riemann", func() outputs.Output {
		return &Riemann{}
	})
}
