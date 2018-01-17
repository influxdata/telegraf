package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Postgresql struct {
	db                *sql.DB
	Address           string
	TagsAsForeignkeys bool
	TagsAsJsonb       bool
	FieldsAsJsonb     bool
	TableTemplate     string
	Tables            map[string]bool
}

func (p *Postgresql) Connect() error {
	db, err := sql.Open("pgx", p.Address)
	if err != nil {
		return err
	}
	p.db = db
	p.Tables = make(map[string]bool)

	return nil
}

func (p *Postgresql) Close() error {
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
	return pgx.Identifier{name}.Sanitize()
}

func quoteLiteral(name string) string {
	return "'" + strings.Replace(name, "'", "''", -1) + "'"
}

var sampleConfig = `
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
  address = "host=localhost user=postgres sslmode=verify-full"

  ## Store tags as foreign keys in the metrics table. Default is false.
  # tags_as_foreignkeys = false

  ## Template to use for generating tables
  ## Available Variables: 
  ##   {TABLE} - tablename as identifier
  ##   {TABLELITERAL} - tablename as string literal
  ##   {COLUMNS} - column definitions
  ##   {KEY_COLUMNS} - comma-separated list of key columns (time + tags)

  ## Default template
  # table_template = "CREATE TABLE {TABLE}({COLUMNS})"
  ## Example for timescale
  # table_template = "CREATE TABLE {TABLE}({COLUMNS}); SELECT create_hypertable({TABLELITERAL},'time',chunk_time_interval := '1 week'::interval);"

  ## Use jsonb datatype for tags
  # tags_as_jsonb = true

  ## Use jsonb datatype for fields
  # fields_as_jsonb = true

`

func (p *Postgresql) SampleConfig() string { return sampleConfig }
func (p *Postgresql) Description() string  { return "Send metrics to PostgreSQL" }

func (p *Postgresql) generateCreateTable(metric telegraf.Metric) string {
	var columns []string
	var pk []string
	var sql []string

	pk = append(pk, quoteIdent("time"))
	columns = append(columns, "time timestamp")

	if p.TagsAsJsonb {
		if len(metric.Tags()) > 0 {
			columns = append(columns, "tags jsonb")
		}
	} else {
		for column, _ := range metric.Tags() {
			if p.TagsAsForeignkeys {
				key := quoteIdent(column + "_id")
				table := quoteIdent(metric.Name() + "_" + column)

				pk = append(pk, key)
				columns = append(columns, fmt.Sprintf("%s int8", key))
				sql = append(sql, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s serial primary key,%s text unique)", table, key, quoteIdent(column)))
			} else {
				pk = append(pk, quoteIdent(column))
				columns = append(columns, fmt.Sprintf("%s text", quoteIdent(column)))
			}
		}
	}

	if p.FieldsAsJsonb {
		columns = append(columns, "fields jsonb")
	} else {
		var datatype string
		for column, v := range metric.Fields() {
			switch v.(type) {
			case int64:
				datatype = "int8"
			case float64:
				datatype = "float8"
			}
			columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(column), datatype))
		}
	}

	query := strings.Replace(p.TableTemplate, "{TABLE}", quoteIdent(metric.Name()), -1)
	query = strings.Replace(query, "{TABLELITERAL}", quoteLiteral(metric.Name()), -1)
	query = strings.Replace(query, "{COLUMNS}", strings.Join(columns, ","), -1)
	query = strings.Replace(query, "{KEY_COLUMNS}", strings.Join(pk, ","), -1)

	sql = append(sql, query)
	return strings.Join(sql, ";")
}

func (p *Postgresql) generateInsert(tablename string, columns []string) string {

	var placeholder, quoted []string
	for i, column := range columns {
		placeholder = append(placeholder, fmt.Sprintf("$%d", i+1))
		quoted = append(quoted, quoteIdent(column))
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", quoteIdent(tablename), strings.Join(quoted, ","), strings.Join(placeholder, ","))
	return sql
}

func (p *Postgresql) tableExists(tableName string) bool {
	stmt := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname NOT IN ('information_schema','pg_catalog');"
	result, err := p.db.Exec(stmt, tableName)
	if err != nil {
		log.Printf("E! Error checking for existence of metric table %s: %v", tableName, err)
		return false
	}
	if count, _ := result.RowsAffected(); count == 1 {
		p.Tables[tableName] = true
		return true
	}
	return false
}

func (p *Postgresql) Write(metrics []telegraf.Metric) error {
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
		var js map[string]interface{}

		columns = append(columns, "time")
		values = append(values, metric.Time())

		if p.TagsAsJsonb {
			js = make(map[string]interface{})
			for column, value := range metric.Tags() {
				js[column] = value
			}

			if len(js) > 0 {
				d, err := json.Marshal(js)
				if err != nil {
					return err
				}

				columns = append(columns, "tags")
				values = append(values, d)
			}
		} else {
			for column, value := range metric.Tags() {
				if p.TagsAsForeignkeys {
					var value_id int

					query := fmt.Sprintf("SELECT %s FROM %s WHERE %s=$1", quoteIdent(column+"_id"), quoteIdent(tablename+"_"+column), quoteIdent(column))
					err := p.db.QueryRow(query, value).Scan(&value_id)
					if err != nil {
						log.Printf("W! Foreign key reference not found %s: %v", tablename, err)
						query := fmt.Sprintf("INSERT INTO %s(%s) VALUES($1) RETURNING %s", quoteIdent(tablename+"_"+column), quoteIdent(column), quoteIdent(column+"_id"))
						err := p.db.QueryRow(query, value).Scan(&value_id)
						if err != nil {
							return err
						}
					}

					columns = append(columns, column+"_id")
					values = append(values, value_id)
				} else {
					columns = append(columns, column)
					values = append(values, value)
				}
			}
		}

		if p.FieldsAsJsonb {
			js = make(map[string]interface{})
			for column, value := range metric.Fields() {
				js[column] = value
			}

			d, err := json.Marshal(js)
			if err != nil {
				return err
			}

			columns = append(columns, "fields")
			values = append(values, d)
		} else {
			for column, value := range metric.Fields() {
				columns = append(columns, column)
				values = append(values, value)
			}
		}

		sql := p.generateInsert(tablename, columns)
		_, err := p.db.Exec(sql, values...)
		if err != nil {
			fmt.Println("Error during insert", err)
			return err
		}
	}
	return nil
}

func init() {
	outputs.Add("postgresql", func() telegraf.Output { return newPostgresql() })
}

func newPostgresql() *Postgresql {
	return &Postgresql{
		TableTemplate: "CREATE TABLE {TABLE}({COLUMNS})",
		TagsAsJsonb:   true,
		FieldsAsJsonb: true,
	}
}
