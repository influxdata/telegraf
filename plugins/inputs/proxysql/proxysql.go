package proxysql

import (
	"database/sql"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/go-sql-driver/mysql"
)

type Proxysql struct {
	Servers                             []string `toml:"servers"`
	PerfEventsStatementsDigestTextLimit int64    `toml:"perf_events_statements_digest_text_limit"`
	PerfEventsStatementsLimit           int64    `toml:"perf_events_statements_limit"`
	PerfEventsStatementsTimeLimit       int64    `toml:"perf_events_statements_time_limit"`
	GatherConnectionPool                bool     `toml:"gather_connection_pool"`
	GatherCommandsCounters              bool     `toml:"gather_commands_counters"`
	GatherGlobalStats                   bool     `toml:"gather_global_stats"`
	MetricVersion                       int      `toml:"metric_version"`
}

// metric queries
const (
	connectionPool = `
	SELECT hostgroup,
	       srv_host,
	       srv_port,
	       status,
	       connused,
	       connfree,
	       connok,
	       connerr,
	       queries,
	       bytes_data_sent,
	       bytes_data_recv,
	       latency_us
	FROM stats.stats_mysql_connection_pool`
	commandsCounters = `
	SELECT command,
	       total_time_us,
	       total_cnt,
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
	       cnt_infs
	FROM stats.stats_mysql_commands_counters`
	globalStats = `
	SELECT Variable_Name, Variable_Value
	FROM stats.stats_mysql_global`
)

var localhost = ""
var sampleConfig = `
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    servers = ["user:passwd@tcp(127.0.0.1:6032)/"]
  ##    servers = ["user@tcp(127.0.0.1:6032)/"]
  #
  ## If no servers are specified, then localhost is used as the host.
  servers = ["tcp(127.0.0.1:6032)/"]

  ## Selects the metric output format.
  ##
  ## This option exists to maintain backwards compatibility, if you have
  ## existing metrics do not set or change this value until you are ready to
  ## migrate to the new format.
  ##
  ## If you do not have existing metrics from this plugin set to the latest
  ## version.
  ##
  ## Telegraf >=1.6: metric_version = 2
  ##           <1.6: metric_version = 1 (or unset)
  metric_version = 2

  ## the limits for metrics form perf_events_statements
  perf_events_statements_digest_text_limit  = 120
  perf_events_statements_limit              = 250
  perf_events_statements_time_limit         = 86400
  #
  ## gather metrics from stats.stats_mysql_connection_pool
  gather_connection_pool                    = true
  #
  ## gather metrics from stats.stats_mysql_commands_counters
  gather_commands_counters                  = true
  #
  ## gather metrics from stats.stats_mysql_global
  gather_global_stats                       = true
`

func (p *Proxysql) SampleConfig() string {
	return sampleConfig
}

func (p *Proxysql) Description() string {
	return "Read metrics from one or many proxysql servers"
}

func (p *Proxysql) Gather(acc telegraf.Accumulator) error {
	if len(p.Servers) == 0 {
		// default to localhost if nothing specified.
		return p.gatherServer(localhost, acc)
	}

	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range p.Servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			acc.AddError(p.gatherServer(s, acc))
		}(server)
	}

	wg.Wait()
	return nil
}

