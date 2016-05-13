package graylog

import (
	"fmt"
	"github.com/Graylog2/go-gelf/gelf"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"io"
)

type Graylog struct {
	Servers    []string
	serializer serializers.Serializer
	closers    []io.Closer
	writer     io.Writer
}

var sampleConfig = `
  ## Udp endpoint for your graylog instance.
  servers = ["127.0.0.1:12201", "192.168.1.1:12201"]
`

func (g *Graylog) Connect() error {
	writers := []io.Writer{}

	serializer, err := serializers.NewGelfSerializer()
	if err != nil {
		return err
	}

	g.serializer = serializer

	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:12201")
	}

	for _, server := range g.Servers {
		w, err := gelf.NewWriter(server)
		if err != nil {
			return err
		}
		writers = append(writers, w)
		g.closers = append(g.closers, w)
	}

	g.writer = io.MultiWriter(writers...)
	return nil
}

func (g *Graylog) Close() error {
	var errS string
	for _, c := range g.closers {
		if err := c.Close(); err != nil {
			errS += err.Error() + "\n"
		}
	}
	if errS != "" {
		return fmt.Errorf(errS)
	}
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
		values, err := g.serializer.Serialize(metric)
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

func init() {
	outputs.Add("graylog", func() telegraf.Output {
		return &Graylog{}
	})
}
