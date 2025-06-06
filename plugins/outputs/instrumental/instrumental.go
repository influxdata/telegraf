//go:generate ../../../tools/readme_config_includer/generator
package instrumental

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
)

//go:embed sample.conf
var sampleConfig string

var (
	ValueIncludesBadChar = regexp.MustCompile("[^[:digit:].]")
	MetricNameReplacer   = regexp.MustCompile("[^-[:alnum:]_.]+")
)

type Instrumental struct {
	Host       string          `toml:"host"`
	Port       int             `toml:"port"`
	APIToken   config.Secret   `toml:"api_token"`
	Prefix     string          `toml:"prefix"`
	DataFormat string          `toml:"data_format"`
	Template   string          `toml:"template"`
	Templates  []string        `toml:"templates"`
	Timeout    config.Duration `toml:"timeout"`
	Debug      bool            `toml:"debug"`

	Log telegraf.Logger `toml:"-"`

	conn       net.Conn
	serializer *graphite.Serializer
}

const (
	DefaultHost     = "collector.instrumentalapp.com"
	DefaultPort     = 8000
	HelloMessage    = "hello version go/telegraf/1.1\n"
	AuthFormat      = "authenticate %s\n"
	HandshakeFormat = HelloMessage + AuthFormat
)

func (*Instrumental) SampleConfig() string {
	return sampleConfig
}

func (i *Instrumental) Init() error {
	s := &graphite.Serializer{
		Prefix:          i.Prefix,
		Template:        i.Template,
		TagSanitizeMode: "strict",
		Separator:       ".",
		Templates:       i.Templates,
	}
	if err := s.Init(); err != nil {
		return err
	}
	i.serializer = s

	return nil
}

func (i *Instrumental) Connect() error {
	addr := fmt.Sprintf("%s:%d", i.Host, i.Port)
	connection, err := net.DialTimeout("tcp", addr, time.Duration(i.Timeout))

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
	err := i.conn.Close()
	i.conn = nil
	return err
}

func (i *Instrumental) Write(metrics []telegraf.Metric) error {
	if i.conn == nil {
		err := i.Connect()
		if err != nil {
			return fmt.Errorf("failed to (re)connect to Instrumental. Error: %w", err)
		}
	}

	var points []string
	var metricType string

	for _, m := range metrics {
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
		metricType = m.Tags()["metric_type"]
		m.RemoveTag("metric_type")

		buf, err := i.serializer.Serialize(m)
		if err != nil {
			i.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		switch metricType {
		case "counter":
			fallthrough
		case "histogram":
			metricType = "increment"
		default:
			metricType = "gauge"
		}

		buffer := bytes.NewBuffer(buf)
		for {
			line, err := buffer.ReadBytes('\n')
			if err != nil {
				break
			}
			stat := string(line)

			// decompose "metric.name value time"
			splitStat := strings.SplitN(stat, " ", 3)
			name := splitStat[0]
			value := splitStat[1]
			timestamp := splitStat[2]

			// replace invalid components of metric name with underscore
			cleanMetric := MetricNameReplacer.ReplaceAllString(name, "_")

			if !ValueIncludesBadChar.MatchString(value) {
				points = append(points, fmt.Sprintf("%s %s %s %s", metricType, cleanMetric, value, timestamp))
			}
		}
	}

	allPoints := strings.Join(points, "")
	if _, err := fmt.Fprint(i.conn, allPoints); err != nil {
		if errors.Is(err, io.EOF) {
			_ = i.Close()
		}

		return err
	}

	// force the connection closed after sending data
	// to deal with various disconnection scenarios and eschew holding
	// open idle connections en masse
	_ = i.Close()

	return nil
}

func (i *Instrumental) authenticate(conn net.Conn) error {
	token, err := i.APIToken.Get()
	if err != nil {
		return fmt.Errorf("getting token failed: %w", err)
	}
	defer token.Destroy()

	if _, err := fmt.Fprintf(conn, HandshakeFormat, token.TemporaryString()); err != nil {
		return err
	}

	// The response here will either be two "ok"s or an error message.
	responses := make([]byte, 512)
	if _, err = conn.Read(responses); err != nil {
		return err
	}

	if string(responses)[:6] != "ok\nok\n" {
		return fmt.Errorf("authentication failed: %s", responses)
	}

	i.conn = conn
	return nil
}

func init() {
	outputs.Add("instrumental", func() telegraf.Output {
		return &Instrumental{
			Host:     DefaultHost,
			Port:     DefaultPort,
			Template: "host.tags.measurement.field", // It is the default template used for graphite serialization
		}
	})
}
