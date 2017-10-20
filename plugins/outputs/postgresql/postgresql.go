package postgresql

import (
	"database/sql"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"strings"
)

type Postgresql struct {
	db      *sql.DB
	Address string
}

func (p *Postgresql) Connect() error {
	fmt.Println("Connect")

	db, err := sql.Open("pgx", p.Address)

	if err != nil {
		fmt.Println("DB Connect failed")
		return nil
	}
	fmt.Println("DB Connect")
	p.db = db

	return nil
}

func (p *Postgresql) Close() error {
	fmt.Println("Close")
	return nil
}

func (p *Postgresql) SampleConfig() string { return "" }
func (p *Postgresql) Description() string  { return "Send metrics to PostgreSQL" }

func (p *Postgresql) Write(metrics []telegraf.Metric) error {

	for _, m := range metrics {
		var keys, values []string
		for k, v := range m.Tags() {
			keys = append(keys, k)
			values = append(values, fmt.Sprintf("'%s'", v))
		}
		for k, v := range m.Fields() {
			keys = append(keys, k)
			switch value := v.(type) {
			case int:
				values = append(values, fmt.Sprintf("%d", value))
			case float64:
				values = append(values, fmt.Sprintf("%f", value))
			case string:
				values = append(values, fmt.Sprintf("'%s'", value))
			}
		}
		fmt.Printf("INSERT INTO %v.%v (%v) VALUES (%v);\n", m.Tags()["host"], m.Name(), strings.Join(keys, ","), strings.Join(values, ","))
	}

	return nil
}

func init() {
	outputs.Add("postgresql", func() telegraf.Output { return &Postgresql{} })
}
