package sql

import (
	gosql "database/sql"
	"fmt"
	"strings"

	//Register sql drivers
	_ "github.com/denisenkom/go-mssqldb"   // mssql (sql server)
	_ "github.com/go-sql-driver/mysql"     // mysql
	_ "github.com/jackc/pgx/stdlib"        // pgx (postgres)
	_ "github.com/snowflakedb/gosnowflake" // snowflake

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

type SQL struct {
	db                  *gosql.DB
	Driver              string
	Address             string
	TableTemplate       string
	TableExistsTemplate string
	InitSQL             string `toml:"init_sql"`
	Convert             ConvertStruct

	Log    telegraf.Logger `toml:"-"`
	Tables map[string]bool `toml:"-"`
}

func (p *SQL) Connect() error {
	db, err := gosql.Open(p.Driver, p.Address)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	if p.InitSQL != "" {
		_, err = db.Exec(p.InitSQL)
		if err != nil {
			return err
		}
	}

	p.db = db
	p.Tables = make(map[string]bool)

	return nil
}

func (p *SQL) Close() error {
	return p.db.Close()
}

// Quote an identifier (table or column name)
func quoteIdent(name string) string {
	return `"` + strings.Replace(sanitizeQuoted(name), `"`, `""`, -1) + `"`
}

// Quote a string literal
func quoteStr(name string) string {
	return "'" + strings.Replace(name, "'", "''", -1) + "'"
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
		datatype = fmt.Sprintf("%s %s", p.Convert.Integer, p.Convert.Unsigned)
	case float64:
		datatype = p.Convert.Real
	case string:
		datatype = p.Convert.Text
	default:
		datatype = p.Convert.Defaultvalue
		p.Log.Errorf("Unknown datatype: '%T' %v", value, value)
	}
	return datatype
}

var sampleConfig = `
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

  ## also NO_BACKSLASH_ESCAPES
  # init_sql = "SET sql_mode='ANSI_QUOTES';"

  ## Convert Telegraf datatypes to these types
  #[outputs.sql.convert]
  #  integer              = "INT"
  #  real                 = "DOUBLE"
  #  text                 = "TEXT"
  #  timestamp            = "TIMESTAMP"
  #  defaultvalue         = "TEXT"
  #  unsigned             = "UNSIGNED"
`

func (p *SQL) SampleConfig() string { return sampleConfig }
func (p *SQL) Description() string  { return "Send metrics to SQL Database" }

func (p *SQL) generateCreateTable(metric telegraf.Metric) string {
	var columns []string
	var pk []string

	pk = append(pk, quoteIdent("timestamp"))
	columns = append(columns, fmt.Sprintf("%s %s", quoteIdent("timestamp"), p.Convert.Timestamp))

	for _, tag := range metric.TagList() {
		pk = append(pk, quoteIdent(tag.Key))
		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(tag.Key), p.Convert.Text))
	}

	var datatype string
	for _, field := range metric.FieldList() {
		datatype = p.deriveDatatype(field.Value)
		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(field.Key), datatype))
	}

	query := p.TableTemplate
	query = strings.Replace(query, "{TABLE}", quoteIdent(metric.Name()), -1)
	query = strings.Replace(query, "{TABLELITERAL}", quoteStr(metric.Name()), -1)
	query = strings.Replace(query, "{COLUMNS}", strings.Join(columns, ","), -1)
	query = strings.Replace(query, "{KEY_COLUMNS}", strings.Join(pk, ","), -1)

	return query
}

func (p *SQL) generateInsert(tablename string, columns []string) string {
	var placeholders, quotedColumns []string
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

	return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)",
		quoteIdent(tablename),
		strings.Join(quotedColumns, ","),
		strings.Join(placeholders, ","))
}

func (p *SQL) tableExists(tableName string) bool {
	stmt := strings.Replace(p.TableExistsTemplate, "{TABLE}", quoteIdent(tableName), -1)

	_, err := p.db.Exec(stmt)
	return err == nil
}

func (p *SQL) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		tablename := metric.Name()

		// create table if needed
		if !p.Tables[tablename] && !p.tableExists(tablename) {
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
		columns = append(columns, "timestamp")
		values = append(values, metric.Time())

		for column, value := range metric.Tags() {
			columns = append(columns, column)
			values = append(values, value)
		}

		for column, value := range metric.Fields() {
			columns = append(columns, column)
			values = append(values, value)
		}

		sql := p.generateInsert(tablename, columns)
		_, err := p.db.Exec(sql, values...)

		if err != nil {
			// check if insert error was caused by column mismatch
			p.Log.Errorf("Error during insert: %v, %v", err, sql)
			return err
		}
	}
	return nil
}

func init() {
	outputs.Add("sql", func() telegraf.Output { return newSQL() })
}

func newSQL() *SQL {
	return &SQL{
		TableTemplate:       "CREATE TABLE {TABLE}({COLUMNS})",
		TableExistsTemplate: "SELECT 1 FROM {TABLE} LIMIT 1",
		Convert: ConvertStruct{
			Integer:      "INT",
			Real:         "DOUBLE",
			Text:         "TEXT",
			Timestamp:    "TIMESTAMP",
			Defaultvalue: "TEXT",
			Unsigned:     "UNSIGNED",
		},
	}
}
