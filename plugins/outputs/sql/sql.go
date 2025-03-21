//go:generate ../../../tools/readme_config_includer/generator
package sql

import (
	gosql "database/sql"
	_ "embed"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"              // clickhouse
	_ "github.com/go-sql-driver/mysql"                      // mysql
	_ "github.com/jackc/pgx/v4/stdlib"                      // pgx (postgres)
	_ "github.com/microsoft/go-mssqldb"                     // mssql (sql server)
	_ "github.com/microsoft/go-mssqldb/integratedauth/krb5" // integrated auth for mssql
	_ "github.com/snowflakedb/gosnowflake"                  // snowflake

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

var defaultConvert = ConvertStruct{
	Integer:         "INT",
	Real:            "DOUBLE",
	Text:            "TEXT",
	Timestamp:       "TIMESTAMP",
	Defaultvalue:    "TEXT",
	Unsigned:        "UNSIGNED",
	Bool:            "BOOL",
	ConversionStyle: "unsigned_suffix",
}

type ConvertStruct struct {
	Integer         string `toml:"integer"`
	Real            string `toml:"real"`
	Text            string `toml:"text"`
	Timestamp       string `toml:"timestamp"`
	Defaultvalue    string `toml:"defaultvalue"`
	Unsigned        string `toml:"unsigned"`
	Bool            string `toml:"bool"`
	ConversionStyle string `toml:"conversion_style"`
}

type SQL struct {
	Driver                string          `toml:"driver"`
	DataSourceName        string          `toml:"data_source_name"`
	TimestampColumn       string          `toml:"timestamp_column"`
	TableTemplate         string          `toml:"table_template"`
	TableExistsTemplate   string          `toml:"table_exists_template"`
	InitSQL               string          `toml:"init_sql"`
	Convert               ConvertStruct   `toml:"convert"`
	ConnectionMaxIdleTime config.Duration `toml:"connection_max_idle_time"`
	ConnectionMaxLifetime config.Duration `toml:"connection_max_lifetime"`
	ConnectionMaxIdle     int             `toml:"connection_max_idle"`
	ConnectionMaxOpen     int             `toml:"connection_max_open"`
	Log                   telegraf.Logger `toml:"-"`

	db     *gosql.DB
	tables map[string]bool
}

func (*SQL) SampleConfig() string {
	return sampleConfig
}

func (p *SQL) Init() error {
	// Set defaults
	if p.TableExistsTemplate == "" {
		p.TableExistsTemplate = "SELECT 1 FROM {TABLE} LIMIT 1"
	}

	if p.TableTemplate == "" {
		if p.Driver == "clickhouse" {
			p.TableTemplate = "CREATE TABLE {TABLE}({COLUMNS}) ORDER BY ({TAG_COLUMN_NAMES}, {TIMESTAMP_COLUMN_NAME})"
		} else {
			p.TableTemplate = "CREATE TABLE {TABLE}({COLUMNS})"
		}
	}

	// Check for a valid driver
	switch p.Driver {
	case "clickhouse":
		// Convert v1-style Clickhouse DSN to v2-style
		p.convertClickHouseDsn()
	case "mssql", "mysql", "pgx", "snowflake", "sqlite":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("unknown driver %q", p.Driver)
	}

	return nil
}

func (p *SQL) Connect() error {
	db, err := gosql.Open(p.Driver, p.DataSourceName)
	if err != nil {
		return fmt.Errorf("creating database client failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("pinging database failed: %w", err)
	}

	db.SetConnMaxIdleTime(time.Duration(p.ConnectionMaxIdleTime))
	db.SetConnMaxLifetime(time.Duration(p.ConnectionMaxLifetime))
	db.SetMaxIdleConns(p.ConnectionMaxIdle)
	db.SetMaxOpenConns(p.ConnectionMaxOpen)

	if p.InitSQL != "" {
		if _, err = db.Exec(p.InitSQL); err != nil {
			return fmt.Errorf("initializing database failed: %w", err)
		}
	}

	p.db = db
	p.tables = make(map[string]bool)

	return nil
}

func (p *SQL) Close() error {
	return p.db.Close()
}

// Quote an identifier (table or column name)
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(sanitizeQuoted(name), `"`, `""`) + `"`
}

// Quote a string literal
func quoteStr(name string) string {
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
}

func sanitizeQuoted(in string) string {
	// https://dev.mysql.com/doc/refman/8.0/en/identifiers.html
	// https://www.postgresql.org/docs/13/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS

	// Whitelist allowed characters
	return strings.Map(func(r rune) rune {
		switch {
		case r >= '\u0001' && r <= '\uFFFF':
			return r
		default:
			return '_'
		}
	}, in)
}

func (p *SQL) deriveDatatype(value interface{}) string {
	var datatype string

	switch value.(type) {
	case int64:
		datatype = p.Convert.Integer
	case uint64:
		if p.Convert.ConversionStyle == "unsigned_suffix" {
			datatype = fmt.Sprintf("%s %s", p.Convert.Integer, p.Convert.Unsigned)
		} else if p.Convert.ConversionStyle == "literal" {
			datatype = p.Convert.Unsigned
		} else {
			p.Log.Errorf("unknown conversion style: %s", p.Convert.ConversionStyle)
		}
	case float64:
		datatype = p.Convert.Real
	case string:
		datatype = p.Convert.Text
	case bool:
		datatype = p.Convert.Bool
	default:
		datatype = p.Convert.Defaultvalue
		p.Log.Errorf("Unknown datatype: '%T' %v", value, value)
	}
	return datatype
}

func (p *SQL) generateCreateTable(metric telegraf.Metric) string {
	columns := make([]string, 0, len(metric.TagList())+len(metric.FieldList())+1)
	tagColumnNames := make([]string, 0, len(metric.TagList()))

	if p.TimestampColumn != "" {
		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(p.TimestampColumn), p.Convert.Timestamp))
	}

	for _, tag := range metric.TagList() {
		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(tag.Key), p.Convert.Text))
		tagColumnNames = append(tagColumnNames, quoteIdent(tag.Key))
	}

	var datatype string
	for _, field := range metric.FieldList() {
		datatype = p.deriveDatatype(field.Value)
		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(field.Key), datatype))
	}

	query := p.TableTemplate
	query = strings.ReplaceAll(query, "{TABLE}", quoteIdent(metric.Name()))
	query = strings.ReplaceAll(query, "{TABLELITERAL}", quoteStr(metric.Name()))
	query = strings.ReplaceAll(query, "{COLUMNS}", strings.Join(columns, ","))
	query = strings.ReplaceAll(query, "{TAG_COLUMN_NAMES}", strings.Join(tagColumnNames, ","))
	query = strings.ReplaceAll(query, "{TIMESTAMP_COLUMN_NAME}", quoteIdent(p.TimestampColumn))

	return query
}

