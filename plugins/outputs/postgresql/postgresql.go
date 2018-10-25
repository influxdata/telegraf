package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"
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
	TagTableSuffix    string
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

func deriveDatatype(value interface{}) string {
	var datatype string

	switch value.(type) {
	case bool:
		datatype = "boolean"
	case uint64:
		datatype = "int8"
	case int64:
		datatype = "int8"
	case float64:
		datatype = "float8"
	case string:
		datatype = "text"
	default:
		datatype = "text"
		log.Printf("E! Unknown datatype %T(%v)", value, value)
	}
	return datatype
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
  # table_template = "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS})"
  ## Example for timescaledb
  # table_template = "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS}); SELECT create_hypertable({TABLELITERAL},'time',chunk_time_interval := '1 week'::interval,if_not_exists := true);"

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
	columns = append(columns, "time timestamptz")

	// handle tags if necessary
	if len(metric.Tags()) > 0 {
		if p.TagsAsForeignkeys {
			// tags in separate table
			var tag_columns []string
			var tag_columndefs []string
			columns = append(columns, "tag_id int")

			if p.TagsAsJsonb {
				tag_columns = append(tag_columns, "tags")
				tag_columndefs = append(tag_columndefs, "tags jsonb")
			} else {
				for column, _ := range metric.Tags() {
					tag_columns = append(tag_columns, quoteIdent(column))
					tag_columndefs = append(tag_columndefs, fmt.Sprintf("%s text", quoteIdent(column)))
				}
			}
			table := quoteIdent(metric.Name() + p.TagTableSuffix)
			sql = append(sql, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(tag_id serial primary key,%s,UNIQUE(%s))", table, strings.Join(tag_columndefs, ","), strings.Join(tag_columns, ",")))
		} else {
			// tags in measurement table
			if p.TagsAsJsonb {
				columns = append(columns, "tags jsonb")
			} else {
				for column, _ := range metric.Tags() {
					pk = append(pk, quoteIdent(column))
					columns = append(columns, fmt.Sprintf("%s text", quoteIdent(column)))
				}
			}
		}
	}

	if p.FieldsAsJsonb {
		columns = append(columns, "fields jsonb")
	} else {
		var datatype string
		for column, v := range metric.Fields() {
			datatype = deriveDatatype(v)
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

	return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", quoteIdent(tablename), strings.Join(quoted, ","), strings.Join(placeholder, ","))
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
	batches := make(map[string][]interface{})
	params := make(map[string][]string)
	colmap := make(map[string][]string)
	tabmap := make(map[string]string)

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

		columns := []string{"time"}
		values := []interface{}{metric.Time()}
		var js map[string]interface{}

		if len(metric.Tags()) > 0 {
			if p.TagsAsForeignkeys {
				// tags in separate table
				var tag_id int
				var where_columns []string
				var where_values []interface{}

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

						where_columns = append(where_columns, "tags")
						where_values = append(where_values, d)
					}
				} else {
					for column, value := range metric.Tags() {
						where_columns = append(where_columns, column)
						where_values = append(where_values, value)
					}
				}

				var where_parts []string
				for i, column := range where_columns {
					where_parts = append(where_parts, fmt.Sprintf("%s = $%d", quoteIdent(column), i+1))
				}
				query := fmt.Sprintf("SELECT tag_id FROM %s WHERE %s", quoteIdent(tablename+p.TagTableSuffix), strings.Join(where_parts, " AND "))

				err := p.db.QueryRow(query, where_values...).Scan(&tag_id)
				if err != nil {
					query := p.generateInsert(tablename+p.TagTableSuffix, where_columns) + " RETURNING tag_id"
					err := p.db.QueryRow(query, where_values...).Scan(&tag_id)
					if err != nil {
						return err
					}
				}

				columns = append(columns, "tag_id")
				values = append(values, tag_id)
			} else {
				// tags in measurement table
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
					var keys []string
					fields := metric.Tags()
					for column := range fields {
						keys = append(keys, column)
					}
					sort.Strings(keys)
					for _, column := range keys {
						columns = append(columns, column)
						values = append(values, fields[column])
					}
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
			var keys []string
			fields := metric.Fields()
			for column := range fields {
				keys = append(keys, column)
			}
			sort.Strings(keys)
			for _, column := range keys {
				columns = append(columns, column)
				values = append(values, fields[column])
			}
		}

		var table_and_cols string
		var placeholder, quoted_columns []string
		for _, column := range columns {
			quoted_columns = append(quoted_columns, quoteIdent(column))
		}
		table_and_cols = fmt.Sprintf("%s(%s)", quoteIdent(tablename), strings.Join(quoted_columns, ","))
		batches[table_and_cols] = append(batches[table_and_cols], values...)
		for i, _ := range columns {
			i += len(params[table_and_cols]) * len(columns)
			placeholder = append(placeholder, fmt.Sprintf("$%d", i+1))
		}
		params[table_and_cols] = append(params[table_and_cols], strings.Join(placeholder, ","))
		colmap[table_and_cols] = columns
		tabmap[table_and_cols] = tablename
	}

	for table_and_cols, values := range batches {
		sql := fmt.Sprintf("INSERT INTO %s VALUES (%s)", table_and_cols, strings.Join(params[table_and_cols], "),("))
		_, err := p.db.Exec(sql, values...)
		if err != nil {
			// check if insert error was caused by column mismatch
			if p.FieldsAsJsonb == false {
				log.Printf("E! Error during insert: %v", err)
				tablename := tabmap[table_and_cols]
				columns := colmap[table_and_cols]
				var quoted_columns []string
				for _, column := range columns {
					quoted_columns = append(quoted_columns, quoteLiteral(column))
				}
				query := "SELECT c FROM unnest(array[%s]) AS c WHERE NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE column_name=c AND table_schema=$1 AND table_name=$2)"
				query = fmt.Sprintf(query, strings.Join(quoted_columns, ","))
				result, err := p.db.Query(query, "public", tablename)
				defer result.Close()
				if err != nil {
					return err
				}
				// some columns are missing

				var column, datatype string
				for result.Next() {
					err := result.Scan(&column)
					if err != nil {
						log.Println(err)
					}
					for i, name := range columns {
						if name == column {
							datatype = deriveDatatype(values[i])
						}
					}
					query := "ALTER TABLE %s.%s ADD COLUMN IF NOT EXISTS %s %s;"
					_, err = p.db.Exec(fmt.Sprintf(query, quoteIdent("public"), quoteIdent(tablename), quoteIdent(column), datatype))
					if err != nil {
						return err
					}
				}
			}

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
		TableTemplate:  "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS})",
		TagsAsJsonb:    true,
		TagTableSuffix: "_tag",
		FieldsAsJsonb:  true,
	}
}
