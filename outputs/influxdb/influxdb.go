package influxdb

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"

	"github.com/influxdb/influxdb/client"
	t "github.com/influxdb/telegraf"
	"github.com/influxdb/telegraf/outputs"
)

type InfluxDB struct {
	// URL is only for backwards compatability
	URL       string
	URLs      []string `toml:"urls"`
	Username  string
	Password  string
	Database  string
	UserAgent string
	Timeout   t.Duration

	conns []*client.Client
}

var sampleConfig = `
	# The full HTTP endpoint URL for your InfluxDB instance
	# Multiple urls can be specified for InfluxDB cluster support.
	urls = ["http://localhost:8086"] # required
	# The target database for metrics (telegraf will create it if not exists)
	database = "telegraf" # required

	# # Connection timeout (for the connection with InfluxDB), formatted as a string.
	# # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	# # If not provided, will default to 0 (no timeout)
	# timeout = "5s"
	# username = "telegraf"
	# password = "metricsmetricsmetricsmetrics"
	# # Set the user agent for the POSTs (can be useful for log differentiation)
	# user_agent = "telegraf"
`

func (i *InfluxDB) Connect() error {
	var urls []*url.URL
	for _, URL := range i.URLs {
		u, err := url.Parse(URL)
		if err != nil {
			return err
		}
		urls = append(urls, u)
	}

	// Backward-compatability with single Influx URL config files
	// This could eventually be removed in favor of specifying the urls as a list
	if i.URL != "" {
		u, err := url.Parse(i.URL)
		if err != nil {
			return err
		}
		urls = append(urls, u)
	}

	var conns []*client.Client
	for _, parsed_url := range urls {
		c, err := client.NewClient(client.Config{
			URL:       *parsed_url,
			Username:  i.Username,
			Password:  i.Password,
			UserAgent: i.UserAgent,
			Timeout:   i.Timeout.Duration,
		})
		if err != nil {
			return err
		}
		conns = append(conns, c)
	}

	// This will get set to nil if a successful connection is made
	err := errors.New("Could not create database on any server")

	for _, conn := range conns {
		_, e := conn.Query(client.Query{
			Command: fmt.Sprintf("CREATE DATABASE %s", i.Database),
		})

		if e != nil && !strings.Contains(e.Error(), "database already exists") {
			log.Println("ERROR: " + e.Error())
		} else {
			err = nil
			break
		}
	}

	i.conns = conns
	return err
}

func (i *InfluxDB) Close() error {
	// InfluxDB client does not provide a Close() function
	return nil
}

func (i *InfluxDB) SampleConfig() string {
	return sampleConfig
}

func (i *InfluxDB) Description() string {
	return "Configuration for influxdb server to send metrics to"
}

// Choose a random server in the cluster to write to until a successful write
// occurs, logging each unsuccessful. If all servers fail, return error.
func (i *InfluxDB) Write(bp client.BatchPoints) error {
	bp.Database = i.Database

	// This will get set to nil if a successful write occurs
	err := errors.New("Could not write to any InfluxDB server in cluster")

	p := rand.Perm(len(i.conns))
	for _, n := range p {
		if _, e := i.conns[n].Write(bp); e != nil {
			log.Println("ERROR: " + e.Error())
		} else {
			err = nil
			break
		}
	}
	return err
}

func init() {
	outputs.Add("influxdb", func() outputs.Output {
		return &InfluxDB{}
	})
}
