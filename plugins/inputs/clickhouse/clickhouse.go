package clickhouse

import (
	"database/sql"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	_ "github.com/kshvakov/clickhouse"
)

const sampleConfig = `
### ClickHouse DSN
dsn = "native://localhost:9000?username=user&password=qwerty"
`

var hostname, _ = os.Hostname()

// ClickHouse Telegraf Input Plugin
type ClickHouse struct {
	DSN     string `toml:"dsn"`
	server  string
	connect *sql.DB
}

// SampleConfig returns the sample config
func (*ClickHouse) SampleConfig() string {
	return sampleConfig
}

// Description return plugin description
func (*ClickHouse) Description() string {
	return "Read metrics from ClickHouse server"
}

// Gather collect data from ClickHouse server
func (ch *ClickHouse) Gather(acc telegraf.Accumulator) (err error) {
	var rows *sql.Rows
	for measurement, query := range measurementMap {
		if rows, err = ch.connect.Query(query); err != nil {
			acc.AddError(err)
			return err
		}
		if err := ch.processRows(measurement, rows, acc); err != nil {
			return err
		}
	}
	if rows, err = ch.connect.Query(systemParts); err != nil {
		acc.AddError(err)
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			table, database           string
			bytes, parts, rowsInTable uint64
		)
		if err := rows.Scan(&database, &table, &bytes, &parts, &rowsInTable); err != nil {
			acc.AddError(err)
			return err
		}
		acc.AddFields("clickhouse_tables",
			map[string]interface{}{
				"bytes": bytes,
				"parts": parts,
				"rows":  rowsInTable,
			},
			map[string]string{
				"table":    table,
				"server":   ch.server,
				"hostname": hostname,
				"database": database,
			})
	}
	return nil
}

func (ch *ClickHouse) processRows(measurement string, rows *sql.Rows, acc telegraf.Accumulator) error {
	defer rows.Close()
	fields := make(map[string]interface{})
	for rows.Next() {
		var (
			key   string
			value uint64
		)
		if err := rows.Scan(&key, &value); err != nil {
			acc.AddError(err)
			return err
		}
		fields[key] = value
	}
	acc.AddFields("clickhouse_"+measurement, fields, map[string]string{
		"server":   ch.server,
		"hostname": hostname,
	})
	return nil
}

// Start ClickHouse input service
func (ch *ClickHouse) Start(telegraf.Accumulator) (err error) {
	if ch.connect, err = sql.Open("clickhouse", ch.DSN); err != nil {
		return err
	}
	{
		ch.connect.SetMaxOpenConns(2)
		ch.connect.SetMaxIdleConns(1)
		ch.connect.SetConnMaxLifetime(20 * time.Second)
	}
	return ch.connect.QueryRow("SELECT hostName()").Scan(&ch.server)
}

// Stop ClickHouse input service
func (ch *ClickHouse) Stop(ClickHouse) {
	if ch.connect != nil {
		ch.connect.Close()
	}
}

func init() {
	inputs.Add("clickhouse", func() telegraf.Input {
		return &ClickHouse{}
	})
}

const (
	systemEventsSQL       = "SELECT event,  CAST(value AS UInt64) AS value FROM system.events"
	systemMetricsSQL      = "SELECT metric, CAST(value AS UInt64) AS value FROM system.metrics"
	systemAsyncMetricsSQL = "SELECT metric, CAST(value AS UInt64) AS value FROM system.asynchronous_metrics"
	systemParts           = `
		SELECT
			database,
			table,
			SUM(bytes) AS bytes,
			COUNT(*)   AS parts,
			SUM(rows)  AS rows 
		FROM system.parts
		WHERE active = 1 
		GROUP BY 
			database, table
		ORDER BY 
			database, table
	`
)

var measurementMap = map[string]string{
	"events":               systemEventsSQL,
	"metrics":              systemMetricsSQL,
	"asynchronous_metrics": systemAsyncMetricsSQL,
}
