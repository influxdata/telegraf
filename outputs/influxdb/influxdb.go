package influxdb

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"

	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/duration"
	"github.com/influxdb/telegraf/outputs"
)

type InfluxDB struct {
	// URL is only for backwards compatability
	URL        string
	URLs       []string `toml:"urls"`
	Username   string
	Password   string
	Database   string
	UserAgent  string
	Precision  string
	MultiWrite bool `toml:"multi_write"`
	Timeout    duration.Duration

	conns []client.Client
}

var sampleConfig = `
  # The full HTTP endpoint URL for your InfluxDB instance
  # Multiple urls can be specified for InfluxDB cluster support.
  urls = ["http://localhost:8086"] # required
  # The target database for metrics (telegraf will create it if not exists)
  database = "telegraf" # required
  # Precision of writes, valid values are n, u, ms, s, m, and h
  # note: using second precision greatly helps InfluxDB compression
  precision = "s"

  # MultiWrite treats the urls section (a list of backend nodes) as distinct
  # nodes to write to, instead of nodes of cluster. In cases where you'd like
  # to specify writing to multiple backends, set this value to true (default: false)
  # NOTE: requires use of "urls" instead of deprecated "url"
  multi_write = false

  # Connection timeout (for the connection with InfluxDB), formatted as a string.
  # If not provided, will default to 0 (no timeout)
  # timeout = "5s"
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"
  # Set the user agent for the POSTs (can be useful for log differentiation)
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

	var conns []client.Client
	for _, parsed_url := range urls {
		c := client.NewClient(client.Config{
			URL:       parsed_url,
			Username:  i.Username,
			Password:  i.Password,
			UserAgent: i.UserAgent,
			Timeout:   i.Timeout.Duration,
		})
		conns = append(conns, c)
	}

	for _, conn := range conns {
		_, e := conn.Query(client.Query{
			Command: fmt.Sprintf("CREATE DATABASE %s", i.Database),
		})

		if e != nil && !strings.Contains(e.Error(), "database already exists") {
			log.Println("Database creation failed: " + e.Error())
		} else {
			break
		}
	}

	i.conns = conns
	return nil
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
func (i *InfluxDB) Write(points []*client.Point) error {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.Database,
		Precision: i.Precision,
	})

	for _, point := range points {
		bp.AddPoint(point)
	}

	// Fork here if we see i.MultiWrite; instead of writing to one node in the
	// list of urls hoping for success, we write too each and every one!
	// THIS IS USED FOR WRITING TO MULTIPLE BACKENDS
	// (e.g. multi_write = true; urls = ["http://prodhost:8086", "http://devhost:8086"]
	if i.MultiWrite {
		var is_err bool = false

		err := errors.New("Could not write to url in list.")
		for _, conn := range i.conns {
			if e := conn.Write(bp); e != nil {
				log.Println("ERROR: " + e.Error())
				is_err = true
			}
		}

		if !is_err {
			return nil
		}

		return err
	}

	// This will get set to nil if a successful write occurs
	err := errors.New("Could not write to any InfluxDB server in cluster")

	p := rand.Perm(len(i.conns))
	for _, n := range p {
		if e := i.conns[n].Write(bp); e != nil {
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
