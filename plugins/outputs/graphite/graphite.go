package graphite

import (
	"bytes"
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
	Servers            []string
	Prefix             string
	Timeout            int
	MetricsNameBuilder map[string][]string
	conns              []net.Conn
}

var sampleConfig = `
  # TCP endpoint for your graphite instance.
  servers = ["localhost:2003"]
  # Prefix metrics name
  prefix = ""
  # timeout in seconds for the write connection to graphite
  timeout = 2
  # # Build custom metric name from tags for each plugins  
  # [graphite.metricsnamebuilder]
  # # Igore unlisted tags and put metric name after disk name
  # diskio = ["host","{{metric}}","name","{{field}}"]
  # disk = ["host","{{metric}}","name","{{field}}"]
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
		// Convert UnixNano to Unix timestamps
		timestamp := metric.UnixNano() / 1000000000
		for field_name, value := range metric.Fields() {
			// Convert value
			value_str := fmt.Sprintf("%#v", value)
			// Write graphite metric
			graphitePoint := fmt.Sprintf("%s %s %d\n",
				g.buildMetricName(metric, field_name),
				value_str,
				timestamp)
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
func (g *Graphite) buildMetricName(metric telegraf.Metric, fieldName string) string {
	metricName := bytes.NewBufferString(g.Prefix)
	metricName.WriteString(".")
	tags := metric.Tags()
	metricsTemplate, ok := g.MetricsNameBuilder[metric.Name()]
	if !ok {
		metricsTemplate = []string{"host"}
		for k := range tags {
			if k != "host" {
				metricsTemplate = append(metricsTemplate, k)
			}
		}
		sort.Strings(metricsTemplate[1:])
		if metric.Name() != fieldName {
			metricsTemplate = append(metricsTemplate, "{{metric}}")
		}
		metricsTemplate = append(metricsTemplate, "{{field}}")
	}
	tags["{{metric}}"] = metric.Name()
	tags["{{field}}"] = fieldName
	for _, tagName := range metricsTemplate {
		tagValue, ok := tags[tagName]
		if !ok || tagValue == "" {
			continue
		}
		metricName.WriteString(strings.Replace(tagValue, ".", "_", -1))
		metricName.WriteString(".")
	}
	return strings.Trim(metricName.String(), ".")
}

func init() {
	outputs.Add("graphite", func() telegraf.Output {
		return &Graphite{}
	})
}
