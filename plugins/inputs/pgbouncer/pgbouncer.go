package pgbouncer

import (
	"bytes"
	"github.com/influxdata/telegraf/plugins/inputs/postgresql"

	// register in driver.
	_ "github.com/jackc/pgx/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type PgBouncer struct {
	postgresql.Service
}

var ignoredColumns = map[string]bool{"user": true, "database": true, "pool_mode": true,
	"avg_req": true, "avg_recv": true, "avg_sent": true, "avg_query": true,
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
  address = "host=localhost user=pgbouncer sslmode=disable"
`

func (p *PgBouncer) SampleConfig() string {
	return sampleConfig
}

func (p *PgBouncer) Description() string {
	return "Read metrics from one or many pgbouncer servers"
}

func (p *PgBouncer) Gather(acc telegraf.Accumulator) error {
	var (
		err     error
		query   string
		columns []string
	)

	query = `SHOW STATS`

	rows, err := p.DB.Query(query)
	if err != nil {
		return err
	}

	defer rows.Close()

	// grab the column information from the result
	if columns, err = rows.Columns(); err != nil {
		return err
	}

	for rows.Next() {
		tags, columnMap, err := p.accRow(rows, acc, columns)

		if err != nil {
			return err
		}

		fields := make(map[string]interface{})
		for col, val := range columnMap {
			_, ignore := ignoredColumns[col]
			if !ignore {
				fields[col] = *val
			}
		}
		acc.AddFields("pgbouncer", fields, tags)
	}

	query = `SHOW POOLS`

	poolRows, err := p.DB.Query(query)
	if err != nil {
		return err
	}

	defer poolRows.Close()

	// grab the column information from the result
	if columns, err = poolRows.Columns(); err != nil {
		return err
	}

	for poolRows.Next() {
		tags, columnMap, err := p.accRow(poolRows, acc, columns)
		if err != nil {
			return err
		}

		if s, ok := (*columnMap["user"]).(string); ok && s != "" {
			tags["user"] = s
		}

		if s, ok := (*columnMap["pool_mode"]).(string); ok && s != "" {
			tags["pool_mode"] = s
		}

		fields := make(map[string]interface{})
		for col, val := range columnMap {
			_, ignore := ignoredColumns[col]
			if !ignore {
				fields[col] = *val
			}
		}
		acc.AddFields("pgbouncer_pools", fields, tags)
	}

	return poolRows.Err()
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (p *PgBouncer) accRow(row scanner, acc telegraf.Accumulator, columns []string) (map[string]string,
	map[string]*interface{}, error) {
	var columnVars []interface{}
	var dbname bytes.Buffer

	// this is where we'll store the column name with its *interface{}
	columnMap := make(map[string]*interface{})

	for _, column := range columns {
		columnMap[column] = new(interface{})
	}

	// populate the array of interface{} with the pointers in the right order
	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[columns[i]])
	}

	// deconstruct array of variables and send to Scan
	err := row.Scan(columnVars...)

	if err != nil {
		return nil, nil, err
	}
	if columnMap["database"] != nil {
		// extract the database name from the column map
		dbname.WriteString((*columnMap["database"]).(string))
	} else {
		dbname.WriteString("postgres")
	}

	var tagAddress string
	tagAddress, err = p.SanitizedAddress()
	if err != nil {
		return nil, nil, err
	}

	// Return basic tags and the mapped columns
	return map[string]string{"server": tagAddress, "db": dbname.String()}, columnMap, nil
}

func init() {
	inputs.Add("pgbouncer", func() telegraf.Input {
		return &PgBouncer{
			Service: postgresql.Service{
				MaxIdle: 1,
				MaxOpen: 1,
				MaxLifetime: internal.Duration{
					Duration: 0,
				},
				IsPgBouncer: true,
			},
		}
	})
}
