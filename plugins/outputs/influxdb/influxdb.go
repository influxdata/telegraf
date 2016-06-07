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

const (
	// Size of the fields in each metrics
	metricFieldNum = 25
)

type InfluxDB struct {
	// URL is only for backwards compatability
	URL              string
	URLs             []string `toml:"urls"`
	Username         string
	Password         string
	Database         string
	UserAgent        string
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
	// The limit for measurement size
	// when the size exceeds given limit,
	// points are splitted into multiple parts.
	MetricFieldLimit int `toml:"metric_field_limit"`

	// Precision is only here for legacy support. It will be ignored.
	Precision string

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

  ## The limit for measurement size, when the size exceeds given limit,
  ## points are splitted into multiple parts before writing into InfluxDB.
  ## To enable this, set it >0
  # metric_field_limit = 0
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
		RetentionPolicy:  i.RetentionPolicy,
		WriteConsistency: i.WriteConsistency,
	})
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		bp.AddPoint(metric.Point())
	}

	p := rand.Perm(len(i.conns))

	for _, n := range p {
		conn := i.conns[n]
		// if the connection is UDP and measurements has over i.MetricFieldLimit fields
		// then, write breaking into small portions
		if strings.HasPrefix(i.URLs[n], "udp") && i.MetricFieldLimit != 0 {
			newbp, err := i.splitPoints(bp)
			if err != nil {
				return err
			}
			err = i.flush(conn, newbp)
		} else {
			err = i.flush(conn, bp)
		}
		if err == nil {
			break
		}
	}
	return err
}

// flush writes batch points to the given connection
func (i *InfluxDB) flush(conn client.Client, bp client.BatchPoints) error {
	// initial error, if write successes then error will be a nil
	err := errors.New("Could not write to any InfluxDB server in cluster")
	if e := conn.Write(bp); e != nil {
		// Log write failure
		log.Printf("ERROR: %s", e)
		// If the database was not found, try to recreate it
		if strings.Contains(e.Error(), "database not found") {
			if errc := createDatabase(conn, i.Database); errc != nil {
				log.Printf("ERROR: Database %s not found and failed to recreate\n",
					i.Database)
			}
		}
	} else {
		err = nil
	}
	return err
}

// splitPoints splits all measurements of point into multiple batches,
// the size of each batch will be metricFieldNum
// returns client.BatchPoints with splitted points and error value
func (i *InfluxDB) splitPoints(bp client.BatchPoints) (client.BatchPoints, error) {
	// create new BatchPoints
	batchPoints, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:         i.Database,
		Precision:        i.Precision,
		RetentionPolicy:  i.RetentionPolicy,
		WriteConsistency: i.WriteConsistency,
	})
	// create new Point slice
	// will hold splitted points
	points := []*client.Point{}
	if err != nil {
		return batchPoints, err
	}
	// range over all points
	for _, point := range bp.Points() {
		// the length of the fields
		length := len(point.Fields())
		// split fields only when its size is over batch size - i.MetricFieldLimit
		if length > i.MetricFieldLimit {
			// fields of the point
			fields := point.Fields()
			// from 0 to length with metricFieldNum range
			// on each iteration, the points with metricFieldNum size
			// are appended to points slice
			for j := 0; j < length; j += metricFieldNum {
				start := j
				stop := j + metricFieldNum
				// if the range is over length of the slice
				// then, it is equal to length
				if stop > length {
					stop = length
				}
				batch := map[string]interface{}{}
				// iterate over fields and add to batch map
				for k, v := range fields {
					// put into the batch
					batch[k] = v
					// delete the metric that already retrieved
					delete(fields, k)
					// break the loop if we reached the bound
					if start++; start == stop {
						break
					}
				}
				// create a new point, it has the same params that
				// the parent point(the point that measurements are splitted) has
				// skip points if creation is failed
				p, err := client.NewPoint(point.Name(), point.Tags(), batch, point.Time())
				if err != nil {
					continue
				}
				// append to the global points slice
				points = append(points, p)
			}
		} else {
			points = append(points, point)
		}

	}
	// add splitted points to new BatchPoints
	batchPoints.AddPoints(points)
	return batchPoints, nil
}

func init() {
	outputs.Add("influxdb", func() telegraf.Output {
		return &InfluxDB{
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
