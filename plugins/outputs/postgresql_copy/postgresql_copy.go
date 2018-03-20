package postgresql_copy

import (
	"database/sql"

	"github.com/lib/pq"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type PostgresqlCopy struct {
	db      *sql.DB
	Address string
	Columns map[string][]string
}

func (p *PostgresqlCopy) Connect() error {
	db, err := sql.Open("postgres", p.Address)
	if err != nil {
		return err
	}
	p.db = db
	p.Columns = make(map[string][]string)
	return nil
}

func (p *PostgresqlCopy) Close() error {
	return p.db.Close()
}

var sampleConfig = `
  # Send metrics to PostgreSQL using COPY
  [[outputs.postgresql_copy]]
    address = "postgres://USER:PWD@HOST:PORT/DATABASE?sslmode=disable"
`

func (p *PostgresqlCopy) SampleConfig() string { return sampleConfig }
func (p *PostgresqlCopy) Description() string  { return "Send metrics to Postgres using Copy" }

func (p *PostgresqlCopy) buildColumns(table string, metric telegraf.Metric) {
	if len(p.Columns[table]) != 0 {
		return
	}
	for key, _ := range metric.Fields() {
		p.Columns[table] = append(p.Columns[table], key)
	}
	for key, _ := range metric.Tags() {
		p.Columns[table] = append(p.Columns[table], key)
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
	tables := make(map[string][][]interface{})
	for _, metric := range metrics {
		table := metric.Name()
		p.buildColumns(table, metric)
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
		stmt, _ := txn.Prepare(pq.CopyIn(table, columns...))
		for _, value := range values {
			_, err = stmt.Exec(value...)
			if err != nil {
				return err
			}
		}
		_, err = stmt.Exec()
		if err != nil {
			return err
		}
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	outputs.Add("postgresql_copy", func() telegraf.Output { return newPostgresqlCopy() })
}

func newPostgresqlCopy() *PostgresqlCopy {
	return &PostgresqlCopy{}
}
