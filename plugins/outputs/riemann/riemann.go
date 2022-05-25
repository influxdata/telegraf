package riemann

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/amir/raidman"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Riemann struct {
	URL                    string          `toml:"url"`
	TTL                    float32         `toml:"ttl"`
	Separator              string          `toml:"separator"`
	MeasurementAsAttribute bool            `toml:"measurement_as_attribute"`
	StringAsState          bool            `toml:"string_as_state"`
	TagKeys                []string        `toml:"tag_keys"`
	Tags                   []string        `toml:"tags"`
	DescriptionText        string          `toml:"description_text"`
	Timeout                config.Duration `toml:"timeout"`
	Log                    telegraf.Logger `toml:"-"`

	client *raidman.Client
}

func (r *Riemann) Connect() error {
	parsedURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	client, err := raidman.DialWithTimeout(parsedURL.Scheme, parsedURL.Host, time.Duration(r.Timeout))
	if err != nil {
		r.client = nil
		return err
	}

	r.client = client
	return nil
}

func (r *Riemann) Close() (err error) {
	if r.client != nil {
		err = r.client.Close()
		r.client = nil
	}
	return err
}

func (r *Riemann) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	if r.client == nil {
		if err := r.Connect(); err != nil {
			return fmt.Errorf("failed to (re)connect to Riemann: %s", err.Error())
		}
	}

	// build list of Riemann events to send
	var events []*raidman.Event
	for _, m := range metrics {
		evs := r.buildRiemannEvents(m)
		events = append(events, evs...)
	}

	if err := r.client.SendMulti(events); err != nil {
		r.Close() //nolint:revive // There is another error which will be returned here
		return fmt.Errorf("failed to send riemann message: %s", err)
	}
	return nil
}

func (r *Riemann) buildRiemannEvents(m telegraf.Metric) []*raidman.Event {
	events := []*raidman.Event{}
	for fieldName, value := range m.Fields() {
		// get host for Riemann event
		host, ok := m.Tags()["host"]
		if !ok {
			if hostname, err := os.Hostname(); err == nil {
				host = hostname
			} else {
				host = "unknown"
			}
		}

		event := &raidman.Event{
			Host:        host,
			Ttl:         r.TTL,
			Description: r.DescriptionText,
			Time:        m.Time().Unix(),

			Attributes: r.attributes(m.Name(), m.Tags()),
			Service:    r.service(m.Name(), fieldName),
			Tags:       r.tags(m.Tags()),
		}

		switch value := value.(type) {
		case string:
			// only send string metrics if explicitly enabled, skip otherwise
			if !r.StringAsState {
				r.Log.Debugf("Riemann event states disabled, skipping metric value [%s]", value)
				continue
			}
			event.State = value
		case int, int64, uint64, float32, float64:
			event.Metric = value
		default:
			r.Log.Debugf("Riemann does not support metric value [%s]", value)
			continue
		}

		events = append(events, event)
	}
	return events
}

func (r *Riemann) attributes(name string, tags map[string]string) map[string]string {
	if r.MeasurementAsAttribute {
		tags["measurement"] = name
	}

	delete(tags, "host") // exclude 'host' tag
	return tags
}

func (r *Riemann) service(name string, field string) string {
	var serviceStrings []string

	// if measurement is not enabled as an attribute then prepend it to service name
	if !r.MeasurementAsAttribute {
		serviceStrings = append(serviceStrings, name)
	}
	serviceStrings = append(serviceStrings, field)

	return strings.Join(serviceStrings, r.Separator)
}

func (r *Riemann) tags(tags map[string]string) []string {
	// always add specified Riemann tags
	values := r.Tags

	// if tag_keys are specified, add those and return tag list
	if len(r.TagKeys) > 0 {
		for _, tagName := range r.TagKeys {
			value, ok := tags[tagName]
			if ok {
				values = append(values, value)
			}
		}
		return values
	}

	// otherwise add all values from telegraf tag key/value pairs
	var keys []string
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if key != "host" { // exclude 'host' tag
			values = append(values, tags[key])
		}
	}
	return values
}

func init() {
	outputs.Add("riemann", func() telegraf.Output {
		return &Riemann{
			Timeout: config.Duration(time.Second * 5),
		}
	})
}
