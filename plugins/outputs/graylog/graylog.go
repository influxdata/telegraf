package graylog

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/vanillahsu/graylog-golang"

	"log"
)

type Graylog struct {
	Servers []string

	serializer serializers.Serializer
	conns      []*gelf.Gelf
}

var sampleConfig = `
  ## Udp endpoint for your graylog instance.
  servers = ["127.0.0.1:12201", "192.168.1.1:12201"]
`

func (g *Graylog) Connect() error {
	serializer, err := serializers.NewGelfSerializer()
	if err != nil {
		return err
	}

	g.serializer = serializer

	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:12201")
	}
	// Get Connections
	var conns []*gelf.Gelf

	for _, server := range g.Servers {
		conn := gelf.New(gelf.Config{
			GraylogEndpoint: server,
		})

		conns = append(conns, conn)
	}

	g.conns = conns
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
		values, err := g.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		for _, value := range values {
			g.conns[0].Log(value)
		}
	}
	return nil
}

func init() {
	outputs.Add("graylog", func() telegraf.Output {
		return &Graylog{}
	})
}
