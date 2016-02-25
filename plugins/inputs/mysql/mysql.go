package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Mysql struct {
	Servers                             []string
	PerfEventsStatementsDigestTextLimit uint32
	PerfEventsStatementsLimit           uint32
	PerfEventsStatementsTimeLimit       uint32
	TableSchemaDatabases                []string
	GatherSlaveStatus                   bool
	GatherBinaryLogs                    bool
	GatherTableIOWaits                  bool
	GatherIndexIOWaits                  bool
	GatherTableSchema                   bool
}

var sampleConfig = `
  # specify servers via a url matching:
  #  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify]]
  #  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  #  e.g.
  #    root:passwd@tcp(127.0.0.1:3306)/?tls=false
  #    root@tcp(127.0.0.1:3306)/?tls=false
  #
  # If no servers are specified, then localhost is used as the host.
  servers = ["tcp(127.0.0.1:3306)/"]
  PerfEventsStatementsDigestTextLimit = 120
  PerfEventsStatementsLimit           = 250
  PerfEventsStatementsTimeLimit       = 86400
  TableSchemaDatabases                = []
  GatherSlaveStatus                   = false
  GatherBinaryLogs                    = false
  GatherTableIOWaits                  = false
  GatherIndexIOWaits                  = false
  GatherTableSchema                   = false
`

var defaultTimeout = time.Second * time.Duration(5)

func (m *Mysql) SampleConfig() string {
	return sampleConfig
}

func (m *Mysql) Description() string {
	return "Read metrics from one or many mysql servers"
}

var localhost = ""

