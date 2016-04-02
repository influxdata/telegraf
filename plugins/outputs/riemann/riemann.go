package riemann

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/amir/raidman"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Riemann struct {
	URL       string
	Transport string
	Separator string

	client *raidman.Client
}

var sampleConfig = `
  ## URL of server
  url = "localhost:5555"
  ## transport protocol to use either tcp or udp
  transport = "tcp"
  ## separator to use between input name and field name in Riemann service name
  separator = " "
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
		evs := buildEvents(p, r.Separator)
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

func buildEvents(p telegraf.Metric, s string) []*raidman.Event {
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
			Service: serviceName(s, p.Name(), p.Tags(), fieldName),
		}

		switch value.(type) {
		case string:
			event.State = value.(string)
		default:
			event.Metric = value
		}

		events = append(events, event)
	}

	return events
}

func serviceName(s string, n string, t map[string]string, f string) string {
	serviceStrings := []string{}
	serviceStrings = append(serviceStrings, n)

	// we'll skip the 'host' tag
	tagStrings := []string{}
	tagNames := []string{}

	for tagName := range t {
		tagNames = append(tagNames, tagName)
	}
	sort.Strings(tagNames)

	for _, tagName := range tagNames {
		if tagName != "host" {
			tagStrings = append(tagStrings, t[tagName])
		}
	}
	var tagString string = strings.Join(tagStrings, s)
	if tagString != "" {
		serviceStrings = append(serviceStrings, tagString)
	}
	serviceStrings = append(serviceStrings, f)
	return strings.Join(serviceStrings, s)
}

func init() {
	outputs.Add("riemann", func() telegraf.Output {
		return &Riemann{}
	})
}
