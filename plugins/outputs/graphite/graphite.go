package graphite

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Graphite struct {
	// URL is only for backwards compatability
	Servers []string
	Timeout int

	conns      []net.Conn
	serializer serializers.Serializer
}

var sampleConfig = `
  ## TCP endpoint for your graphite instance.
  ## If multiple endpoints are configured, output will be load balanced.
  ## Only one of the endpoints will be written to with each iteration.
  servers = ["localhost:2003"]
  ## timeout in seconds for the write connection to graphite
  timeout = 2
  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "graphite"
  # prefix each graphite bucket
  prefix = ""
  # Graphite output template
  template = "host.tags.measurement.field"
  # graphite protocol with plain/text or json.
  # If no value is set, plain/text is default.
  protocol = "plain/text"
`

func (g *Graphite) SetSerializer(serializer serializers.Serializer) {
	g.serializer = serializer
}

func (g *Graphite) Connect() error {
	// Set default values
	if g.Timeout <= 0 {
		g.Timeout = 2
	}
	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "127.0.0.1:2003")
	}
	// Get Connections
	var conns []net.Conn
	for _, server := range g.Servers {
		conn, err := net.DialTimeout("tcp", server, time.Duration(g.Timeout)*time.Second)
		if err == nil {
			conns = append(conns, conn)
		}
	}
	g.conns = conns
	return nil
}

func (g *Graphite) Close() error {
	// Closing all connections
	for _, conn := range g.conns {
		conn.Close()
	}
	return nil
}

func (g *Graphite) SampleConfig() string {
	return sampleConfig
}

func (g *Graphite) Description() string {
	return "Configuration for Graphite server to send metrics to"
}

// Choose a random server in the cluster to write to until a successful write
// occurs, logging each unsuccessful. If all servers fail, return error.
func (g *Graphite) Write(metrics []telegraf.Metric) error {
	// Prepare data
	var batch []byte

	for _, metric := range metrics {
		buf, err := g.serializer.Serialize(metric)
		if err != nil {
			log.Printf("E! Error serializing some metrics to graphite: %s", err.Error())
		}
		batch = append(batch, buf...)
	}

	// This will get set to nil if a successful write occurs
	err := errors.New("Could not write to any Graphite server in cluster\n")

	// Send data to a random server
	p := rand.Perm(len(g.conns))
	for _, n := range p {
		if g.Timeout > 0 {
			g.conns[n].SetWriteDeadline(time.Now().Add(time.Duration(g.Timeout) * time.Second))
		}
		if _, e := g.conns[n].Write(batch); e != nil {
			// Error
			log.Println("E! Graphite Error: " + e.Error())
			// Let's try the next one
		} else {
			// Success
			err = nil
			break
		}
	}
	// try to reconnect
	if err != nil {
		g.Connect()
	}
	return err
}

func init() {
	outputs.Add("graphite", func() telegraf.Output {
		return &Graphite{}
	})
}