func (m *Mysql) Gather(acc telegraf.Accumulator) error {
	if len(m.Servers) == 0 {
		// if we can't get stats in this case, thats fine, don't report
		// an error.
		m.gatherServer(localhost, acc)
		return nil
	}

	for _, serv := range m.Servers {
		err := m.gatherServer(serv, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

type mapping struct {
	onServer string
	inExport string
}

var mappings = []*mapping{
	{
		onServer: "Aborted_",
		inExport: "aborted_",
	},
	{
		onServer: "Bytes_",
		inExport: "bytes_",
	},
	{
		onServer: "Com_",
		inExport: "commands_",
	},
	{
		onServer: "Created_",
		inExport: "created_",
	},
	{
		onServer: "Handler_",
		inExport: "handler_",
	},
	{
		onServer: "Innodb_",
		inExport: "innodb_",
	},
	{
		onServer: "Key_",
		inExport: "key_",
	},
	{
		onServer: "Open_",
		inExport: "open_",
	},
	{
		onServer: "Opened_",
		inExport: "opened_",
	},
	{
		onServer: "Qcache_",
		inExport: "qcache_",
	},
	{
		onServer: "Table_",
		inExport: "table_",
	},
	{
		onServer: "Tokudb_",
		inExport: "tokudb_",
	},
	{
		onServer: "Threads_",
		inExport: "threads_",
	},
}

var (
	// status counter
	generalThreadStates = map[string]uint32{
		"after create":              uint32(0),
		"altering table":            uint32(0),
		"analyzing":                 uint32(0),
		"checking permissions":      uint32(0),
		"checking table":            uint32(0),
		"cleaning up":               uint32(0),
		"closing tables":            uint32(0),
		"converting heap to myisam": uint32(0),
		"copying to tmp table":      uint32(0),
		"creating sort index":       uint32(0),
		"creating table":            uint32(0),
		"creating tmp table":        uint32(0),
		"deleting":                  uint32(0),
		"executing":                 uint32(0),
		"execution of init_command": uint32(0),
		"end":                     uint32(0),
		"freeing items":           uint32(0),
		"flushing tables":         uint32(0),
		"fulltext initialization": uint32(0),
		"idle":                      uint32(0),
		"init":                      uint32(0),
		"killed":                    uint32(0),
		"waiting for lock":          uint32(0),
		"logging slow query":        uint32(0),
		"login":                     uint32(0),
		"manage keys":               uint32(0),
		"opening tables":            uint32(0),
		"optimizing":                uint32(0),
		"preparing":                 uint32(0),
		"reading from net":          uint32(0),
		"removing duplicates":       uint32(0),
		"removing tmp table":        uint32(0),
		"reopen tables":             uint32(0),
		"repair by sorting":         uint32(0),
		"repair done":               uint32(0),
		"repair with keycache":      uint32(0),
		"replication master":        uint32(0),
		"rolling back":              uint32(0),
		"searching rows for update": uint32(0),
		"sending data":              uint32(0),
		"sorting for group":         uint32(0),
		"sorting for order":         uint32(0),
		"sorting index":             uint32(0),
		"sorting result":            uint32(0),
		"statistics":                uint32(0),
		"updating":                  uint32(0),
		"waiting for tables":        uint32(0),
		"waiting for table flush":   uint32(0),
		"waiting on cond":           uint32(0),
		"writing to net":            uint32(0),
		"other":                     uint32(0),
	}
	// plaintext statuses
	stateStatusMappings = map[string]string{
		"user sleep":                               "idle",
		"creating index":                           "altering table",
		"committing alter table to storage engine": "altering table",
		"discard or import tablespace":             "altering table",
		"rename":                                   "altering table",
		"setup":                                    "altering table",
		"renaming result table":                    "altering table",
		"preparing for alter table":                "altering table",
		"copying to group table":                   "copying to tmp table",
		"copy to tmp table":                        "copying to tmp table",
		"query end":                                "end",
		"update":                                   "updating",
		"updating main table":                      "updating",
		"updating reference tables":                "updating",
		"system lock":                              "waiting for lock",
		"user lock":                                "waiting for lock",
		"table lock":                               "waiting for lock",
		"deleting from main table":                 "deleting",
		"deleting from reference tables":           "deleting",
	}
)

func dsnAddTimeout(dsn string) (string, error) {

	// DSN "?timeout=5s" is not valid, but "/?timeout=5s" is valid ("" and "/"
	// are the same DSN)
	if dsn == "" {
		dsn = "/"
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	v := u.Query()

	// Only override timeout if not already defined
	if _, ok := v["timeout"]; ok == false {
		v.Add("timeout", defaultTimeout.String())
		u.RawQuery = v.Encode()
	}
	return u.String(), nil
}

// Math constants
const (
	picoSeconds = 1e12
)

// metric queries
const (
	globalStatusQuery          = `SHOW GLOBAL STATUS`
	globalVariablesQuery       = `SHOW GLOBAL VARIABLES`
	slaveStatusQuery           = `SHOW SLAVE STATUS`
	binaryLogsQuery            = `SHOW BINARY LOGS`
	infoSchemaProcessListQuery = `
        SELECT COALESCE(command,''),COALESCE(state,''),count(*)
        FROM information_schema.processlist
        WHERE ID != connection_id()
        GROUP BY command,state
        ORDER BY null`
	infoSchemaAutoIncQuery = `
        SELECT table_schema, table_name, column_name, auto_increment,
          pow(2, case data_type
            when 'tinyint'   then 7
            when 'smallint'  then 15
            when 'mediumint' then 23
            when 'int'       then 31
            when 'bigint'    then 63
            end+(column_type like '% unsigned'))-1 as max_int
          FROM information_schema.tables t
          JOIN information_schema.columns c USING (table_schema,table_name)
          WHERE c.extra = 'auto_increment' AND t.auto_increment IS NOT NULL
    `
	perfTableIOWaitsQuery = `
        SELECT OBJECT_SCHEMA, OBJECT_NAME, COUNT_FETCH, COUNT_INSERT, COUNT_UPDATE, COUNT_DELETE,
        SUM_TIMER_FETCH, SUM_TIMER_INSERT, SUM_TIMER_UPDATE, SUM_TIMER_DELETE
        FROM performance_schema.table_io_waits_summary_by_table
        WHERE OBJECT_SCHEMA NOT IN ('mysql', 'performance_schema')
    `
	perfIndexIOWaitsQuery = `
        SELECT OBJECT_SCHEMA, OBJECT_NAME, ifnull(INDEX_NAME, 'NONE') as INDEX_NAME,
        COUNT_FETCH, COUNT_INSERT, COUNT_UPDATE, COUNT_DELETE,
        SUM_TIMER_FETCH, SUM_TIMER_INSERT, SUM_TIMER_UPDATE, SUM_TIMER_DELETE
        FROM performance_schema.table_io_waits_summary_by_index_usage
        WHERE OBJECT_SCHEMA NOT IN ('mysql', 'performance_schema')
    `
	perfTableLockWaitsQuery = `
        SELECT
            OBJECT_SCHEMA,
            OBJECT_NAME,
            COUNT_READ_NORMAL,
            COUNT_READ_WITH_SHARED_LOCKS,
            COUNT_READ_HIGH_PRIORITY,
            COUNT_READ_NO_INSERT,
            COUNT_READ_EXTERNAL,
            COUNT_WRITE_ALLOW_WRITE,
            COUNT_WRITE_CONCURRENT_INSERT,
            COUNT_WRITE_DELAYED,
            COUNT_WRITE_LOW_PRIORITY,
            COUNT_WRITE_NORMAL,
            COUNT_WRITE_EXTERNAL,
            SUM_TIMER_READ_NORMAL,
            SUM_TIMER_READ_WITH_SHARED_LOCKS,
            SUM_TIMER_READ_HIGH_PRIORITY,
            SUM_TIMER_READ_NO_INSERT,
            SUM_TIMER_READ_EXTERNAL,
            SUM_TIMER_WRITE_ALLOW_WRITE,
            SUM_TIMER_WRITE_CONCURRENT_INSERT,
            SUM_TIMER_WRITE_DELAYED,
            SUM_TIMER_WRITE_LOW_PRIORITY,
            SUM_TIMER_WRITE_NORMAL,
            SUM_TIMER_WRITE_EXTERNAL
        FROM performance_schema.table_lock_waits_summary_by_table
        WHERE OBJECT_SCHEMA NOT IN ('mysql', 'performance_schema', 'information_schema')
    `
	perfEventsStatementsQuery = `
        SELECT
            ifnull(SCHEMA_NAME, 'NONE') as SCHEMA_NAME,
            DIGEST,
            LEFT(DIGEST_TEXT, %d) as DIGEST_TEXT,
            COUNT_STAR,
            SUM_TIMER_WAIT,
            SUM_ERRORS,
            SUM_WARNINGS,
            SUM_ROWS_AFFECTED,
            SUM_ROWS_SENT,
            SUM_ROWS_EXAMINED,
            SUM_CREATED_TMP_DISK_TABLES,
            SUM_CREATED_TMP_TABLES,
            SUM_SORT_MERGE_PASSES,
            SUM_SORT_ROWS,
            SUM_NO_INDEX_USED
        FROM performance_schema.events_statements_summary_by_digest
        WHERE SCHEMA_NAME NOT IN ('mysql', 'performance_schema', 'information_schema')
            AND last_seen > DATE_SUB(NOW(), INTERVAL %d SECOND)
        ORDER BY SUM_TIMER_WAIT DESC
        LIMIT %d
    `
	perfEventWaitsQuery = `
        SELECT EVENT_NAME, COUNT_STAR, SUM_TIMER_WAIT
        FROM performance_schema.events_waits_summary_global_by_event_name
    `
	perfFileEventsQuery = `
        SELECT
            EVENT_NAME,
            COUNT_READ, SUM_TIMER_READ, SUM_NUMBER_OF_BYTES_READ,
            COUNT_WRITE, SUM_TIMER_WRITE, SUM_NUMBER_OF_BYTES_WRITE,
            COUNT_MISC, SUM_TIMER_MISC
        FROM performance_schema.file_summary_by_event_name
    `
	tableSchemaQuery = `
        SELECT
            TABLE_SCHEMA,
            TABLE_NAME,
            TABLE_TYPE,
            ifnull(ENGINE, 'NONE') as ENGINE,
            ifnull(VERSION, '0') as VERSION,
            ifnull(ROW_FORMAT, 'NONE') as ROW_FORMAT,
            ifnull(TABLE_ROWS, '0') as TABLE_ROWS,
            ifnull(DATA_LENGTH, '0') as DATA_LENGTH,
            ifnull(INDEX_LENGTH, '0') as INDEX_LENGTH,
            ifnull(DATA_FREE, '0') as DATA_FREE,
            ifnull(CREATE_OPTIONS, 'NONE') as CREATE_OPTIONS
        FROM information_schema.tables
        WHERE TABLE_SCHEMA = '%s'
    `
	dbListQuery = `
        SELECT
            SCHEMA_NAME
            FROM information_schema.schemata
        WHERE SCHEMA_NAME NOT IN ('mysql', 'performance_schema', 'information_schema')
    `
	perfSchemaTablesQuery = `
		SELECT
			table_name
			FROM information_schema.tables
		WHERE table_schema = 'performance_schema' AND table_name = ?
	`
)

func (m *Mysql) gatherServer(serv string, acc telegraf.Accumulator) error {
	serv, err := dsnAddTimeout(serv)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", serv)
	if err != nil {
		return err
	}

	defer db.Close()

	err = m.gatherGlobalStatuses(db, serv, acc)
	if err != nil {
		return err
	}

	err = m.gatherGlobalVariables(db, serv, acc)
	if err != nil {
		return err
	}

	if m.GatherSlaveStatus {
		err = m.gatherBinaryLogs(db, serv, acc)
		if err != nil {
			return err
		}
	}

	err = m.GatherProcessListStatuses(db, serv, acc)
	if err != nil {
		return err
	}

	if m.GatherSlaveStatus {
		err = m.gatherSlaveStatuses(db, serv, acc)
		if err != nil {
			return err
		}
	}

	err = m.gatherInfoSchemaAutoIncStatuses(db, serv, acc)
	if err != nil {
		return err
	}

	if m.GatherTableIOWaits {
		err = m.gatherPerfTableIOWaits(db, serv, acc)
		if err != nil {
			return err
		}
	}

	if m.GatherIndexIOWaits {
		err = m.gatherPerfIndexIOWaits(db, serv, acc)
		if err != nil {
			return err
		}
	}

	err = m.gatherPerfTableLockWaits(db, serv, acc)
	if err != nil {
		return err
	}

	err = m.gatherPerfEventWaits(db, serv, acc)
	if err != nil {
		return err
	}

	err = m.gatherPerfFileEventsStatuses(db, serv, acc)
	if err != nil {
		return err
	}

	err = m.gatherPerfEventsStatements(db, serv, acc)
	if err != nil {
		return err
	}

	if m.GatherTableSchema {
		err = m.gatherTableSchema(db, serv, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Mysql) gatherGlobalVariables(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(globalVariablesQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	var key string
	var val sql.RawBytes

	servtag, err := parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})
	for rows.Next() {
		if err := rows.Scan(&key, &val); err != nil {
			return err
		}
		key = strings.ToLower(key)
		if floatVal, ok := parseValue(val); ok {
			fields[key] = floatVal
		}
	}
	acc.AddFields("mysql_variables", fields, tags)
	return nil
}

// gatherSlaveStatuses can be used to get replication analytics
// When the server is slave, then it returns only one row.
// If the multi-source replication is set, then everything works differently
// This code does not work with multi-source replication.
func (m *Mysql) gatherSlaveStatuses(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(slaveStatusQuery)

	if err != nil {
		return err
	}
	defer rows.Close()

	servtag, err := parseDSN(serv)

	if err != nil {
		servtag = "localhost"
	}
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})
	if rows.Next() {
		cols, err := rows.Columns()

		if err != nil {
			return err
		}
		vals := make([]interface{}, len(cols))

		for i := range vals {
			vals[i] = &sql.RawBytes{}
		}

		if err = rows.Scan(vals...); err != nil {
			return err
		}

		for i, col := range cols {
			// skip unparsable values
			if value, ok := parseValue(*vals[i].(*sql.RawBytes)); ok {
				//acc.Add("slave_"+col, value, tags)
				fields["slave_"+col] = value
			}
		}
		acc.AddFields("mysql", fields, tags)
	}

	return nil
}

// gatherBinaryLogs can be used to collect size and count of all binary files
func (m *Mysql) gatherBinaryLogs(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(binaryLogsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var servtag string
	servtag, err = parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})
	var (
		size     uint64 = 0
		count    uint64 = 0
		fileSize uint64
		fileName string
	)

	for rows.Next() {
		if err := rows.Scan(&fileName, &fileSize); err != nil {
			return err
		}
		size += fileSize
		count++
	}
	fields["binary_size_bytes"] = size
	fields["binary_files_count"] = count
	acc.AddFields("mysql", fields, tags)
	return nil
}

func (m *Mysql) gatherGlobalStatuses(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	// If user forgot the '/', add it
	if strings.HasSuffix(serv, ")") {
		serv = serv + "/"
	} else if serv == "localhost" {
		serv = ""
	}

	rows, err := db.Query(globalStatusQuery)
	if err != nil {
		return err
	}

	var servtag string
	servtag, err = parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})
	for rows.Next() {
		var name string
		var val interface{}

		err = rows.Scan(&name, &val)
		if err != nil {
			return err
		}

		var found bool

		for _, mapped := range mappings {
			if strings.HasPrefix(name, mapped.onServer) {
				i, _ := strconv.Atoi(string(val.([]byte)))
				fields[mapped.inExport+name[len(mapped.onServer):]] = i
				found = true
			}
		}

		if found {
			continue
		}

		switch name {
		case "Queries":
			i, err := strconv.ParseInt(string(val.([]byte)), 10, 64)
			if err != nil {
				return err
			}

			fields["queries"] = i
		case "Slow_queries":
			i, err := strconv.ParseInt(string(val.([]byte)), 10, 64)
			if err != nil {
				return err
			}

			fields["slow_queries"] = i
		}
	}
	acc.AddFields("mysql", fields, tags)

	conn_rows, err := db.Query("SELECT user, sum(1) FROM INFORMATION_SCHEMA.PROCESSLIST GROUP BY user")

	for conn_rows.Next() {
		var user string
		var connections int64

		err = conn_rows.Scan(&user, &connections)
		if err != nil {
			return err
		}

		tags := map[string]string{"server": servtag, "user": user}
		fields := make(map[string]interface{})

		if err != nil {
			return err
		}
		fields["connections"] = connections
		acc.AddFields("mysql_users", fields, tags)
	}

	return nil
}