func (p *Proxysql) gatherServer(serv string, acc telegraf.Accumulator) error {
	serv, err := dsnAddTimeout(serv)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", serv)
	if err != nil {
		return err
	}

	defer db.Close()

	if p.GatherGlobalStats {
		err = p.gatherGlobalStats(db, serv, acc)
		if err != nil {
			return err
		}
	}

	if p.GatherConnectionPool {
		err = p.gatherConnectionPool(db, serv, acc)
		if err != nil {
			return err
		}
	}

	if p.GatherCommandsCounters {
		err = p.gatherCommandsCounters(db, serv, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Proxysql) gatherGlobalStats(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(globalStats)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		variable_name  string
		variable_value uint64
	)

	servtag := getDSNTag(serv)

	for rows.Next() {
		err = rows.Scan(
			&variable_name,
			&variable_value,
		)

		if err != nil {
			return err
		}

		tags := map[string]string{"server": servtag}

		fields := make(map[string]interface{})
		fields[variable_name] = variable_value

		acc.AddFields("proxysql_global_stats", fields, tags)
	}

	return nil
}

func (p *Proxysql) gatherConnectionPool(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(connectionPool)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		srv_host, status, hostgroup, srv_port        string
		connused, connfree, connok, connerr, queries uint64
		bytes_data_sent, bytes_data_recv, latency_us uint64
	)

	servtag := getDSNTag(serv)

	for rows.Next() {
		err = rows.Scan(
			&hostgroup,
			&srv_host,
			&srv_port,
			&status,
			&connused,
			&connfree,
			&connok,
			&connerr,
			&queries,
			&bytes_data_sent,
			&bytes_data_recv,
			&latency_us,
		)

		if err != nil {
			return err
		}

		tags := map[string]string{
			"server":    servtag,
			"hostgroup": hostgroup,
			"srv_host":  srv_host,
			"srv_port":  srv_port,
		}

		fields := make(map[string]interface{})
		fields["status"] = status
		fields["connused"] = connused
		fields["connfree"] = connfree
		fields["connok"] = connok
		fields["connerr"] = connerr
		fields["queries"] = queries
		fields["bytes_data_sent"] = bytes_data_sent
		fields["bytes_data_recv"] = bytes_data_recv
		fields["latency_us"] = latency_us

		acc.AddFields("proxysql_connection_pool", fields, tags)
	}

	return nil
}

func (p *Proxysql) gatherCommandsCounters(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(commandsCounters)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		command                                                   string
		total_time_us, total_cnt, cnt_100us, cnt_500us, cnt_1ms   uint64
		cnt_5ms, cnt_10ms, cnt_50ms, cnt_100ms, cnt_500ms, cnt_1s uint64
		cnt_5s, cnt_10s, cnt_infs                                 uint64
	)

	servtag := getDSNTag(serv)

	for rows.Next() {
		err = rows.Scan(
			&command,
			&total_time_us,
			&total_cnt,
			&cnt_100us,
			&cnt_500us,
			&cnt_1ms,
			&cnt_5ms,
			&cnt_10ms,
			&cnt_50ms,
			&cnt_100ms,
			&cnt_500ms,
			&cnt_1s,
			&cnt_5s,
			&cnt_10s,
			&cnt_infs,
		)

		if err != nil {
			return err
		}

		tags := map[string]string{
			"server":  servtag,
			"command": command,
		}

		fields := make(map[string]interface{})
		fields["total_time_us"] = total_time_us
		fields["total_cnt"] = total_cnt
		fields["cnt_100us"] = cnt_100us
		fields["cnt_500us"] = cnt_500us
		fields["cnt_1ms"] = cnt_1ms
		fields["cnt_5ms"] = cnt_5ms
		fields["cnt_10ms"] = cnt_10ms
		fields["cnt_50ms"] = cnt_50ms
		fields["cnt_100ms"] = cnt_100ms
		fields["cnt_500ms"] = cnt_500ms
		fields["cnt_1s"] = cnt_1s
		fields["cnt_5s"] = cnt_5s
		fields["cnt_10s"] = cnt_10s
		fields["cnt_infs"] = cnt_infs

		acc.AddFields("proxysql_commands_counters", fields, tags)
	}

	return nil
}

func dsnAddTimeout(dsn string) (string, error) {
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}

	if conf.Timeout == 0 {
		conf.Timeout = time.Second * 5
	}

	return conf.FormatDSN(), nil
}

func getDSNTag(dsn string) string {
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "127.0.0.1:6032"
	}
	return conf.Addr
}

func init() {
	inputs.Add("proxysql", func() telegraf.Input {
		return &Proxysql{}
	})
}
