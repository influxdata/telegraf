package riemann

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/riemann/riemann-go-client"
)

type Riemann struct {
	URL                    string
	CertPath               string
	KeyPath                string
	Insecure               bool
	TTL                    float32
	Separator              string
	MeasurementAsAttribute bool
	StringAsState          bool
	TagKeys                []string
	Tags                   []string
	DescriptionText        string
	Timeout                internal.Duration
	client                 riemanngo.Client
}

var sampleConfig = `
  ## The full TCP, UDP or TLS URL of the Riemann server
  url = "tcp://localhost:5555"

  ## Certificate path for TLS connection
  # cert_path = "cert.pem"

  ## Key path for TLS connection
  # key_path = "key.key"

  ## Activate insecure TLS connection
  # insecure = false

  ## Riemann event TTL, floating-point time in seconds.
  ## Defines how long that an event is considered valid for in Riemann
  # ttl = 30.0

  ## Separator to use between measurement and field name in Riemann service name
  ## This does not have any effect if 'measurement_as_attribute' is set to 'true'
  separator = "/"

  ## Set measurement name as Riemann attribute 'measurement', instead of prepending it to the Riemann service name
  # measurement_as_attribute = false

  ## Send string metrics as Riemann event states.
  ## Unless enabled all string metrics will be ignored
  # string_as_state = false

  ## A list of tag keys whose values get sent as Riemann tags.
  ## If empty, all Telegraf tag values will be sent as tags
  # tag_keys = ["telegraf","custom_tag"]

  ## Additional Riemann tags to send.
  # tags = ["telegraf-output"]

  ## Description for Riemann event
  # description_text = "metrics collected from telegraf"

  ## Riemann client connect and write timeout, defaults to "5s" if not set.
  # timeout = "5s"
`

func (r *Riemann) Connect() error {
	parsed_url, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	var client riemanngo.Client

	if parsed_url.Scheme == "udp" {
		client = riemanngo.NewUdpClient(parsed_url.Host, r.Timeout.Duration)
	} else if parsed_url.Scheme == "tcp" {
		client = riemanngo.NewTcpClient(parsed_url.Host, r.Timeout.Duration)
	} else if parsed_url.Scheme == "tls" {
		client, err = riemanngo.NewTlsClient(parsed_url.Host, r.CertPath, r.KeyPath, r.Insecure, r.Timeout.Duration)
		if err != nil {
			return fmt.Errorf("Error creating TLS connection: %s", err.Error())
		}
	} else {
		return fmt.Errorf("Unknown protocol %s", parsed_url.Scheme)
	}
	err = client.Connect()
	if err != nil {
		r.client = nil
		return err
	}

	r.client = client
	return nil
}

func (r *Riemann) Close() error {
	if r.client != nil {
		err := r.client.Close()
		r.client = nil
		if err != nil {
			return err
		}
	}
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
		if err := r.Connect(); err != nil {
			return fmt.Errorf("Failed to (re)connect to Riemann: %s", err.Error())
		}
	}

	// build list of Riemann events to send
	var events []riemanngo.Event
	for _, m := range metrics {
		evs := r.buildRiemannEvents(m)
		events = append(events, evs...)
	}
	_, err := riemanngo.SendEvents(r.client, &events)
	if err != nil {
		errclose := r.Close()
		if errclose != nil {
			return fmt.Errorf("Failed to send riemann message: %s "+
				"and failed to close Connection : %s",
				err.Error(), errclose.Error())
		}
		return fmt.Errorf("Failed to send riemann message: %s", err.Error())
	}
	return nil
}

func (r *Riemann) buildRiemannEvents(m telegraf.Metric) []riemanngo.Event {
	events := []riemanngo.Event{}
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

		event := riemanngo.Event{
			Host:        host,
			Ttl:         r.TTL,
			Description: r.DescriptionText,
			Time:        m.Time(),

			Attributes: r.attributes(m.Name(), m.Tags()),
			Service:    r.service(m.Name(), fieldName),
			Tags:       r.tags(m.Tags()),
		}

		switch value.(type) {
		case string:
			// only send string metrics if explicitly enabled, skip otherwise
			if !r.StringAsState {
				log.Printf("D! Riemann event states disabled, skipping metric value [%s]\n", value)
				continue
			}
			event.State = value.(string)
		case int, int32, int64, uint, uint32, uint64, float32, float64:
			event.Metric = value
		default:
			log.Printf("D! Riemann does not support metric value [%s]\n", value)
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
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
