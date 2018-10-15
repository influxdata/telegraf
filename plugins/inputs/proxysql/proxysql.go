package proxysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/go-sql-driver/mysql"
)

type Proxysql struct {
	Servers                []string `toml:"servers"`
	GatherGlobal           bool     `toml:"gather_global"`
	GatherCommandsCounters bool     `toml:"gather_commands_counters"`
	GatherConnectionPool   bool     `toml:"gather_connection_pool"`
	GatherUsers            bool     `toml:"gather_users"`
	GatherRules            bool     `toml:"gather_rules"`
	GatherQueries          bool     `toml:"gather_queries"`
	GatherMemoryMetrics    bool     `toml:"gather_memory_metrics"`
	GatherProcessList      bool     `toml:"gather_process_list"`
	IntervalSlow           string   `toml:"interval_slow"`
	tls.ClientConfig
}

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify|custom]]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    servers = ["user:passwd@tcp(127.0.0.1:6032)/?tls=false"]
  ##    servers = ["user@tcp(127.0.0.1:6032)/?tls=false"]
  #
  ## If no servers are specified, then localhost is used as the host.
  servers = ["tcp(127.0.0.1:6032)/"]
  #
  ## gather metrics from stats_mysql_global
  gather_global                             = true  
  ## gather metrics from stats_mysql_commands_counters
  gather_commands_counters                  = true
  ## gather metrics from stats_mysql_connection_pool
  gather_connection_pool                    = true
  ## gather metrics from stats_mysql_users
  gather_users                              = true
  ## gather metrics from stats_mysql_query_rules
  gather_rules                              = true
  ## gather metrics from stats_mysql_query_digest
  gather_queries                            = true
  ## gather metrics from stats_memory_metrics
  gather_memory_metrics                     = true
  ## gather thread state counts from stats_mysql_processlist
  gather_process_list                       = true
  #
  ## Some queries we may want to run less often (such as SHOW GLOBAL VARIABLES)
  interval_slow                             = "30m"

  ## Optional TLS Config (will be used if tls=custom parameter specified in server uri)
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

var defaultTimeout = time.Second * time.Duration(5)

func (m *Proxysql) SampleConfig() string {
	return sampleConfig
}

func (m *Proxysql) Description() string {
	return "Read metrics from one or many proxysql servers"
}

var (
	localhost        = "tcp(127.0.0.1:6032)/"
	lastT            time.Time
	initDone         = false
	scanIntervalSlow uint32
)

func (m *Proxysql) InitProxysql() {
	if len(m.IntervalSlow) > 0 {
		interval, err := time.ParseDuration(m.IntervalSlow)
		if err == nil && interval.Seconds() >= 1.0 {
			scanIntervalSlow = uint32(interval.Seconds())
		}
	}
	initDone = true
}

func (m *Proxysql) Gather(acc telegraf.Accumulator) error {
	if len(m.Servers) == 0 {
		// default to localhost if nothing specified.
		return m.gatherServer(localhost, acc)
	}
	// Initialise additional query intervals
	if !initDone {
		m.InitProxysql()
	}

	tlsConfig, err := m.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("registering TLS config: %s", err)
	}

	if tlsConfig != nil {
		mysql.RegisterTLSConfig("custom", tlsConfig)
	}

	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range m.Servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			acc.AddError(m.gatherServer(s, acc))
		}(server)
	}

	wg.Wait()
	return nil
}

// metric queries
const (
	globalQuery          = `SELECT Variable_Name, Variable_Value FROM stats_mysql_global`
	globalVariablesQuery = `SELECT variable_name, variable_value FROM global_variables`
	commandsCounterQuery = `SELECT Command, Total_cnt FROM stats_mysql_commands_counters`
	connectionPoolQuery  = `SELECT hostgroup, srv_host, srv_port, ConnUsed, ConnFree, ConnOK, ConnERR, Queries, Bytes_data_sent, Bytes_data_recv, Latency_us FROM stats_mysql_connection_pool`
	usersQuery           = `SELECT username, frontend_connections, frontend_max_connections FROM stats_mysql_users`
	rulesQuery           = `SELECT rule_id, hits FROM stats_mysql_query_rules`
	queriesQuery         = `SELECT hostgroup, schemaname, username, digest, digest_text, count_star, first_seen, last_seen, sum_time, min_time, max_time FROM stats_mysql_query_digest`
	memoryMetricsQuery   = `SELECT Variable_Name, Variable_Value FROM stats_memory_metrics`
	processListQuery     = `SELECT COALESCE(command,''),count(*) FROM stats_mysql_processlist GROUP BY command`
)

