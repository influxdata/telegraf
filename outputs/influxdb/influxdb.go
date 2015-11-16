package influxdb

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"

	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/internal"
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
	Timeout    internal.Duration
	UDPPayload int `toml:"udp_payload"`

	conns []client.Client
}

var sampleConfig = `
  # The full HTTP or UDP endpoint URL for your InfluxDB instance
  # Multiple urls can be specified for InfluxDB cluster support.
  # urls = ["udp://localhost:8089"] # UDP endpoint example
  urls = ["http://localhost:8086"] # required
  # The target database for metrics (telegraf will create it if not exists)
  database = "telegraf" # required
  # Precision of writes, valid values are n, u, ms, s, m, and h
  # note: using second precision greatly helps InfluxDB compression
  precision = "s"

  # Connection timeout (for the connection with InfluxDB), formatted as a string.
  # If not provided, will default to 0 (no timeout)
  # timeout = "5s"
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"
  # Set the user agent for HTTP POSTs (can be useful for log differentiation)
  # user_agent = "telegraf"
  # Set UDP payload size, defaults to InfluxDB UDP Client default (512 bytes)
  # udp_payload = 512
`

func (i *InfluxDB) Connect() error {
	var urls []string
	for _, u := range i.URLs {
		urls = append(urls, u)
	}

	// Backward-compatability with single Influx URL config files
	// This could eventually be removed in favor of specifying the urls as a list
	if i.URL != "" {
		urls = append(urls, i.URL)
	}

	var conns []client.Client
	for _, u := range urls {
		switch {
		case strings.HasPrefix(u, "udp"):
			parsed_url, err := url.Parse(u)
			if err != nil {
				return err
			}

			if i.UDPPayload == 0 {
				i.UDPPayload = client.UDPPayloadSize
			}
			c, err := client.NewUDPClient(client.UDPConfig{
				Addr:        parsed_url.Host,
				PayloadSize: i.UDPPayload,
			})
			if err != nil {
				return err
			}
			conns = append(conns, c)
		default:
			// If URL doesn't start with "udp", assume HTTP client
			c, err := client.NewHTTPClient(client.HTTPConfig{
				Addr:      u,
				Username:  i.Username,
				Password:  i.Password,
				UserAgent: i.UserAgent,
				Timeout:   i.Timeout.Duration,
			})
			if err != nil {
				return err
			}

			// Create Database if it doesn't exist
			_, e := c.Query(client.Query{
				Command: fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", i.Database),
			})

			if e != nil {
				log.Println("Database creation failed: " + e.Error())
			}

			conns = append(conns, c)
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
