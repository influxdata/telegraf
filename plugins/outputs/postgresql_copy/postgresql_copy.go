package postgresql_copy

import (
	"database/sql"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/lib/pq"
)

type PostgresqlCopy struct {
	db                 *sql.DB
	Address            string
	Columns            map[string][]string
	IgnoreInsertErrors bool
}

func (p *PostgresqlCopy) Connect() error {
	db, err := sql.Open("postgres", p.Address)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

func (p *PostgresqlCopy) Close() error {
	return p.db.Close()
}

var sampleConfig = `
  # Send metrics to PostgreSQL using COPY
  address = "postgres://USER:PWD@HOST:PORT/DATABASE?sslmode=disable"
`

func (p *PostgresqlCopy) SampleConfig() string { return sampleConfig }
func (p *PostgresqlCopy) Description() string  { return "Send metrics to Postgres using Copy" }

func (p *PostgresqlCopy) buildColumns(metrics []telegraf.Metric) {
	table_columns := make(map[string]map[string]bool)
	for _, metric := range metrics {
		table := metric.Name()
		if table_columns[table] == nil {
			table_columns[table] = map[string]bool{}
		}
		for key := range metric.Fields() {
			table_columns[table][key] = true
		}
		for key := range metric.Tags() {
			table_columns[table][key] = true
		}
	}

	p.Columns = make(map[string][]string)
	for table, columns := range table_columns {
		for column := range columns {
			p.Columns[table] = append(p.Columns[table], column)
		}
	}
}

func buildValues(metric telegraf.Metric, columns []string) []interface{} {
	var values []interface{}
	all_metric := metric.Fields()
	for key, value := range metric.Tags() {
		all_metric[key] = value
	}
	for _, column := range columns {
		values = append(values, all_metric[column])
	}
	values = append(values, metric.Time())
	return values
}

func (p *PostgresqlCopy) Write(metrics []telegraf.Metric) error {
	p.buildColumns(metrics)
	tables := make(map[string][][]interface{})
	for _, metric := range metrics {
		table := metric.Name()
		tables[table] = append(tables[table], buildValues(metric, p.Columns[table]))
	}

	txn, err := p.db.Begin()
	if err != nil {
		return err
	}
	for table, values := range tables {
		if len(values) == 0 {
			continue
		}

		columns := append(p.Columns[table], "time")
		stmt, err := txn.Prepare(pq.CopyIn(table, columns...))
		if err != nil {
			if p.IgnoreInsertErrors {
				log.Printf("E! Error in stmt execute %s", err)
				txn, err = p.db.Begin()
				continue
			} else {
				return err
			}
		}

		for _, value := range values {
			_, err = stmt.Exec(value...)
			if err == nil {
				continue
			}

			if p.IgnoreInsertErrors {
				log.Printf("E! Could not insert into %s: %s", table, values)
				continue
			}
			return err
		}
		_, err = stmt.Exec()
		if err == nil {
			continue
		}
		if p.IgnoreInsertErrors {
			log.Printf("E! Error in stmt execute %s", err)
			continue
		}
	}
	err = txn.Commit()
	if err == nil {
		return nil
	}
	if p.IgnoreInsertErrors {
		log.Printf("E! Error in commit %s", err)
		return nil
	}
	return err
}

func init() {
	outputs.Add("postgresql_copy", func() telegraf.Output { return newPostgresqlCopy() })
}

func newPostgresqlCopy() *PostgresqlCopy {
	return &PostgresqlCopy{}
}
