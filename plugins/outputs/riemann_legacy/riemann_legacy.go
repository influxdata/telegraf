package riemann_legacy

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/amir/raidman"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const deprecationMsg = "Error: this Riemann output plugin will be deprecated in a future release, see https://github.com/influxdata/telegraf/issues/1878 for more details & discussion."

type Riemann struct {
	URL       string          `toml:"url"`
	Transport string          `toml:"transport"`
	Separator string          `toml:"separator"`
	Log       telegraf.Logger `toml:"-"`

	client *raidman.Client
}

func (r *Riemann) Connect() error {
	r.Log.Error(deprecationMsg)
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
	err := r.client.Close()
	r.client = nil
	return err
}

func (r *Riemann) Write(metrics []telegraf.Metric) error {
	r.Log.Error(deprecationMsg)
	if len(metrics) == 0 {
		return nil
	}

	if r.client == nil {
		err := r.Connect()
		if err != nil {
			return fmt.Errorf("failed to (re)connect to Riemann, error: %s", err)
		}
	}

	var events []*raidman.Event
	for _, p := range metrics {
		evs := buildEvents(p, r.Separator)
		events = append(events, evs...)
	}

	var senderr = r.client.SendMulti(events)
	if senderr != nil {
		r.Close() //nolint:revive // There is another error which will be returned here
		return fmt.Errorf("failed to send riemann message (will try to reconnect), error: %s", senderr)
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

		switch value := value.(type) {
		case string:
			event.State = value
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
	var tagString = strings.Join(tagStrings, s)
	if tagString != "" {
		serviceStrings = append(serviceStrings, tagString)
	}
	serviceStrings = append(serviceStrings, f)
	return strings.Join(serviceStrings, s)
}

func init() {
	outputs.Add("riemann_legacy", func() telegraf.Output {
		return &Riemann{}
	})
}
