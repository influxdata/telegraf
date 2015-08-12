package influxdb

import (
	"net/url"

	"github.com/influxdb/influxdb/client"
	"github.com/influxdb/telegraf/outputs"
)

type InfluxDB struct {
	URL       string
	Username  string
	Password  string
	Database  string
	UserAgent string
	Timeout   Duration

	conn *client.Client
}

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
		Command: fmt.Sprintf("CREATE DATABASE telegraf"),
	})

	if err != nil && !strings.Contains(err.Error(), "database already exists") {
		log.Fatal(err)
	}

	i.conn = c
	return nil
}

func (i *InfluxDB) Write(bp client.BatchPoints) error {
	bp.Database = i.Database
	bp.Tags = i.Tags
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
