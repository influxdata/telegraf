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
	URL          string
	URLs         []string `toml:"urls"`
	DBRoutingTag string   `toml:"database_routing_tag"`
	Username     string
	Password     string
	Database     string
	UserAgent    string
	Precision    string
	Timeout      internal.Duration
	UDPPayload   int `toml:"udp_payload"`

	conns []client.Client
}

var sampleConfig = `
  # The full HTTP or UDP endpoint URL for your InfluxDB instance.
  # Multiple urls can be specified but it is assumed that they are part of the same
  # cluster, this means that only ONE of the urls will be written to each interval.
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

  # Route metrics to an InfluxDB database based on value of a tag
  # note: if this tag has many unique values, write performance may suffer
  # database_routing_tag = "mytag"
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
	var bps map[string]client.BatchPoints = make(map[string]client.BatchPoints)

	// Default database
	bps["___default"], _ = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.Database,
		Precision: i.Precision,
	})

	for _, point := range points {
		if i.DBRoutingTag != "" {
			if h, ok := point.Tags()[i.DBRoutingTag]; ok {
				// Create a new batch point config for this tag value
				if _, ok := bps[h]; !ok {
					bps[h], _ = client.NewBatchPoints(client.BatchPointsConfig{
						Database:  h,
						Precision: i.Precision,
					})
				}
				bps[h].AddPoint(point)
			}
		}

		// Nothing found in overrides, lets push this into the default bucket
		bps["___default"].AddPoint(point)
	}

	// This will get set to nil if a successful write occurs
	err := errors.New("Could not write to any InfluxDB server in cluster")

	for k, bp := range bps {
		p := rand.Perm(len(i.conns))
		for _, n := range p {
			if e := i.conns[n].Write(bp); e != nil {
				log.Println("ERROR: " + e.Error())

				// Stop trying immediately if the error is for a missing database
				// and we are trying a database routing tag
				if k != "___default" && strings.HasPrefix(e.Error(), "database not found") {
					err = nil
					break
				}
			} else {
				err = nil
				break
			}
		}
	}
	return err
}

func init() {
	outputs.Add("influxdb", func() outputs.Output {
		return &InfluxDB{}
	})
}