func (m *Mysql) GatherProcessListStatuses(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(infoSchemaProcessListQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	var (
		command string
		state   string
		count   uint32
	)

	var servtag string
	fields := make(map[string]interface{})
	servtag, err = parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}

	// mapping of state with its counts
	stateCounts := make(map[string]uint32, len(generalThreadStates))
	// set map with keys and default values
	for k, v := range generalThreadStates {
		stateCounts[k] = v
	}

	for rows.Next() {
		err = rows.Scan(&command, &state, &count)
		if err != nil {
			return err
		}
		foundState := findThreadState(command, state)
		stateCounts[foundState] += count
	}

	tags := map[string]string{"server": servtag}
	for s, c := range stateCounts {
		fields[newNamespace("threads", s)] = c
	}
	acc.AddFields("mysql_info_schema", fields, tags)
	return nil
}

// gatherPerfTableIOWaits can be used to get total count and time
// of I/O wait event for each table and process
func (m *Mysql) gatherPerfTableIOWaits(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(perfTableIOWaitsQuery)
	if err != nil {
		return err
	}

	defer rows.Close()
	var (
		objSchema, objName, servtag                       string
		countFetch, countInsert, countUpdate, countDelete uint64
		timeFetch, timeInsert, timeUpdate, timeDelete     uint64
	)

	servtag, err = parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}

	for rows.Next() {
		err = rows.Scan(&objSchema, &objName,
			&countFetch, &countInsert, &countUpdate, &countDelete,
			&timeFetch, &timeInsert, &timeUpdate, &timeDelete,
		)

		if err != nil {
			return err
		}

		tags := map[string]string{
			"server": servtag,
			"schema": objSchema,
			"name":   objName,
		}
		fields := make(map[string]interface{})
		fields["table_io_waits_total_fetch"] = float64(countFetch)
		fields["table_io_waits_total_insert"] = float64(countInsert)
		fields["table_io_waits_total_update"] = float64(countUpdate)
		fields["table_io_waits_total_delete"] = float64(countDelete)

		fields["table_io_waits_seconds_total_fetch"] = float64(timeFetch) / picoSeconds
		fields["table_io_waits_seconds_total_insert"] = float64(timeInsert) / picoSeconds
		fields["table_io_waits_seconds_total_update"] = float64(timeUpdate) / picoSeconds
		fields["table_io_waits_seconds_total_delete"] = float64(timeDelete) / picoSeconds

		acc.AddFields("mysql_perf_schema", fields, tags)
	}
	return nil
}

