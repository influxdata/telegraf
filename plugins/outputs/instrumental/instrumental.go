package instrumental

import (
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
)

type Instrumental struct {
	Host       string
	ApiToken   string
	Prefix     string
	DataFormat string
	Template   string
	Timeout    internal.Duration
	Debug      bool

	conn net.Conn
}

const (
	DefaultHost = "collector.instrumentalapp.com"
	AuthFormat  = "hello version go/telegraf/1.0\nauthenticate %s\n"
)

var (
	StatIncludesBadChar = regexp.MustCompile("[^[:alnum:][:blank:]-_.]")
)

var sampleConfig = `
  ## Project API Token (required)
  api_token = "API Token" # required
  ## Prefix the metrics with a given name
  prefix = ""
  ## Stats output template (Graphite formatting)
  ## see https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#graphite
  template = "host.tags.measurement.field"
  ## Timeout in seconds to connect
  timeout = "2s"
  ## Display Communcation to Instrumental
  debug = false
`

func (i *Instrumental) Connect() error {
	connection, err := net.DialTimeout("tcp", i.Host+":8000", i.Timeout.Duration)
	if err != nil {
		i.conn = nil
		return err
	}

	err = i.authenticate(connection)
	if err != nil {
		i.conn = nil
		return err
	}

	return nil
}

func (i *Instrumental) Close() error {
	i.conn.Close()
	i.conn = nil
	return nil
}

func (i *Instrumental) Write(metrics []telegraf.Metric) error {
	if i.conn == nil {
		err := i.Connect()
		if err != nil {
			return fmt.Errorf("FAILED to (re)connect to Instrumental. Error: %s\n", err)
		}
	}

	s, err := serializers.NewGraphiteSerializer(i.Prefix, i.Template)
	if err != nil {
		return err
	}

	var points []string
	var metricType string
	var toSerialize telegraf.Metric
	var newTags map[string]string

	for _, metric := range metrics {
		// Pull the metric_type out of the metric's tags. We don't want the type
		// to show up with the other tags pulled from the system, as they go in the
		// beginning of the line instead.
		// e.g we want:
		//
		//  increment some_prefix.host.tag1.tag2.tag3.field value timestamp
		//
		// vs
		//
		//  increment some_prefix.host.tag1.tag2.tag3.counter.field value timestamp
		//
		newTags = metric.Tags()
		metricType = newTags["metric_type"]
		delete(newTags, "metric_type")

		toSerialize, _ = telegraf.NewMetric(
			metric.Name(),
			newTags,
			metric.Fields(),
			metric.Time(),
		)

		stats, err := s.Serialize(toSerialize)
		if err != nil {
			log.Printf("Error serializing a metric to Instrumental: %s", err)
		}

		switch metricType {
		case "counter":
			fallthrough
		case "histogram":
			metricType = "increment"
		default:
			metricType = "gauge"
		}

		for _, stat := range stats {
			if !StatIncludesBadChar.MatchString(stat) {
				points = append(points, fmt.Sprintf("%s %s", metricType, stat))
			} else if i.Debug {
				log.Printf("Unable to send bad stat: %s", stat)
			}
		}
	}

	allPoints := strings.Join(points, "\n") + "\n"
	_, err = fmt.Fprintf(i.conn, allPoints)

	if i.Debug {
		log.Println(allPoints)
	}

	if err != nil {
		if err == io.EOF {
			i.Close()
		}

		return err
	}

	return nil
}

func (i *Instrumental) Description() string {
	return "Configuration for sending metrics to an Instrumental project"
}

func (i *Instrumental) SampleConfig() string {
	return sampleConfig
}

func (i *Instrumental) authenticate(conn net.Conn) error {
	_, err := fmt.Fprintf(conn, AuthFormat, i.ApiToken)
	if err != nil {
		return err
	}

	// The response here will either be two "ok"s or an error message.
	responses := make([]byte, 512)
	if _, err = conn.Read(responses); err != nil {
		return err
	}

	if string(responses)[:6] != "ok\nok\n" {
		return fmt.Errorf("Authentication failed: %s", responses)
	}

	i.conn = conn
	return nil
}

func init() {
	outputs.Add("instrumental", func() telegraf.Output {
		return &Instrumental{
			Host:     DefaultHost,
			Template: graphite.DEFAULT_TEMPLATE,
		}
	})
}