func (p *SQL) generateInsert(tablename string, columns []string) string {
	placeholders := make([]string, 0, len(columns))
	quotedColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		quotedColumns = append(quotedColumns, quoteIdent(column))
	}
	if p.Driver == "pgx" {
		// Postgres uses $1 $2 $3 as placeholders
		for i := 0; i < len(columns); i++ {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		}
	} else {
		// Everything else uses ? ? ? as placeholders
		for i := 0; i < len(columns); i++ {
			placeholders = append(placeholders, "?")
		}
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		quoteIdent(tablename),
		strings.Join(quotedColumns, ","),
		strings.Join(placeholders, ","))
}

func (p *SQL) tableExists(tableName string) bool {
	stmt := strings.ReplaceAll(p.TableExistsTemplate, "{TABLE}", quoteIdent(tableName))

	_, err := p.db.Exec(stmt)
	return err == nil
}

func (p *SQL) Write(metrics []telegraf.Metric) error {
	var err error

	for _, metric := range metrics {
		tablename := metric.Name()

		// create table if needed
		if !p.tables[tablename] && !p.tableExists(tablename) {
			createStmt := p.generateCreateTable(metric)
			_, err := p.db.Exec(createStmt)
			if err != nil {
				return err
			}
		}
		p.tables[tablename] = true

		var columns []string
		var values []interface{}

		if p.TimestampColumn != "" {
			columns = append(columns, p.TimestampColumn)
			values = append(values, metric.Time())
		}

		for column, value := range metric.Tags() {
			columns = append(columns, column)
			values = append(values, value)
		}

		for column, value := range metric.Fields() {
			columns = append(columns, column)
			values = append(values, value)
		}

		sql := p.generateInsert(tablename, columns)

		switch p.Driver {
		case "clickhouse":
			// ClickHouse needs to batch inserts with prepared statements
			tx, err := p.db.Begin()
			if err != nil {
				return fmt.Errorf("begin failed: %w", err)
			}
			stmt, err := tx.Prepare(sql)
			if err != nil {
				return fmt.Errorf("prepare failed: %w", err)
			}
			defer stmt.Close() //nolint:revive,gocritic // done on purpose, closing will be executed properly

			_, err = stmt.Exec(values...)
			if err != nil {
				return fmt.Errorf("execution failed: %w", err)
			}
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit failed: %w", err)
			}
		default:
			_, err = p.db.Exec(sql, values...)
			if err != nil {
				return fmt.Errorf("execution failed: %w", err)
			}
		}
	}
	return nil
}

