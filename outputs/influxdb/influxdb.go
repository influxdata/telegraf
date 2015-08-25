package influxdb

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/influxdb/influxdb/client"
	t "github.com/influxdb/telegraf"
	"github.com/influxdb/telegraf/outputs"
)

type InfluxDB struct {
	URL       string
	Username  string
	Password  string
	Database  string
	UserAgent string
	Timeout   t.Duration

	conn *client.Client
}

var sampleConfig = `
	# The full HTTP endpoint URL for your InfluxDB instance
	url = "http://localhost:8086" # required.

	# The target database for metrics. This database must already exist
	database = "telegraf" # required.

	# Connection timeout (for the connection with InfluxDB), formatted as a string.
	# Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	# If not provided, will default to 0 (no timeout)
	# timeout = "5s"

	# username = "telegraf"
	# password = "metricsmetricsmetricsmetrics"

	# Set the user agent for the POSTs (can be useful for log differentiation)
	# user_agent = "telegraf"
`

func (i *InfluxDB) Connect() error {
	u, err := url.Parse(i.URL)
	if err != nil {
		return err
	}

	c, err := client.NewClient(client.Config{
		URL:       *u,
		Username:  i.Username,
		Password:  i.Password,
		UserAgent: i.UserAgent,
		Timeout:   i.Timeout.Duration,
	})

	if err != nil {
		return err
	}

	_, err = c.Query(client.Query{
		Command: fmt.Sprintf("CREATE DATABASE %s", i.Database),
	})

	if err != nil && !strings.Contains(err.Error(), "database already exists") {
		log.Fatal(err)
	}

	i.conn = c
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

func (i *InfluxDB) Write(bp client.BatchPoints) error {
	bp.Database = i.Database
	if _, err := i.conn.Write(bp); err != nil {
		return err
	}
	return nil
}

func init() {
	outputs.Add("influxdb", func() outputs.Output {
		return &InfluxDB{}
	})
}
