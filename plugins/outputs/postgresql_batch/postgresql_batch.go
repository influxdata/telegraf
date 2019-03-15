package postgresql_batch

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type PostgresqlBatch struct {
	db      *sql.DB
	Address string
	Inserts map[string]string
	Columns map[string][]string
}

func (p *PostgresqlBatch) Connect() error {
	db, err := sql.Open("pgx", p.Address)
	if err != nil {
		return err
	}
	p.db = db

	p.Inserts = make(map[string]string)
	p.Columns = make(map[string][]string)

	return nil
}

func (p *PostgresqlBatch) Close() error {
	return p.db.Close()
}

func quoteIdent(name string) string {
	return pgx.Identifier{name}.Sanitize()
}

var sampleConfig = `
  # Send metrics to PostgreSQL using COPY
  address = "host=localhost user=postgres sslmode=verify-full"
`

func (p *PostgresqlBatch) SampleConfig() string { return sampleConfig }
func (p *PostgresqlBatch) Description() string  { return "Send metrics to Postgresql in batch" }

func (p *PostgresqlBatch) generateInsert(tablename string, columns []string) string {
	var quoted []string
	for _, column := range columns {
		quoted = append(quoted, quoteIdent(column))
	}

	return fmt.Sprintf("INSERT INTO %s(%s) VALUES ", quoteIdent(tablename), strings.Join(quoted, ","))
}

func (p *PostgresqlBatch) buildTableInsert(metric telegraf.Metric) {
	table := metric.Name()
	if p.Inserts[table] == "" {
		for key, _ := range metric.Tags() {
			p.Columns[table] = append(p.Columns[table], key)
		}
		for key, _ := range metric.Fields() {
			p.Columns[table] = append(p.Columns[table], key)
		}
		p.Inserts[table] = p.generateInsert(table, append(p.Columns[table], "time"))
	}
}

func quoted(value interface{}) interface{} {
	switch value.(type) {
	case string:
		return "'" + value.(string) + "'"
	case time.Time:
		return quoted(value.(time.Time).Format("2006-01-02 15:04:05"))
	default:
		return value
	}
}

func joinValues(values []interface{}) string {
	strs := make([]string, len(values))
	for i, value := range values {
		strs[i] = fmt.Sprintf("%v", value)
	}
	return strings.Join(strs, ", ")
}

func buildValues(metric telegraf.Metric, columns []string) string {
	var values []interface{}
	mapString := metric.Tags()
	for key, value := range metric.Fields() {
		mapString[key] = fmt.Sprintf("%v", value)
	}
	for _, column := range columns {
		values = append(values, quoted(mapString[column]))
	}
	values = append(values, quoted(metric.Time()))
	return "(" + joinValues(values) + ")"
}

func (p *PostgresqlBatch) Write(metrics []telegraf.Metric) error {
	values := make(map[string][]string)
	for _, metric := range metrics {
		p.buildTableInsert(metric)
		table := metric.Name()
		values[table] = append(values[table], buildValues(metric, p.Columns[table]))
	}
	for table, values := range values {
		if len(values) == 0 {
			continue
		}
		sql := p.Inserts[table] + strings.Join(values, ",")
		_, err := p.db.Exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	outputs.Add("postgresql_batch", func() telegraf.Output { return newPostgresqlBatch() })
}

func newPostgresqlBatch() *PostgresqlBatch {
	return &PostgresqlBatch{}
}