// Convert a DSN possibly using v1 parameters to clickhouse-go v2 format
func (p *SQL) convertClickHouseDsn() {
	u, err := url.Parse(p.DataSourceName)
	if err != nil {
		return
	}

	query := u.Query()

	// Log warnings for parameters no longer supported in clickhouse-go v2
	unsupported := []string{"tls_config", "no_delay", "write_timeout", "block_size", "check_connection_liveness"}
	for _, paramName := range unsupported {
		if query.Has(paramName) {
			p.Log.Warnf("DSN parameter '%s' is no longer supported by clickhouse-go v2", paramName)
			query.Del(paramName)
		}
	}
	if query.Get("connection_open_strategy") == "time_random" {
		p.Log.Warn("DSN parameter 'connection_open_strategy' can no longer be 'time_random'")
	}

	// Convert the read_timeout parameter to a duration string
	if d := query.Get("read_timeout"); d != "" {
		if _, err := strconv.ParseFloat(d, 64); err == nil {
			p.Log.Warn("Legacy DSN parameter 'read_timeout' interpreted as seconds")
			query.Set("read_timeout", d+"s")
		}
	}

	// Move database to the path
	if d := query.Get("database"); d != "" {
		p.Log.Warn("Legacy DSN parameter 'database' converted to new format")
		query.Del("database")
		u.Path = d
	}

	// Move alt_hosts to the host part
	if altHosts := query.Get("alt_hosts"); altHosts != "" {
		p.Log.Warn("Legacy DSN parameter 'alt_hosts' converted to new format")
		query.Del("alt_hosts")
		u.Host = u.Host + "," + altHosts
	}

	u.RawQuery = query.Encode()
	p.DataSourceName = u.String()
}

func init() {
	outputs.Add("sql", func() telegraf.Output {
		return &SQL{
			Convert: defaultConvert,

			// Allow overriding the timestamp column to empty by the user
			TimestampColumn: "timestamp",

			// Defaults for the connection settings (ConnectionMaxIdleTime,
			// ConnectionMaxLifetime, ConnectionMaxIdle, and ConnectionMaxOpen)
			// mirror the golang defaults. As of go 1.18 all of them default to 0
			// except max idle connections which is 2. See
			// https://pkg.go.dev/database/sql#DB.SetMaxIdleConns
			ConnectionMaxIdle: 2,
		}
	})
}