// gatherPerfIndexIOWaits can be used to get total count and time
// of I/O wait event for each index and process
func (m *Mysql) gatherPerfIndexIOWaits(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(perfIndexIOWaitsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		objSchema, objName, indexName, servtag            string
		countFetch, countInsert, countUpdate, countDelete uint64
		timeFetch, timeInsert, timeUpdate, timeDelete     uint64
	)

	servtag, err = parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}

	for rows.Next() {
		err = rows.Scan(&objSchema, &objName, &indexName,
			&countFetch, &countInsert, &countUpdate, &countDelete,
			&timeFetch, &timeInsert, &timeUpdate, &timeDelete,
		)

		if err != nil {
			return err
		}

		tags := map[string]string{
			"server": servtag,
			"schema": objSchema,
			"name":   objName,
			"index":  indexName,
		}
		fields := make(map[string]interface{})
		fields["index_io_waits_total_fetch"] = float64(countFetch)
		fields["index_io_waits_seconds_total_fetch"] = float64(timeFetch) / picoSeconds

		// update write columns only when index is NONE
		if indexName == "NONE" {
			fields["index_io_waits_total_insert"] = float64(countInsert)
			fields["index_io_waits_total_update"] = float64(countUpdate)
			fields["index_io_waits_total_delete"] = float64(countDelete)

			fields["index_io_waits_seconds_total_insert"] = float64(timeInsert) / picoSeconds
			fields["index_io_waits_seconds_total_update"] = float64(timeUpdate) / picoSeconds
			fields["index_io_waits_seconds_total_delete"] = float64(timeDelete) / picoSeconds
		}

		acc.AddFields("mysql_perf_schema", fields, tags)
	}
	return nil
}

