package sql

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx"
	// These SQL drivers can be enabled if
	// they are added to depencies
	// _ "github.com/lib/pq"
	// _ "github.com/mattn/go-sqlite3"
	// _ "github.com/zensqlmonitor/go-mssqldb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type ConvertStruct struct {
	Integer      string
	Real         string
	Text         string
	Timestamp    string
	Defaultvalue string
	Unsigned     string
}

type Sql struct {
	db                  *sql.DB
	Driver              string
	Address             string
	TableTemplate       string
	TableExistsTemplate string
	TagTableSuffix      string
	Tables              map[string]bool
	Convert             []ConvertStruct
}

func (p *Sql) Connect() error {
	db, err := sql.Open(p.Driver, p.Address)
	if err != nil {
		return err
	}
	p.db = db
	p.Tables = make(map[string]bool)

	return nil
}

func (p *Sql) Close() error {
	return p.db.Close()
}

func contains(haystack []string, needle string) bool {
	for _, key := range haystack {
		if key == needle {
			return true
		}
	}
	return false
}

func quoteIdent(name string) string {
	return name
}

func quoteLiteral(name string) string {
	return "'" + strings.Replace(name, "'", "''", -1) + "'"
}

func (p *Sql) deriveDatatype(value interface{}) string {
	var datatype string

	switch value.(type) {
	case int64:
		datatype = p.Convert[0].Integer
	case uint64:
		datatype = fmt.Sprintf("%s %s", p.Convert[0].Integer, p.Convert[0].Unsigned)
	case float64:
		datatype = p.Convert[0].Real
	case string:
		datatype = p.Convert[0].Text
	default:
		datatype = p.Convert[0].Defaultvalue
		log.Printf("E! Unknown datatype: '%T' %v", value, value)
	}
	return datatype
}

var sampleConfig = `
# Send metrics to SQL-Database (Example configuration for MySQL/MariaDB)
[[outputs.sql]]
  ## Database Driver, required.
  ## Valid options: mssql (SQLServer), mysql (MySQL), postgres (Postgres), sqlite3 (SQLite), [oci8 ora.v4 (Oracle)]
  driver = "mysql"

  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
  ##
  ## All connection parameters are optional.
  ##
  ## Without the dbname parameter, the driver will default to a database
  ## with the same name as the user. This dbname is just for instantiating a
  ## connection with the server and doesn't restrict the databases we are trying
  ## to grab metrics for.
  ##
  address = "username:password@tcp(server:port)/table"

  ## Available Variables:
  ##   {TABLE} - tablename as identifier
  ##   {TABLELITERAL} - tablename as string literal
  ##   {COLUMNS} - column definitions
  ##   {KEY_COLUMNS} - comma-separated list of key columns (time + tags)
  ##

  ## Check with this is table exists
  ##
  ## Template for MySQL is "SELECT 1 FROM {TABLE} LIMIT 1"
  ##
  table_exists_template = "SELECT 1 FROM {TABLE} LIMIT 1"

  ## Template to use for generating tables

  ## Default template
  ##
  # table_template = "CREATE TABLE {TABLE}({COLUMNS})"

  ## Convert Telegraf datatypes to these types
  [[outputs.sql.convert]]
    integer              = "INT"
    real                 = "DOUBLE"
    text                 = "TEXT"
    timestamp            = "TIMESTAMP"
    defaultvalue         = "TEXT"
    unsigned             = "UNSIGNED"
`

func (p *Sql) SampleConfig() string { return sampleConfig }
func (p *Sql) Description() string  { return "Send metrics to SQL Database" }

func (p *Sql) generateCreateTable(metric telegraf.Metric) string {
	var columns []string
	var pk []string
	var sql []string

	pk = append(pk, quoteIdent("timestamp"))
	columns = append(columns, fmt.Sprintf("timestamp %s", p.Convert[0].Timestamp))

	// handle tags if necessary
	if len(metric.Tags()) > 0 {
		// tags in measurement table
		for column := range metric.Tags() {
			pk = append(pk, quoteIdent(column))
			columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(column), p.Convert[0].Text))
		}
	}

	var datatype string
	for column, v := range metric.Fields() {
		datatype = p.deriveDatatype(v)
		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(column), datatype))
	}

	query := strings.Replace(p.TableTemplate, "{TABLE}", quoteIdent(metric.Name()), -1)
	query = strings.Replace(query, "{TABLELITERAL}", quoteLiteral(metric.Name()), -1)
	query = strings.Replace(query, "{COLUMNS}", strings.Join(columns, ","), -1)
	query = strings.Replace(query, "{KEY_COLUMNS}", strings.Join(pk, ","), -1)

	sql = append(sql, query)
	return strings.Join(sql, ";")
}

func (p *Sql) generateInsert(tablename string, columns []string) string {

	var placeholder, quoted []string
	for _, column := range columns {
		placeholder = append(placeholder, fmt.Sprintf("?"))
		quoted = append(quoted, quoteIdent(column))
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", quoteIdent(tablename), strings.Join(quoted, ","), strings.Join(placeholder, ","))
	return sql
}

func (p *Sql) tableExists(tableName string) bool {
	stmt := strings.Replace(p.TableExistsTemplate, "{TABLE}", quoteIdent(tableName), -1)

	_, err := p.db.Exec(stmt)
	if err != nil {
		return false
	}
	return true
}

func (p *Sql) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		tablename := metric.Name()

		// create table if needed
		if p.Tables[tablename] == false && p.tableExists(tablename) == false {
			createStmt := p.generateCreateTable(metric)
			_, err := p.db.Exec(createStmt)
			if err != nil {
				return err
			}
			p.Tables[tablename] = true
		}

		var columns []string
		var values []interface{}

		// We assume that SQL is making auto timestamp
		//columns = append(columns, "timestamp")
		//values = append(values, metric.Time())

		if len(metric.Tags()) > 0 {
			// tags in measurement table
			for column, value := range metric.Tags() {
				columns = append(columns, column)
				values = append(values, value)
			}
		}

		for column, value := range metric.Fields() {
			columns = append(columns, column)
			values = append(values, value)
		}

		sql := p.generateInsert(tablename, columns)
		_, err := p.db.Exec(sql, values...)

		if err != nil {
			// check if insert error was caused by column mismatch
			log.Printf("E! Error during insert: %v", err)
			return err
		}
	}
	return nil
}

func init() {
	outputs.Add("sql", func() telegraf.Output { return newSql() })
}

func newSql() *Sql {
	return &Sql{
		TableTemplate:       "CREATE TABLE {TABLE}({COLUMNS})",
		TableExistsTemplate: "SELECT 1 FROM {TABLE} LIMIT 1",
		TagTableSuffix:      "_tag",
	}
}
