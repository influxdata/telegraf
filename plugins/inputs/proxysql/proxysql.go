package proxysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	errNoServersSpecified = fmt.Errorf("no servers specified")

	defaultTimeout = 5 * time.Second
	sampleConfig   = `
  ## specify servers via a url matching:
  ##  username:password@protocol(address)/
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##	servers = ["admin:admin@tcp(127.0.0.1:6032)/"]
  ##	servers = ["admin:admin@unix(/tmp/proxysql_admin.sock)/"]
  ## NOTE: Connection options are not supported
  servers = ["admin:admin@tcp(127.0.0.1:6032)/"]
`
)

const (
	globalStatsQuery    = `SELECT * FROM stats.stats_mysql_global`
	connectionPoolQuery = `		
		SELECT
			hostgroup,
			srv_host,
			srv_port,
			status,
			ConnUsed,
			ConnFree,
			ConnOK,
			ConnERR,
			Queries,
			Bytes_data_sent,
			Bytes_data_recv
		FROM stats.stats_mysql_connection_pool
	`
	commandCounterQuery = `
		SELECT
			Command,
			Total_Time_us,
			Total_cnt,
			cnt_100us,
			cnt_500us,
			cnt_1ms,
			cnt_5ms,
			cnt_10ms,
			cnt_50ms,
			cnt_100ms,
			cnt_500ms,
			cnt_1s,
			cnt_5s,
			cnt_10s,
			cnt_INFs
		FROM stats.stats_mysql_commands_counters
	`
)

type database interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (databaseRows, error)
}

type databaseRows interface {
	Next() bool
	Close() error
	Scan(dest ...interface{}) error
}

type dbWrapper struct {
	database *sql.DB
}

func (d *dbWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (databaseRows, error) {
	return d.database.QueryContext(ctx, query, args...)
}

type ProxySQL struct {
	Servers []string `toml:"servers"`
}

func (p *ProxySQL) SampleConfig() string {
	return sampleConfig
}

func (p *ProxySQL) Description() string {
	return "Read metrics from one or more ProxySQL hosts"
}

func (p *ProxySQL) Gather(acc telegraf.Accumulator) error {
	// Bail out if there are no servers specified
	if len(p.Servers) == 0 {
		return errNoServersSpecified
	}

	// Since we can't get the agent interval, we are going to set a default timeout of 5 seconds
	// for all SQL commands (using a context object)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	wg := &sync.WaitGroup{}
	for _, server := range p.Servers {
		wg.Add(1)
		go func(s string) {
			acc.AddError(p.gatherStats(ctx, acc, s))
			wg.Done()
		}(server)
	}

	// Wait for all the servers to be gathered, this should not lock because we are using
	// a timeout context in our sql commands
	wg.Wait()
	return nil
}

// gatherStats gathers all the stats
func (p *ProxySQL) gatherStats(ctx context.Context, acc telegraf.Accumulator, server string) error {
	dsn, err := mysql.ParseDSN(server)
	if err != nil {
		return err
	}

	// Set a dial timeout
	dsn.Timeout = defaultTimeout

	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return fmt.Errorf("Error opening mysql connection: %v", err)
	}

	defer db.Close()
	wrapper := &dbWrapper{
		database: db,
	}

	// Add server by default to all metrics
	defaultTags := map[string]string{
		"server": dsn.Addr,
	}

	// Gather the global stats
	if err := p.gatherGlobalStats(ctx, wrapper, acc, defaultTags); err != nil {
		return fmt.Errorf("Error gathering global stats: %v", err)
	}

	// Gather the connection pool stats
	if err := p.gatherConnectionPoolStats(ctx, wrapper, acc, defaultTags); err != nil {
		return fmt.Errorf("Error gathering connection pool stats: %v", err)
	}

	// Gather the command counter stats
	if err := p.gatherCommandCounterStats(ctx, wrapper, acc, defaultTags); err != nil {
		return fmt.Errorf("Error gathering command counter stats: %v", err)
	}
	return nil
}