// gatherInfoSchemaAutoIncStatuses can be used to get auto incremented value of the column
func (m *Mysql) gatherInfoSchemaAutoIncStatuses(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(infoSchemaAutoIncQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		schema, table, column string
		incValue, maxInt      uint64
	)

	servtag, err := parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}

	for rows.Next() {
		if err := rows.Scan(&schema, &table, &column, &incValue, &maxInt); err != nil {
			return err
		}
		tags := map[string]string{
			"server": servtag,
			"schema": schema,
			"table":  table,
			"column": column,
		}
		fields := make(map[string]interface{})
		fields["auto_increment_column"] = incValue
		fields["auto_increment_column_max"] = maxInt

		acc.AddFields("mysql_info_schema", fields, tags)
	}
	return nil
}

// gatherPerfTableLockWaits can be used to get
// the total number and time for SQL and external lock wait events
// for each table and operation
func (m *Mysql) gatherPerfTableLockWaits(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	// check if table exists,
	// if performance_schema is not enabled, tables do not exist
	// then there is no need to scan them
	var tableName string
	err := db.QueryRow(perfSchemaTablesQuery, "table_lock_waits_summary_by_table").Scan(&tableName)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	rows, err := db.Query(perfTableLockWaitsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	servtag, err := parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}

	var (
		objectSchema               string
		objectName                 string
		countReadNormal            uint64
		countReadWithSharedLocks   uint64
		countReadHighPriority      uint64
		countReadNoInsert          uint64
		countReadExternal          uint64
		countWriteAllowWrite       uint64
		countWriteConcurrentInsert uint64
		countWriteDelayed          uint64
		countWriteLowPriority      uint64
		countWriteNormal           uint64
		countWriteExternal         uint64
		timeReadNormal             uint64
		timeReadWithSharedLocks    uint64
		timeReadHighPriority       uint64
		timeReadNoInsert           uint64
		timeReadExternal           uint64
		timeWriteAllowWrite        uint64
		timeWriteConcurrentInsert  uint64
		timeWriteDelayed           uint64
		timeWriteLowPriority       uint64
		timeWriteNormal            uint64
		timeWriteExternal          uint64
	)

	for rows.Next() {
		err = rows.Scan(
			&objectSchema,
			&objectName,
			&countReadNormal,
			&countReadWithSharedLocks,
			&countReadHighPriority,
			&countReadNoInsert,
			&countReadExternal,
			&countWriteAllowWrite,
			&countWriteConcurrentInsert,
			&countWriteDelayed,
			&countWriteLowPriority,
			&countWriteNormal,
			&countWriteExternal,
			&timeReadNormal,
			&timeReadWithSharedLocks,
			&timeReadHighPriority,
			&timeReadNoInsert,
			&timeReadExternal,
			&timeWriteAllowWrite,
			&timeWriteConcurrentInsert,
			&timeWriteDelayed,
			&timeWriteLowPriority,
			&timeWriteNormal,
			&timeWriteExternal,
		)

		if err != nil {
			return err
		}
		tags := map[string]string{
			"server": servtag,
			"schema": objectSchema,
			"table":  objectName,
		}
		fields := make(map[string]interface{})

		tags["operation"] = "read_normal"
		fields["sql_lock_waits_total"] = float64(countReadNormal)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "read_with_shared_locks"
		fields["sql_lock_waits_total"] = float64(countReadWithSharedLocks)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "read_high_priority"
		fields["sql_lock_waits_total"] = float64(countReadHighPriority)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "read_no_insert"
		fields["sql_lock_waits_total"] = float64(countReadNoInsert)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_normal"
		fields["sql_lock_waits_total"] = float64(countWriteNormal)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_allow_write"
		fields["sql_lock_waits_total"] = float64(countWriteAllowWrite)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_concurrent_insert"
		fields["sql_lock_waits_total"] = float64(countWriteConcurrentInsert)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_delayed"
		fields["sql_lock_waits_total"] = float64(countWriteDelayed)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_low_priority"
		fields["sql_lock_waits_total"] = float64(countWriteLowPriority)
		acc.AddFields("mysql_perf_schema", fields, tags)

		delete(fields, "sql_lock_waits_total")

		tags["operation"] = "read"
		fields["external_lock_waits_total"] = float64(countReadExternal)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write"
		fields["external_lock_waits_total"] = float64(countWriteExternal)
		acc.AddFields("mysql_perf_schema", fields, tags)

		delete(fields, "external_lock_waits_total")

		tags["operation"] = "read_normal"
		fields["sql_lock_waits_seconds_total"] = float64(timeReadNormal / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "read_with_shared_locks"
		fields["sql_lock_waits_seconds_total"] = float64(timeReadWithSharedLocks / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "read_high_priority"
		fields["sql_lock_waits_seconds_total"] = float64(timeReadHighPriority / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "read_no_insert"
		fields["sql_lock_waits_seconds_total"] = float64(timeReadNoInsert / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_normal"
		fields["sql_lock_waits_seconds_total"] = float64(timeWriteNormal / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_allow_write"
		fields["sql_lock_waits_seconds_total"] = float64(timeWriteAllowWrite / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_concurrent_insert"
		fields["sql_lock_waits_seconds_total"] = float64(timeWriteConcurrentInsert / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_delayed"
		fields["sql_lock_waits_seconds_total"] = float64(timeWriteDelayed / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write_low_priority"
		fields["sql_lock_waits_seconds_total"] = float64(timeWriteLowPriority / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		delete(fields, "sql_lock_waits_seconds_total")

		tags["operation"] = "read"
		fields["external_lock_waits_seconds_total"] = float64(timeReadExternal / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["operation"] = "write"
		fields["external_lock_waits_seconds_total"] = float64(timeWriteExternal / picoSeconds)
		acc.AddFields("mysql_perf_schema", fields, tags)
	}
	return nil
}

// gatherPerfEventWaits can be used to get total time and number of event waits
func (m *Mysql) gatherPerfEventWaits(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(perfEventWaitsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		event               string
		starCount, timeWait uint64
	)

	servtag, err := parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}
	tags := map[string]string{
		"server": servtag,
	}
	for rows.Next() {
		if err := rows.Scan(&event, &starCount, &timeWait); err != nil {
			return err
		}
		tags["event_name"] = event
		fields := make(map[string]interface{})
		fields["events_waits_total"] = float64(starCount)
		fields["events_waits_seconds_total"] = float64(timeWait) / picoSeconds

		acc.AddFields("mysql_perf_schema", fields, tags)
	}
	return nil
}

// gatherPerfFileEvents can be used to get stats on file events
func (m *Mysql) gatherPerfFileEventsStatuses(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	rows, err := db.Query(perfFileEventsQuery)
	if err != nil {
		return err
	}

	defer rows.Close()

	var (
		eventName                                 string
		countRead, countWrite, countMisc          uint64
		sumTimerRead, sumTimerWrite, sumTimerMisc uint64
		sumNumBytesRead, sumNumBytesWrite         uint64
	)

	servtag, err := parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}
	tags := map[string]string{
		"server": servtag,
	}
	for rows.Next() {
		err = rows.Scan(
			&eventName,
			&countRead, &sumTimerRead, &sumNumBytesRead,
			&countWrite, &sumTimerWrite, &sumNumBytesWrite,
			&countMisc, &sumTimerMisc,
		)
		if err != nil {
			return err
		}

		tags["event_name"] = eventName
		fields := make(map[string]interface{})

		tags["mode"] = "misc"
		fields["file_events_total"] = float64(countWrite)
		fields["file_events_seconds_total"] = float64(sumTimerMisc) / picoSeconds
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["mode"] = "read"
		fields["file_events_total"] = float64(countRead)
		fields["file_events_seconds_total"] = float64(sumTimerRead) / picoSeconds
		fields["file_events_bytes_totals"] = float64(sumNumBytesRead)
		acc.AddFields("mysql_perf_schema", fields, tags)

		tags["mode"] = "write"
		fields["file_events_total"] = float64(countWrite)
		fields["file_events_seconds_total"] = float64(sumTimerWrite) / picoSeconds
		fields["file_events_bytes_totals"] = float64(sumNumBytesWrite)
		acc.AddFields("mysql_perf_schema", fields, tags)

	}
	return nil
}

// gatherPerfEventsStatements can be used to get attributes of each event
func (m *Mysql) gatherPerfEventsStatements(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	query := fmt.Sprintf(
		perfEventsStatementsQuery,
		m.PerfEventsStatementsDigestTextLimit,
		m.PerfEventsStatementsTimeLimit,
		m.PerfEventsStatementsLimit,
	)

	rows, err := db.Query(query)
	if err != nil {
		return err
	}

	defer rows.Close()

	var (
		schemaName, digest, digest_text      string
		count, queryTime, errors, warnings   uint64
		rowsAffected, rowsSent, rowsExamined uint64
		tmpTables, tmpDiskTables             uint64
		sortMergePasses, sortRows            uint64
		noIndexUsed                          uint64
	)

	servtag, err := parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}
	tags := map[string]string{
		"server": servtag,
	}

	for rows.Next() {
		err = rows.Scan(
			&schemaName, &digest, &digest_text,
			&count, &queryTime, &errors, &warnings,
			&rowsAffected, &rowsSent, &rowsExamined,
			&tmpTables, &tmpDiskTables,
			&sortMergePasses, &sortRows,
		)

		if err != nil {
			return err
		}
		tags["schema"] = schemaName
		tags["digest"] = digest
		tags["digest_text"] = digest_text

		fields := make(map[string]interface{})

		fields["events_statements_total"] = float64(count)
		fields["events_statements_seconds_total"] = float64(queryTime) / picoSeconds
		fields["events_statements_errors_total"] = float64(errors)
		fields["events_statements_warnings_total"] = float64(warnings)
		fields["events_statements_rows_affected_total"] = float64(rowsAffected)
		fields["events_statements_rows_sent_total"] = float64(rowsSent)
		fields["events_statements_rows_examined_total"] = float64(rowsExamined)
		fields["events_statements_tmp_tables_total"] = float64(tmpTables)
		fields["events_statements_tmp_disk_tables_total"] = float64(tmpDiskTables)
		fields["events_statements_sort_merge_passes_total"] = float64(sortMergePasses)
		fields["events_statements_sort_rows_total"] = float64(sortRows)
		fields["events_statements_no_index_used_total"] = float64(noIndexUsed)

		acc.AddFields("mysql_perf_schema", fields, tags)
	}
	return nil
}

func (m *Mysql) gatherTableSchema(db *sql.DB, serv string, acc telegraf.Accumulator) error {
	var (
		dbList  []string
		servtag string
	)
	servtag, err := parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}

	// if the list of databases if empty, then get all databases
	if len(m.TableSchemaDatabases) == 0 {
		rows, err := db.Query(dbListQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		var database string
		for rows.Next() {
			err = rows.Scan(&database)
			if err != nil {
				return err
			}

			dbList = append(dbList, database)
		}
	} else {
		dbList = m.TableSchemaDatabases
	}

	for _, database := range dbList {
		rows, err := db.Query(fmt.Sprintf(tableSchemaQuery, database))
		if err != nil {
			return err
		}
		defer rows.Close()
		var (
			tableSchema   string
			tableName     string
			tableType     string
			engine        string
			version       uint64
			rowFormat     string
			tableRows     uint64
			dataLength    uint64
			indexLength   uint64
			dataFree      uint64
			createOptions string
		)
		for rows.Next() {
			err = rows.Scan(
				&tableSchema,
				&tableName,
				&tableType,
				&engine,
				&version,
				&rowFormat,
				&tableRows,
				&dataLength,
				&indexLength,
				&dataFree,
				&createOptions,
			)
			if err != nil {
				return err
			}
			tags := map[string]string{"server": servtag}
			tags["schema"] = tableSchema
			tags["table"] = tableName
			versionTags := tags

			acc.Add(newNamespace("info_schema", "table_rows"), float64(tableRows), tags)

			tags["component"] = "data_length"
			acc.Add(newNamespace("info_schema", "table_size", "data_length"), float64(dataLength), tags)

			tags["component"] = "index_length"
			acc.Add(newNamespace("info_schema", "table_size", "index_length"), float64(indexLength), tags)

			tags["component"] = "data_free"
			acc.Add(newNamespace("info_schema", "table_size", "data_free"), float64(dataFree), tags)

			versionTags["type"] = tableType
			versionTags["engine"] = engine
			versionTags["row_format"] = rowFormat
			versionTags["create_options"] = createOptions

			acc.Add(newNamespace("info_schema", "table_version"), float64(version), versionTags)
		}
	}
	return nil
}

// parseValue can be used to convert values such as "ON","OFF","Yes","No" to 0,1
func parseValue(value sql.RawBytes) (float64, bool) {
	if bytes.Compare(value, []byte("Yes")) == 0 || bytes.Compare(value, []byte("ON")) == 0 {
		return 1, true
	}

	if bytes.Compare(value, []byte("No")) == 0 || bytes.Compare(value, []byte("OFF")) == 0 {
		return 0, false
	}
	n, err := strconv.ParseFloat(string(value), 64)
	return n, err == nil
}

// findThreadState can be used to find thread state by command and plain state
func findThreadState(rawCommand, rawState string) string {
	var (
		// replace '_' symbol with space
		command = strings.Replace(strings.ToLower(rawCommand), "_", " ", -1)
		state   = strings.Replace(strings.ToLower(rawState), "_", " ", -1)
	)
	// if the state is already valid, then return it
	if _, ok := generalThreadStates[state]; ok {
		return state
	}

	// if state is plain, return the mapping
	if mappedState, ok := stateStatusMappings[state]; ok {
		return mappedState
	}
	// if the state is any lock, return the special state
	if strings.Contains(state, "waiting for") && strings.Contains(state, "lock") {
		return "waiting for lock"
	}

	if command == "sleep" && state == "" {
		return "idle"
	}

	if command == "query" {
		return "executing"
	}

	if command == "binlog dump" {
		return "replication master"
	}
	// if no mappings found and state is invalid, then return "other" state
	return "other"
}

// newNamespace can be used to make a namespace
func newNamespace(words ...string) string {
	return strings.Replace(strings.Join(words, "_"), " ", "_", -1)
}

func init() {
	inputs.Add("mysql", func() telegraf.Input {
		return &Mysql{}
	})
}
