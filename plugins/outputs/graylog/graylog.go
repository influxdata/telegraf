package graylog

import (
	ejson "encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/vanillahsu/graylog-golang"
	"io"
	"os"
)

type Graylog struct {
	Servers []string
	writer  io.Writer
}

var sampleConfig = `
  ## Udp endpoint for your graylog instance.
  servers = ["127.0.0.1:12201", "192.168.1.1:12201"]
`

func (g *Graylog) Connect() error {
	writers := []io.Writer{}

	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:12201")
	}

	for _, server := range g.Servers {
		w := gelf.New(gelf.Config{GraylogEndpoint: server})
		writers = append(writers, w)
	}

	g.writer = io.MultiWriter(writers...)
	return nil
}

func (g *Graylog) Close() error {
	return nil
}

func (g *Graylog) SampleConfig() string {
	return sampleConfig
}

func (g *Graylog) Description() string {
	return "Send telegraf metrics to graylog(s)"
}

func (g *Graylog) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		values, err := serialize(metric)
		if err != nil {
			return err
		}

		for _, value := range values {
			_, err := g.writer.Write([]byte(value))
			if err != nil {
				return fmt.Errorf("FAILED to write message: %s, %s", value, err)
			}
		}
	}
	return nil
}

func serialize(metric telegraf.Metric) ([]string, error) {
	out := []string{}

	m := make(map[string]interface{})
	m["version"] = "1.1"
	m["timestamp"] = metric.UnixNano() / 1000000000
	m["short_message"] = " "
	m["name"] = metric.Name()

	if host, ok := metric.Tags()["host"]; ok {
		m["host"] = host
	} else {
		host, err := os.Hostname()
		if err != nil {
			return []string{}, err
		}
		m["host"] = host
	}

	for key, value := range metric.Fields() {
		nkey := fmt.Sprintf("_%s", key)
		m[nkey] = value
	}

	serialized, err := ejson.Marshal(m)
	if err != nil {
		return []string{}, err
	}
	out = append(out, string(serialized))

	return out, nil
}

func init() {
	outputs.Add("graylog", func() telegraf.Output {
		return &Graylog{}
	})
}
