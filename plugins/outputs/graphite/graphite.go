package graphite

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"log"
	"math/rand"
	"net"
	"sort"
	"strings"
	"time"
)

type Graphite struct {
	// URL is only for backwards compatability
	Servers []string
	Prefix  string
	Timeout int
	conns   []net.Conn
}

var sampleConfig = `
  ### TCP endpoint for your graphite instance.
  servers = ["localhost:2003"]
  ### Prefix metrics name
  prefix = ""
  ### timeout in seconds for the write connection to graphite
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
	for _, metric := range metrics {
		// Get name
		name := metric.Name()
		// Convert UnixNano to Unix timestamps
		timestamp := metric.UnixNano() / 1000000000
		tag_str := buildTags(metric)

		for field_name, value := range metric.Fields() {
			// Convert value
			value_str := fmt.Sprintf("%#v", value)
			// Write graphite metric
			var graphitePoint string
			if name == field_name {
				graphitePoint = fmt.Sprintf("%s.%s %s %d\n",
					tag_str,
					strings.Replace(name, ".", "_", -1),
					value_str,
					timestamp)
			} else {
				graphitePoint = fmt.Sprintf("%s.%s.%s %s %d\n",
					tag_str,
					strings.Replace(name, ".", "_", -1),
					strings.Replace(field_name, ".", "_", -1),
					value_str,
					timestamp)
			}
			if g.Prefix != "" {
				graphitePoint = fmt.Sprintf("%s.%s", g.Prefix, graphitePoint)
			}
			bp = append(bp, graphitePoint)
		}
	}
	graphitePoints := strings.Join(bp, "")

	// This will get set to nil if a successful write occurs
	err := errors.New("Could not write to any Graphite server in cluster\n")

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

func buildTags(metric telegraf.Metric) string {
	var keys []string
	tags := metric.Tags()
	for k := range tags {
		if k == "host" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var tag_str string
	if host, ok := tags["host"]; ok {
		if len(keys) > 0 {
			tag_str = strings.Replace(host, ".", "_", -1) + "."
		} else {
			tag_str = strings.Replace(host, ".", "_", -1)
		}
	}

	for i, k := range keys {
		tag_value := strings.Replace(tags[k], ".", "_", -1)
		if i == 0 {
			tag_str += tag_value
		} else {
			tag_str += "." + tag_value
		}
	}
	return tag_str
}

func init() {
	outputs.Add("graphite", func() telegraf.Output {
		return &Graphite{}
	})
}
