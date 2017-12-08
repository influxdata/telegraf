package kentik

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/kentik/libkflow"
)

var (
	allowedChars = regexp.MustCompile(`[^a-zA-Z0-9-_./\p{L}]`)
	hypenChars   = strings.NewReplacer(
		"@", "-",
		"*", "-",
		`%`, "-",
		"#", "-",
		"$", "-")
)

const (
	PLUGIN_NAME    = "telegraph"
	PLUGIN_VERSION = "1.0.0"
)

type Kentik struct {
	Prefix string

	Email    string
	Token    string
	DeviceID int
	FlowDest string

	Debug bool

	IgnoreField string

	client    *libkflow.Sender
	customIds map[string]uint32
}

var sampleConfig = `
  ## prefix for metrics keys
  prefix = "my.specific.prefix."

  ## Kentik user email
  email = ""

  ## Kentik user api token
  token = ""

  ## Kentik device id
  deviceID = 0

  ## DNS name of the Kentik server. Defaults to flow.kentik.com
  flowDest = ""

  ## Debug true - Prints Kentik communication
  debug = false

  ## IgnoreField "" - If fieldName matches this, don't add the field name to the metric passed to TSDB.
  ignoreField = ""
`

func (o *Kentik) Connect() error {

	config := libkflow.NewConfig(o.Email, o.Token, PLUGIN_NAME, PLUGIN_VERSION)

	if o.FlowDest != "" {
		config.SetFlow(o.FlowDest)
	}

	errors := make(chan error, 0)
	client, err := libkflow.NewSenderWithDeviceID(o.DeviceID, errors, config)
	if err != nil {
		return fmt.Errorf("Cannot start client: %v", err)
	}
	go o.handleErrors(errors)
	o.client = client

	o.customIds = map[string]uint32{}
	for _, c := range client.Device.Customs {
		o.customIds[c.Name] = uint32(c.ID)
	}

	return nil
}

func (o *Kentik) handleErrors(errors chan error) {
	for {
		select {
		case msg := <-errors:
			log.Printf("LibError: %v", msg)
		}
	}
}

func (o *Kentik) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	return o.WriteHttp(metrics)
}

func (o *Kentik) WriteHttp(metrics []telegraf.Metric) error {
	for _, m := range metrics {
		now := m.UnixNano() / 1000000000
		tags := cleanTags(m.Tags())

		for fieldName, value := range m.Fields() {
			bval, err := buildValue(value)
			if err != nil {
				log.Printf("D! Kentik does not support metric value: [%s] of type [%T]. %v\n", value, value, err)
				continue
			}

			var metricName string
			if fieldName != o.IgnoreField {
				metricName = sanitize(fmt.Sprintf("%s%s_%s", o.Prefix, m.Name(), fieldName))
			} else {
				metricName = sanitize(fmt.Sprintf("%s%s", o.Prefix, m.Name()))
			}

			metric := &KentikMetric{
				Metric:    metricName,
				Tags:      tags,
				Timestamp: now,
				Value:     bval,
			}

			flow := ToFlow(o.customIds, metric)
			o.client.Send(flow)

			if o.Debug {
				metric.Print()
			}
		}
	}

	return nil
}

func cleanTags(tags map[string]string) map[string]string {
	tagSet := make(map[string]string, len(tags))
	for k, v := range tags {
		tagSet[sanitize(k)] = sanitize(v)
	}
	return tagSet
}

func buildValue(v interface{}) (uint64, error) {
	var retv uint64
	switch p := v.(type) {
	case int64:
		retv = uint64(p)
	case uint64:
		retv = uint64(p)
	case float64:
		retv = uint64(p)
	default:
		return retv, fmt.Errorf("unexpected type %T with value %v for Kentik", v, v)
	}
	return retv, nil
}

func (o *Kentik) SampleConfig() string {
	return sampleConfig
}

func (o *Kentik) Description() string {
	return "Configuration for Kentik server to send metrics to"
}

func (o *Kentik) Close() error {
	return nil
}

func sanitize(value string) string {
	// Apply special hypenation rules to preserve backwards compatibility
	value = hypenChars.Replace(value)
	// Replace any remaining illegal chars
	return allowedChars.ReplaceAllLiteralString(value, "_")
}

func init() {
	outputs.Add("kentik", func() telegraf.Output {
		return &Kentik{}
	})
}