// gatherGlobalStats gathers *all* the global stats ProxySQL provides and lower cases the keys
func (p *ProxySQL) gatherGlobalStats(ctx context.Context, db database, acc telegraf.Accumulator, defaultTags map[string]string) error {
	rows, err := db.QueryContext(ctx, globalStatsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		key string
		val int64
	)

	tags := copyTags(defaultTags)
	fields := map[string]interface{}{}
	for rows.Next() {
		if err := rows.Scan(&key, &val); err != nil {
			return err
		}
		key = strings.ToLower(key)
		fields[key] = val

		// Send 20 fields at a time
		if len(fields) >= 20 {
			acc.AddFields("proxysql", fields, tags)
			fields = map[string]interface{}{}
		}
	}

	if len(fields) > 0 {
		acc.AddFields("proxysql", fields, tags)
	}
	return nil
}

// gatherConnectionPoolStats gathers the connection pool stats, reporting them by hostgroup and each host that ProxySQL is proxying to
func (p *ProxySQL) gatherConnectionPoolStats(ctx context.Context, db database, acc telegraf.Accumulator, defaultTags map[string]string) error {
	rows, err := db.QueryContext(ctx, connectionPoolQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		hostgroup     string
		srvHost       string
		srvPort       string
		status        string
		connUsed      int64
		connFree      int64
		connOK        int64
		connErr       int64
		queries       int64
		bytesDataSent int64
		bytesDataRecv int64
	)

	for rows.Next() {
		if err := rows.Scan(
			&hostgroup, &srvHost, &srvPort, &status, &connUsed, &connFree,
			&connOK, &connErr, &queries, &bytesDataSent, &bytesDataRecv,
		); err != nil {
			return err
		}

		tags := copyTags(defaultTags)
		tags["hostgroup"] = hostgroup
		tags["hostgroup_host"] = fmt.Sprintf("%s:%s", srvHost, srvPort)
		tags["status"] = status

		fields := map[string]interface{}{
			"connections_used": connUsed,
			"connections_free": connFree,
			"connections_ok":   connOK,
			"connections_err":  connErr,
			"queries":          queries,
			"bytes_sent":       bytesDataSent,
			"bytes_received":   bytesDataRecv,
		}
		acc.AddFields("proxysql_connection_pool", fields, tags)
	}
	return nil
}

// gatherCommandCounterStats gathers stats for each type of command, note that the time-based counts are for commands that took at least
// that long but longer than the previous bucket. eg. a 400us query will ONLY be registered in the count_500us field (and the count_total field)
func (p *ProxySQL) gatherCommandCounterStats(ctx context.Context, db database, acc telegraf.Accumulator, defaultTags map[string]string) error {
	rows, err := db.QueryContext(ctx, commandCounterQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		command       string
		totalTime     int64
		totalCount    int64
		count100us    int64
		count500us    int64
		count1ms      int64
		count5ms      int64
		count10ms     int64
		count50ms     int64
		count100ms    int64
		count500ms    int64
		count1s       int64
		count5s       int64
		count10s      int64
		countInfinite int64
	)

	for rows.Next() {
		if err := rows.Scan(
			&command, &totalTime, &totalCount, &count100us, &count500us,
			&count1ms, &count5ms, &count10ms, &count50ms, &count100ms,
			&count500ms, &count1s, &count5s, &count10s, &countInfinite,
		); err != nil {
			return err
		}

		tags := copyTags(defaultTags)
		tags["command"] = command

		fields := map[string]interface{}{
			"total_time":  totalTime,
			"count_total": totalCount,
			"count_100us": count100us,
			"count_500us": count500us,
			"count_1ms":   count1ms,
			"count_5ms":   count5ms,
			"count_10ms":  count10ms,
			"count_50ms":  count50ms,
			"count_100ms": count100ms,
			"count_500ms": count500ms,
			"count_1s":    count1s,
			"count_5s":    count5s,
			"count_10s":   count10s,
			"count_inf":   countInfinite,
		}
		acc.AddFields("proxysql_commands", fields, tags)
	}
	return nil
}

// copyTags copies the default tags so we can modify them
func copyTags(def map[string]string) map[string]string {
	tags := map[string]string{}
	for k, v := range def {
		tags[k] = v
	}
	return tags
}

func init() {
	inputs.Add("proxysql", func() telegraf.Input {
		return &ProxySQL{}
	})
}