func (m *Proxysql) gatherServer(serv string, acc telegraf.Accumulator) error {
	serv, err := dsnAddTimeout(serv)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", serv)
	if err != nil {
		return err
	}

	defer db.Close()

	if m.GatherGlobal {
		err = m.gatherVariables(db, serv, acc, globalQuery, "proxysql")
		if err != nil {
			return err
		}
	}

	// Global Variables may be gathered less often
	if len(m.IntervalSlow) > 0 {
		if uint32(time.Since(lastT).Seconds()) >= scanIntervalSlow {
			err = m.gatherVariables(db, serv, acc, globalVariablesQuery, "proxysql_variables")
			if err != nil {
				return err
			}
			lastT = time.Now()
		}
	}

	if m.GatherCommandsCounters {
		err = m.gatherCommandsCounters(db, serv, acc)
		if err != nil {
			return err
		}
	}

	if m.GatherConnectionPool {
		err = m.gatherConnectionPool(db, serv, acc)
		if err != nil {
			return err
		}
	}

	if m.GatherUsers {
		err = m.gatherUsers(db, serv, acc)
		if err != nil {
			return err
		}
	}

	if m.GatherRules {
		err = m.gatherRules(db, serv, acc)
		if err != nil {
			return err
		}
	}

	if m.GatherQueries {
		err = m.gatherQueries(db, serv, acc)
		if err != nil {
			return err
		}
	}

	if m.GatherMemoryMetrics {
		err = m.gatherVariables(db, serv, acc, memoryMetricsQuery, "proxysql_memory")
		if err != nil {
			return err
		}
	}

	if m.GatherProcessList {
		err = m.gatherProcessListStatuses(db, serv, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

// gatherGlobalVariables can be used to fetch all global variables from
// MySQL environment.
func (m *Proxysql) gatherVariables(db *sql.DB, serv string, acc telegraf.Accumulator, query string, measurement string) error {
	// run query
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var key string
	var val sql.RawBytes

	// parse DSN and save server tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})
	for rows.Next() {
		if err := rows.Scan(&key, &val); err != nil {
			return err
		}
		key = strings.ToLower(key)
		// parse mysql version and put into field and tag
		if strings.Contains(key, "version") {
			fields[key] = string(val)
			tags[key] = string(val)
		}
		if value, ok := parseValue(val); ok {
			fields[key] = value
		}
		// Send 20 fields at a time
		if len(fields) >= 20 {
			acc.AddFields(measurement, fields, tags)
			fields = make(map[string]interface{})
		}
	}
	// Send any remaining fields
	if len(fields) > 0 {
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}

func (m *Proxysql) gatherCommandsCounters(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	// run query
	rows, err := db.Query(commandsCounterQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var command string
	var count uint32

	// parse DSN and save server tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})
	for rows.Next() {
		if err := rows.Scan(&command, &count); err != nil {
			return err
		}
		command = strings.ToLower(command)
		fields[newNamespace("commands", command)] = count
		// Send 20 fields at a time
		if len(fields) >= 20 {
			acc.AddFields("proxysql_commands", fields, tags)
			fields = make(map[string]interface{})
		}
	}
	// Send any remaining fields
	if len(fields) > 0 {
		acc.AddFields("proxysql_commands", fields, tags)
	}
	return nil
}

func (m *Proxysql) gatherConnectionPool(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	// run query
	rows, err := db.Query(connectionPoolQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var hostgroup string
	var host string
	var port string
	var connUsed uint64
	var connFree uint64
	var connOk uint64
	var connErr uint64
	var queries uint64
	var bytesDataSent uint64
	var bytesDataRecv uint64
	var latency uint64

	// parse DSN and save server tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})

	for rows.Next() {
		if err := rows.Scan(&hostgroup, &host, &port, &connUsed, &connFree, &connOk, &connErr, &queries, &bytesDataSent, &bytesDataRecv, &latency); err != nil {
			return err
		}
		tags["hostgroup"] = hostgroup
		tags["connection"] = strings.Join([]string{host, port}, ":")
		fields["connections_used"] = connUsed
		fields["connections_free"] = connFree
		fields["connections_ok"] = connOk
		fields["connections_err"] = connErr
		fields["queries"] = queries
		fields["bytes_data_sent"] = bytesDataSent
		fields["bytes_data_recv"] = bytesDataRecv
		fields["latency"] = latency

		acc.AddFields("proxysql_connection_pool", fields, tags)
	}
	return nil
}

func (m *Proxysql) gatherUsers(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	// run query
	rows, err := db.Query(usersQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var username string
	var frontendConnections uint64
	var frontendMaxConnections uint64

	// parse DSN and save server tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})

	for rows.Next() {
		if err := rows.Scan(&username, &frontendConnections, &frontendMaxConnections); err != nil {
			return err
		}
		tags["user"] = username
		fields["frontend_connections"] = frontendConnections
		fields["frontend_max_connections"] = frontendMaxConnections

		acc.AddFields("proxysql_users", fields, tags)
	}
	return nil
}

func (m *Proxysql) gatherRules(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	// run query
	rows, err := db.Query(rulesQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ruleId string
	var hits uint64

	// parse DSN and save server tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})

	for rows.Next() {
		if err := rows.Scan(&ruleId, &hits); err != nil {
			return err
		}
		tags["rule_id"] = ruleId
		fields["hits"] = hits

		acc.AddFields("proxysql_rules", fields, tags)
	}
	return nil
}

func (m *Proxysql) gatherQueries(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	// run query
	rows, err := db.Query(queriesQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var hostgroup string
	var schemaName string
	var username string
	var digest string
	var digestText string
	var countStar uint64
	var firstSeen uint64
	var lastSeen uint64
	var sumTime uint64
	var minTime uint64
	var maxTime uint64

	// parse DSN and save server tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})

	for rows.Next() {
		if err := rows.Scan(&hostgroup, &schemaName, &username, &digest, &digestText, &countStar, &firstSeen, &lastSeen, &sumTime, &minTime, &maxTime); err != nil {
			return err
		}
		tags["hostgroup"] = hostgroup
		tags["schema_name"] = schemaName
		tags["user"] = username
		tags["digest"] = digest
		fields["digest_text"] = digestText
		fields["count_star"] = countStar
		fields["first_seen"] = firstSeen
		fields["last_seen"] = lastSeen
		fields["sum_time"] = sumTime
		fields["min_time"] = minTime
		fields["max_time"] = maxTime

		acc.AddFields("proxysql_queries", fields, tags)
	}
	return nil
}

// GatherProcessList can be used to collect metrics on each running command
// and its state with its running count
func (m *Proxysql) gatherProcessListStatuses(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	// run query
	rows, err := db.Query(processListQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var servtag string
	servtag = getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})

	for rows.Next() {
		var (
			command string
			count   uint32
		)

		err = rows.Scan(&command, &count)
		if err != nil {
			return err
		}

		command = strings.ToLower(command)
		fields[newNamespace("threads", command)] = count
		// Send 20 fields at a time
		if len(fields) >= 20 {
			acc.AddFields("proxysql_process_list", fields, tags)
			fields = make(map[string]interface{})
		}
	}
	// Send any remaining fields
	if len(fields) > 0 {
		acc.AddFields("proxysql_process_list", fields, tags)
	}

	return nil
}

// parseValue can be used to convert values such as "ON","OFF","Yes","No" to 0,1
func parseValue(value sql.RawBytes) (interface{}, bool) {
	if bytes.EqualFold(value, []byte("YES")) || bytes.Compare(value, []byte("ON")) == 0 {
		return 1, true
	}

	if bytes.EqualFold(value, []byte("NO")) || bytes.Compare(value, []byte("OFF")) == 0 {
		return 0, true
	}

	if val, err := strconv.ParseInt(string(value), 10, 64); err == nil {
		return val, true
	}
	if val, err := strconv.ParseFloat(string(value), 64); err == nil {
		return val, true
	}

	if len(string(value)) > 0 {
		return string(value), true
	}
	return nil, false
}

// newNamespace can be used to make a namespace
func newNamespace(words ...string) string {
	return strings.Replace(strings.Join(words, "_"), " ", "_", -1)
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
		return &Proxysql{
			GatherGlobal: true,
		}
	})
}
