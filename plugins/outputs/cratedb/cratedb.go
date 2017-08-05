package cratedb

import (
	"context"
	"database/sql"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/lib/pq"
)

type CrateDB struct {
	URL         string
	Timeout     internal.Duration
	Table       string
	TableCreate bool
	DB          *sql.DB
}

var sampleConfig = `
  # A lib/pq connection string.
  # See http://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
  url = "postgres://user:password@localhost/?sslmode=disable.
  # The timouet for writing metrics.
  timeout = "5s"
`

func (c *CrateDB) Connect() error {
	db, err := sql.Open("postgres", c.URL)
	if err != nil {
		return err
	} else if c.TableCreate {
		// Insert
		sql := `
CREATE TABLE IF NOT EXISTS ` + c.Table + ` (
  "timestamp" TIMESTAMP,
  "name" STRING,
  "tags_hash" STRING,
  "tags" OBJECT(DYNAMIC),
  "value" DOUBLE,
  PRIMARY KEY ("timestamp", "tags_hash")
);
`
		ctx, _ := context.WithTimeout(context.Background(), c.Timeout.Duration)
		if _, err := db.ExecContext(ctx, sql); err != nil {
			return err
		}
	}
	c.DB = db
	return nil
}

func (c *CrateDB) Write(metrics []telegraf.Metric) error {
	sql := `
INSERT INTO ` + c.Table + ` ("name", "timestamp", "tags", "tags_hash")
VALUES ($1, $2, $3, $4);
`
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout.Duration)
	defer cancel()

	//stmt, err := c.DB.PrepareContext(ctx, sql)
	//if err != nil {
	//return err
	//}
	//defer stmt.Close()

	//_ = stmt
	for _, m := range metrics {
		if _, err := c.DB.ExecContext(ctx, sql, m.Name(), m.Time(), 1, 2); err != nil {
			return err
		}
	}
	return nil
}

func (c *CrateDB) SampleConfig() string {
	return sampleConfig
}

func (c *CrateDB) Description() string {
	return "Configuration for CrateDB to send metrics to."
}

func (c *CrateDB) Close() error {
	return c.DB.Close()
}

func init() {
	outputs.Add("cratedb", func() telegraf.Output {
		return &CrateDB{
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
