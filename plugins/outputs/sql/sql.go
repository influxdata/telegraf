package sql

import (
	gosql "database/sql"
	"fmt"
	"strings"

	//Register sql drivers
	_ "github.com/ClickHouse/clickhouse-go" // clickhouse
	_ "github.com/denisenkom/go-mssqldb"    // mssql (sql server)
	_ "github.com/go-sql-driver/mysql"      // mysql
	_ "github.com/jackc/pgx/v4/stdlib"      // pgx (postgres)
	_ "github.com/snowflakedb/gosnowflake"  // snowflake

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type ConvertStruct struct {
	Integer         string
	Real            string
	Text            string
	Timestamp       string
	Defaultvalue    string
	Unsigned        string
	Bool            string
	ConversionStyle string
}

type SQL struct {
	Driver              string
	DataSourceName      string
	TimestampColumn     string
	TableTemplate       string
	TableExistsTemplate string
	InitSQL             string `toml:"init_sql"`
	Convert             ConvertStruct

	db     *gosql.DB
	Log    telegraf.Logger `toml:"-"`
	tables map[string]bool
}

func (p *SQL) Connect() error {
	db, err := gosql.Open(p.Driver, p.DataSourceName)
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
	p.tables = make(map[string]bool)

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
		if p.Convert.ConversionStyle == "unsigned_suffix" {
			datatype = fmt.Sprintf("%s %s", p.Convert.Integer, p.Convert.Unsigned)
		} else if p.Convert.ConversionStyle == "literal" {
			datatype = p.Convert.Unsigned
		} else {
			p.Log.Errorf("unknown converstaion style: %s", p.Convert.ConversionStyle)
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
	var columns []string
	//  ##  {KEY_COLUMNS} is a comma-separated list of key columns (timestamp and tags)
	//var pk []string

	if p.TimestampColumn != "" {
		//pk = append(pk, quoteIdent(p.TimestampColumn))
		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(p.TimestampColumn), p.Convert.Timestamp))
	}

	for _, tag := range metric.TagList() {
		//pk = append(pk, quoteIdent(tag.Key))
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
	//query = strings.Replace(query, "{KEY_COLUMNS}", strings.Join(pk, ","), -1)

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
			p.tables[tablename] = true
		}

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
				return fmt.Errorf("begin failed: %v", err)
			}
			stmt, err := tx.Prepare(sql)
			if err != nil {
				return fmt.Errorf("prepare failed: %v", err)
			}
			defer stmt.Close() //nolint:revive // We cannot do anything about a failing close.

			_, err = stmt.Exec(values...)
			if err != nil {
				return fmt.Errorf("execution failed: %v", err)
			}
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit failed: %v", err)
			}
		default:
			_, err = p.db.Exec(sql, values...)
			if err != nil {
				return fmt.Errorf("execution failed: %v", err)
			}
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
		TimestampColumn:     "timestamp",
		Convert: ConvertStruct{
			Integer:         "INT",
			Real:            "DOUBLE",
			Text:            "TEXT",
			Timestamp:       "TIMESTAMP",
			Defaultvalue:    "TEXT",
			Unsigned:        "UNSIGNED",
			Bool:            "BOOL",
			ConversionStyle: "unsigned_suffix",
		},
	}
}
