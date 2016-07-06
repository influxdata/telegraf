package influxdb

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxDB struct {
	// URL is only for backwards compatability
	URL              string
	URLs             []string `toml:"urls"`
	Username         string
	Password         string
	Database         string
	UserAgent        string
	Precision        string
	RetentionPolicy  string
	WriteConsistency string
	Timeout          internal.Duration
	UDPPayload       int `toml:"udp_payload"`

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	conns []client.Client
}

var sampleConfig = `
  ## The full HTTP or UDP endpoint URL for your InfluxDB instance.
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  # urls = ["udp://localhost:8089"] # UDP endpoint example
  urls = ["http://localhost:8086"] # required
  ## The target database for metrics (telegraf will create it if not exists).
  database = "telegraf" # required
  ## Precision of writes, valid values are "ns", "us" (or "µs"), "ms", "s", "m", "h".
  ## note: using "s" precision greatly improves InfluxDB compression.
  precision = "s"

  ## Retention policy to write to.
  retention_policy = "default"
  ## Write consistency (clusters only), can be: "any", "one", "quorom", "all"
  write_consistency = "any"

  ## Write timeout (for the InfluxDB client), formatted as a string.
  ## If not provided, will default to 5s. 0s means no timeout (not recommended).
  timeout = "5s"
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"
  ## Set the user agent for HTTP POSTs (can be useful for log differentiation)
  # user_agent = "telegraf"
  ## Set UDP payload size, defaults to InfluxDB UDP Client default (512 bytes)
  # udp_payload = 512

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
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

	tlsCfg, err := internal.GetTLSConfig(
		i.SSLCert, i.SSLKey, i.SSLCA, i.InsecureSkipVerify)
	if err != nil {
		return err
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
				TLSConfig: tlsCfg,
			})
			if err != nil {
				return err
			}

			err = createDatabase(c, i.Database)
			if err != nil {
				log.Println("Database creation failed: " + err.Error())
				continue
			}

			conns = append(conns, c)
		}
	}

	i.conns = conns
	rand.Seed(time.Now().UnixNano())
	return nil
}

func createDatabase(c client.Client, database string) error {
	// Create Database if it doesn't exist
	_, err := c.Query(client.Query{
		Command: fmt.Sprintf("CREATE DATABASE IF NOT EXISTS \"%s\"", database),
	})
	return err
}

func (i *InfluxDB) Close() error {
	var errS string
	for j, _ := range i.conns {
		if err := i.conns[j].Close(); err != nil {
			errS += err.Error()
		}
	}
	if errS != "" {
		return fmt.Errorf("output influxdb close failed: %s", errS)
	}
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
func (i *InfluxDB) Write(metrics []telegraf.Metric) error {
	if len(i.conns) == 0 {
		err := i.Connect()
		if err != nil {
			return err
		}
	}
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:         i.Database,
		Precision:        i.Precision,
		RetentionPolicy:  i.RetentionPolicy,
		WriteConsistency: i.WriteConsistency,
	})
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		bp.AddPoint(metric.Point())
	}

	// This will get set to nil if a successful write occurs
	err = errors.New("Could not write to any InfluxDB server in cluster")

	p := rand.Perm(len(i.conns))
	for _, n := range p {
		if e := i.conns[n].Write(bp); e != nil {
			// Log write failure
			log.Printf("ERROR: %s", e)
			// If the database was not found, try to recreate it
			if strings.Contains(e.Error(), "database not found") {
				if errc := createDatabase(i.conns[n], i.Database); errc != nil {
					log.Printf("ERROR: Database %s not found and failed to recreate\n",
						i.Database)
				}
			}
		} else {
			err = nil
			break
		}
	}

	return err
}

func init() {
	outputs.Add("influxdb", func() telegraf.Output {
		return &InfluxDB{
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
