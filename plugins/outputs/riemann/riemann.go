package riemann

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/amir/raidman"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Riemann struct {
	URL              string
	Transport        string
	Separator        string
	MeasurementAsTag bool
	Tags             []string

	client *raidman.Client
}

var sampleConfig = `
  ## URL of server
  url = "localhost:5555"
  ## transport protocol to use either tcp or udp
  transport = "tcp"
  ## separator to use between measurement name and field name in Riemann service name
  separator = " "
  ## set measurement name as a Riemann tag instead of prepending it to the Riemann service name
  measurement_as_tag = false
  ## list of Riemann tags, if specified use these instead of any Telegraf tags
  tags = ["telegraf","custom_tag"]
`

func (r *Riemann) Connect() error {
	c, err := raidman.Dial(r.Transport, r.URL)

	if err != nil {
		r.client = nil
		return err
	}

	r.client = c
	return nil
}

func (r *Riemann) Close() error {
	if r.client == nil {
		return nil
	}
	r.client.Close()
	r.client = nil
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

	if r.client == nil {
		err := r.Connect()
		if err != nil {
			return fmt.Errorf("FAILED to (re)connect to Riemann. Error: %s\n", err)
		}
	}

	var events []*raidman.Event
	for _, p := range metrics {
		evs := r.buildEvents(p)
		for _, ev := range evs {
			events = append(events, ev)
		}
	}

	var senderr = r.client.SendMulti(events)
	if senderr != nil {
		r.Close() // always retuns nil
		return fmt.Errorf("FAILED to send riemann message (will try to reconnect). Error: %s\n",
			senderr)
	}

	return nil
}

func (r *Riemann) buildEvents(p telegraf.Metric) []*raidman.Event {
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
			Host:       host,
			Service:    r.service(p.Name(), fieldName),
			Tags:       r.tags(p.Name(), p.Tags()),
			Attributes: r.attributes(p.Name(), p.Tags()),
			Time:       p.Time().Unix(),
		}

		switch value.(type) {
		case string:
			state := []byte(value.(string))
			event.State = string(state[:254]) // Riemann states must be less than 255 bytes, e.g. "ok", "warning", "critical"
		default:
			event.Metric = value
		}

		events = append(events, event)
	}

	return events
}

func (r *Riemann) attributes(name string, tags map[string]string) map[string]string {
	if r.MeasurementAsTag {
		tags["measurement"] = name
	}
	return tags
}

func (r *Riemann) tags(name string, tags map[string]string) []string {
	if len(r.Tags) > 0 {
		return r.Tags
	}

	var tagNames, tagValues []string

	for tagName := range tags {
		tagNames = append(tagNames, tagName)
	}
	sort.Strings(tagNames)

	if r.MeasurementAsTag {
		tagValues = append(tagValues, name)
	}

	for _, tagName := range tagNames {
		if tagName != "host" { // we'll skip the 'host' tag
			tagValues = append(tagValues, tags[tagName])
		}
	}

	return tagValues
}

func (r *Riemann) service(name string, field string) string {
	var serviceStrings []string

	if !r.MeasurementAsTag {
		serviceStrings = append(serviceStrings, name)
	}
	serviceStrings = append(serviceStrings, field)

	return strings.Join(serviceStrings, r.Separator)
}

func init() {
	outputs.Add("riemann", func() telegraf.Output {
		return &Riemann{}
	})
}
