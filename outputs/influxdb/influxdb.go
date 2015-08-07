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
	Tags      map[string]string

	conn *client.Client
}

func (i *InfluxDB) Connect(host string) error {
	u, err := url.Parse(i.URL)
	if err != nil {
		return err
	}

	c, err := client.NewClient(client.Config{
		URL:       *u,
		Username:  i.Username,
		Password:  i.Password,
		UserAgent: i.UserAgent,
	})

	if err != nil {
		return err
	}

	if i.Tags == nil {
		i.Tags = make(map[string]string)
	}
	i.Tags["host"] = host

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
