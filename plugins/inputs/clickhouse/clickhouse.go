package clickhouse

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	_ "github.com/kshvakov/clickhouse"
)

const sampleConfig = `
  ### ClickHouse DSN
  dsn     = "native://localhost:9000?username=user&password=qwerty"
  cluster = false
  ignored_clusters = ["test_shard_localhost"]
`

type connect struct {
	*sql.DB
	cluster, shardNum, hostname string
}

// ClickHouse Telegraf Input Plugin
type ClickHouse struct {
	DSN             string   `toml:"dsn"`
	Cluster         bool     `toml:"cluster"`
	IgnoredClusters []string `toml:"ignored_clusters"`
	connect         *connect
	clustersConn    map[string]*connect
}

// SampleConfig returns the sample config
func (*ClickHouse) SampleConfig() string {
	return sampleConfig
}

// Description return plugin description
func (*ClickHouse) Description() string {
	return "Read metrics from ClickHouse server"
}

// Start ClickHouse input service
func (ch *ClickHouse) Start(telegraf.Accumulator) (err error) {
	if ch.connect.DB, err = sql.Open("clickhouse", ch.DSN); err != nil {
		return err
	}
	setConnLimits(ch.connect.DB)
	return ch.connect.QueryRow("SELECT hostName()").Scan(&ch.connect.hostname)
}

// Stop ClickHouse input service
func (ch *ClickHouse) Stop() {
	if ch.connect != nil {
		ch.connect.Close()
	}
	for _, conn := range ch.clustersConn {
		conn.Close()
	}
}

// Gather collect data from ClickHouse server
func (ch *ClickHouse) Gather(acc telegraf.Accumulator) (err error) {
	if !ch.Cluster {
		if err := ch.gather(ch.connect, acc); err != nil {
			acc.AddError(err)
			return err
		}
		return nil
	}
	conns, err := ch.conns(acc)
	if err != nil {
		acc.AddError(err)
		return err
	}
	for _, conn := range conns {
		if err := ch.gather(conn, acc); err != nil {
			acc.AddError(err)
		}
	}
	return nil
}

func (ch *ClickHouse) gather(conn *connect, acc telegraf.Accumulator) (err error) {
	var rows *sql.Rows
	for measurement, query := range measurementMap {
		if rows, err = conn.Query(query); err != nil {
			return err
		}
		if err := ch.processRows(measurement, conn, rows, acc); err != nil {
			return err
		}
	}
	if rows, err = conn.Query(systemParts); err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			table, database           string
			bytes, parts, rowsInTable uint64
		)
		if err := rows.Scan(&database, &table, &bytes, &parts, &rowsInTable); err != nil {
			return err
		}
		tags := map[string]string{
			"table":    table,
			"database": database,
			"hostname": conn.hostname,
		}
		if len(conn.cluster) != 0 {
			tags["cluster"] = conn.cluster
		}
		if len(conn.shardNum) != 0 {
			tags["shard_num"] = conn.shardNum
		}
		acc.AddFields("clickhouse_tables",
			map[string]interface{}{
				"bytes": bytes,
				"parts": parts,
				"rows":  rowsInTable,
			},
			tags,
		)
	}
	return nil
}

func (ch *ClickHouse) conns(acc telegraf.Accumulator) ([]*connect, error) {
	var (
		ignore = func(cluster string) bool {
			for _, ignored := range ch.IgnoredClusters {
				if cluster == ignored {
					return true
				}
			}
			return false
		}
		rows, err = ch.connect.Query(systemClustersSQL)
	)
	if err != nil {
		return nil, err
	}
	baseDSN, err := url.Parse(ch.DSN)
	if err != nil {
		return nil, err
	}
	baseQuery := baseDSN.Query()
	{
		baseQuery.Del("alt_hosts")
		baseDSN.RawQuery = baseQuery.Encode()
	}
	for rows.Next() {
		var (
			port, shardNum             int
			hostname, address, cluster string
		)
		if err := rows.Scan(&cluster, &shardNum, &hostname, &address, &port); err != nil {
			acc.AddError(err)
			continue
		}
		connID := fmt.Sprintf("%s_%d", address, shardNum)
		if _, found := ch.clustersConn[connID]; !found {
			if ignore(cluster) {
				continue
			}
			baseDSN.Host = net.JoinHostPort(address, strconv.Itoa(port))
			conn, err := sql.Open("clickhouse", baseDSN.String())
			if err != nil {
				acc.AddError(err)
				continue
			}
			setConnLimits(conn)
			ch.clustersConn[connID] = &connect{
				DB:       conn,
				cluster:  cluster,
				shardNum: strconv.Itoa(shardNum),
				hostname: hostname,
			}
		}
	}
	conns := make([]*connect, 0, len(ch.clustersConn))
	for _, conn := range ch.clustersConn {
		if err := conn.Ping(); err == nil {
			conns = append(conns, conn)
		}
	}
	return conns, nil
}

func (ch *ClickHouse) processRows(measurement string, conn *connect, rows *sql.Rows, acc telegraf.Accumulator) error {
	defer rows.Close()
	fields := make(map[string]interface{})
	for rows.Next() {
		var (
			key   string
			value uint64
		)
		if err := rows.Scan(&key, &value); err != nil {
			return err
		}
		fields[key] = value
	}
	tags := map[string]string{
		"hostname": conn.hostname,
	}
	if len(conn.cluster) != 0 {
		tags["cluster"] = conn.cluster
	}
	if len(conn.shardNum) != 0 {
		tags["shard_num"] = conn.shardNum
	}
	acc.AddFields("clickhouse_"+measurement, fields, tags)
	return nil
}

func init() {
	inputs.Add("clickhouse", func() telegraf.Input {
		return &ClickHouse{
			connect:      &connect{},
			clustersConn: make(map[string]*connect),
		}
	})
}

func setConnLimits(conn *sql.DB) {
	conn.SetMaxOpenConns(2)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(10 * time.Minute)
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
	systemClustersSQL = `
		SELECT
			cluster,
			shard_num,
			host_name,
			host_address,
			port
		FROM system.clusters`
)

var measurementMap = map[string]string{
	"events":               systemEventsSQL,
	"metrics":              systemMetricsSQL,
	"asynchronous_metrics": systemAsyncMetricsSQL,
}

var _ telegraf.ServiceInput = &ClickHouse{}
