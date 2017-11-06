package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	// register in driver.
	_ "github.com/jackc/pgx/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Postgresql struct {
	db                *sql.DB
	Address           string
	IgnoredTags       []string
	TagsAsForeignkeys bool
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

  ## A list of tags to exclude from storing. If not specified, all tags are stored.
  # ignored_tags = ["foo", "bar"]

  ## Store tags as foreign keys in the metrics table. Default is false.
  # tags_as_foreignkeys = false

`

func (p *Postgresql) SampleConfig() string { return sampleConfig }
func (p *Postgresql) Description() string  { return "Send metrics to PostgreSQL" }

func (p *Postgresql) generateCreateTable(metric telegraf.Metric) string {
	var columns []string
	var pk []string
	var sql []string

	pk = append(pk, "time")
	columns = append(columns, "time timestamptz")

	for column, _ := range metric.Tags() {
		if contains(p.IgnoredTags, column) {
			continue
		}
		if p.TagsAsForeignkeys {
			pk = append(pk, column+"_id")
			columns = append(columns, fmt.Sprintf("%s_id int8", column))
			sql = append(sql, fmt.Sprintf("CREATE TABLE %s_%s(%s_id serial primary key,%s text unique)", metric.Name(), column, column, column))
		} else {
			pk = append(pk, column)
			columns = append(columns, fmt.Sprintf("%s text", column))
		}
	}

	var datatype string
	for column, v := range metric.Fields() {
		switch v.(type) {
		case int64:
			datatype = "int8"
		case float64:
			datatype = "float8"
		}
		columns = append(columns, fmt.Sprintf("%s %s", column, datatype))
	}

	sql = append(sql, fmt.Sprintf("CREATE TABLE %[1]s(%[2]s,PRIMARY KEY(%[3]s))", metric.Name(), strings.Join(columns, ","), strings.Join(pk, ",")))
	return strings.Join(sql, ";")
}

func (p *Postgresql) generateInsert(tablename string, columns []string, values []interface{}) (string, []interface{}) {

	var placeholder []string
	for i := 1; i <= len(values); i++ {
		placeholder = append(placeholder, fmt.Sprintf("$%d", i))
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", tablename, strings.Join(columns, ","), strings.Join(placeholder, ","))
	return sql, values
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

		columns = append(columns, "time")
		values = append(values, metric.Time())

		for column, value := range metric.Tags() {
			if contains(p.IgnoredTags, column) {
				continue
			}

			if p.TagsAsForeignkeys {
				var value_id int

				query := fmt.Sprintf("SELECT %s_id FROM %s_%s WHERE %s=$1", column, tablename, column, column)
				err := p.db.QueryRow(query, value).Scan(&value_id)
				if err != nil {
					query := fmt.Sprintf("INSERT INTO %s_%s(%s) VALUES($1) RETURNING %s_id", tablename, column, column, column)
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

		for column, value := range metric.Fields() {
			columns = append(columns, column)
			values = append(values, value)
		}

		var placeholder []string
		for i := 1; i <= len(values); i++ {
			placeholder = append(placeholder, fmt.Sprintf("$%d", i))
		}

		sql, values := p.generateInsert(tablename, columns, values)
		_, err := p.db.Exec(sql, values...)
		if err != nil {
			fmt.Println("Error during insert", err)
			return err
		}
	}
	return nil
}

func init() {
	outputs.Add("postgresql", func() telegraf.Output { return &Postgresql{} })
}
