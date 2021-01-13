package odyssey

import (
	"bytes"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
	_ "github.com/jackc/pgx/stdlib" // register driver
)

type Odyssey struct {
	postgresql.Service
	IncludeQuery []string `toml:"include_query"`
}

var ignoredColumns = map[string]bool{"user": true, "database": true, "pool_mode": true,
	"avg_req": true, "avg_recv": true, "avg_sent": true, "avg_query": true,
}

var sampleConfig = `
  ## specify address via a url matching:
  ##   postgres://[console[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=console password=... sslmode=... dbname=app_production
  ##
  ## include_query = []string - commands for psql
  ##
  ## All connection parameters are optional.
  ##
  address = "host=localhost user=postgresql sslmode=disable dbname=console port=6432"

  include_query = ['SHOW STATS', 'SHOW POOLS']
`

func (p *Odyssey) SampleConfig() string {
	return sampleConfig
}

func (p *Odyssey) Description() string {
	return "Read metrics from Odyssey"
}

func (p *Odyssey) Gather(acc telegraf.Accumulator) error {
	var (
		err     error
		query   string
		columns []string
	)

	for _, value := range p.IncludeQuery {
		query = value

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

			if user, ok := columnMap["user"]; ok {
				if s, ok := (*user).(string); ok && s != "" {
					tags["user"] = s
				}
			}

			if poolMode, ok := columnMap["pool_mode"]; ok {
				if s, ok := (*poolMode).(string); ok && s != "" {
					tags["pool_mode"] = s
				}
			}

			fields := make(map[string]interface{})
			for col, val := range columnMap {
				_, ignore := ignoredColumns[col]
				if !ignore {
					fields[col] = *val
				}
			}
			acc.AddFields(query, fields, tags)
		}
	}

	return err
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (p *Odyssey) accRow(row scanner, acc telegraf.Accumulator, columns []string) (map[string]string,
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
	inputs.Add("odyssey", func() telegraf.Input {
		return &Odyssey{
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
