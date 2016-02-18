package graphite

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Graphite struct {
	// URL is only for backwards compatability
	Servers []string
	Prefix  string
	Timeout int
	conns   []net.Conn
}

var sampleConfig = `
  ## TCP endpoint for your graphite instance.
  servers = ["localhost:2003"]
  ## Prefix metrics name
  prefix = ""
  ## timeout in seconds for the write connection to graphite
  timeout = 2
`

func (g *Graphite) Connect() error {
	// Set default values
	if g.Timeout <= 0 {
		g.Timeout = 2
	}
	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:2003")
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
	var bp []string
	s, err := serializers.NewGraphiteSerializer(g.Prefix)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		gMetrics, err := s.Serialize(metric)
		if err != nil {
			log.Printf("Error serializing some metrics to graphite: %s", err.Error())
		}
		bp = append(bp, gMetrics...)
	}
	graphitePoints := strings.Join(bp, "\n") + "\n"

	// This will get set to nil if a successful write occurs
	err = errors.New("Could not write to any Graphite server in cluster\n")

	// Send data to a random server
	p := rand.Perm(len(g.conns))
	for _, n := range p {
		if _, e := fmt.Fprintf(g.conns[n], graphitePoints); e != nil {
			// Error
			log.Println("ERROR: " + err.Error())
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
