package cratedb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

	stmt, err := c.DB.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	//_ = stmt
	for _, m := range metrics {
		if _, err := stmt.ExecContext(ctx, m.Name(), m.Time(), 1, 2); err != nil {
			return err
		}
	}
	return nil
}

func insertSQL(table string, metrics []telegraf.Metric) (string, error) {
	rows := make([]string, len(metrics))
	for i, m := range metrics {
		cols := []interface{}{
			m.Name(),
			m.Time(),
			m.Tags(),
		}
		escapedCols := make([]string, len(cols))
		for i, col := range cols {
			escaped, err := escapeValue(col)
			if err != nil {
				return "", err
			}
			escapedCols[i] = escaped
		}
		rows[i] = `(` + strings.Join(escapedCols, ",") + `)`
	}
	sql := `INSERT INTO ` + table + ` ("name", "timestamp", "tags", "tags_hash", "value")
VALUES
` + strings.Join(rows, "  ,\n") + `;`
	return sql, nil
}

// escapeValue returns a string version of val that is suitable for being used
// inside of a VALUES expression or similar. Unsupported types return an error.
//
// Warning: This is not ideal from a security perspective, but unfortunately
// CrateDB does not support enough of the PostgreSQL wire protocol to allow
// using lib/pq with $1, $2 placeholders. Security conscious users of this
// plugin should probably refrain from using it in combination with untrusted
// inputs.
func escapeValue(val interface{}) (string, error) {
	switch t := val.(type) {
	case string:
		return escapeString(t, `'`), nil
	case time.Time:
		// see https://crate.io/docs/crate/reference/sql/data_types.html#timestamp
		return escapeValue(t.Format("2006-01-02T15:04:05.999-0700"))
	case map[string]string:
		// There is a decent chance that the implementation below doesn't catch all
		// edge cases, but it's hard to tell since the format seems to be a bit
		// underspecified. Anyway, luckily we only have to deal with a
		// map[string]string here, giving a higher chance that the code below is
		// correct.
		// See https://crate.io/docs/crate/reference/sql/data_types.html#object
		pairs := make([]string, 0, len(t))
		for k, v := range t {
			val, err := escapeValue(v)
			if err != nil {
				return "", err
			}
			pairs = append(pairs, escapeString(k, `"`)+" = "+val)
		}
		return `{` + strings.Join(pairs, ", ") + `}`, nil
	default:
		// This might be panic worthy under normal circumstances, but it's probably
		// better to not shut down the entire telegraf process because of one
		// misbehaving plugin.
		return "", fmt.Errorf("unexpected type: %#v", t)
	}
}

func escapeString(s string, quote string) string {
	return quote + strings.Replace(s, quote, quote+quote, -1) + quote
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
