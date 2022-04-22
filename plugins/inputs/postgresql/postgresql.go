package postgresql

import (
	"bytes"
	"fmt"
	"strings"

	// register in driver.
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Postgresql struct {
	Service
	Databases          []string `toml:"databases"`
	IgnoredDatabases   []string `toml:"ignored_databases"`
	PreparedStatements bool     `toml:"prepared_statements"`
}

var ignoredColumns = map[string]bool{"stats_reset": true}

func (p *Postgresql) IgnoredColumns() map[string]bool {
	return ignoredColumns
}

func (p *Postgresql) Init() error {
	p.Service.IsPgBouncer = !p.PreparedStatements
	return nil
}

func (p *Postgresql) Gather(acc telegraf.Accumulator) error {
	var (
		err     error
		query   string
		columns []string
	)

	if len(p.Databases) == 0 && len(p.IgnoredDatabases) == 0 {
		query = `SELECT * FROM pg_stat_database`
	} else if len(p.IgnoredDatabases) != 0 {
		query = fmt.Sprintf(`SELECT * FROM pg_stat_database WHERE datname NOT IN ('%s')`,
			strings.Join(p.IgnoredDatabases, "','"))
	} else {
		query = fmt.Sprintf(`SELECT * FROM pg_stat_database WHERE datname IN ('%s')`,
			strings.Join(p.Databases, "','"))
	}

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
		err = p.accRow(rows, acc, columns)
		if err != nil {
			return err
		}
	}

	query = `SELECT * FROM pg_stat_bgwriter`

	bgWriterRow, err := p.DB.Query(query)
	if err != nil {
		return err
	}

	defer bgWriterRow.Close()

	// grab the column information from the result
	if columns, err = bgWriterRow.Columns(); err != nil {
		return err
	}

	for bgWriterRow.Next() {
		err = p.accRow(bgWriterRow, acc, columns)
		if err != nil {
			return err
		}
	}

	return bgWriterRow.Err()
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (p *Postgresql) accRow(row scanner, acc telegraf.Accumulator, columns []string) error {
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
		return err
	}
	if columnMap["datname"] != nil {
		// extract the database name from the column map
		if dbNameStr, ok := (*columnMap["datname"]).(string); ok {
			if _, err := dbname.WriteString(dbNameStr); err != nil {
				return err
			}
		} else {
			// PG 12 adds tracking of global objects to pg_stat_database
			if _, err := dbname.WriteString("postgres_global"); err != nil {
				return err
			}
		}
	} else {
		if _, err := dbname.WriteString("postgres"); err != nil {
			return err
		}
	}

	var tagAddress string
	tagAddress, err = p.SanitizedAddress()
	if err != nil {
		return err
	}

	tags := map[string]string{"server": tagAddress, "db": dbname.String()}

	fields := make(map[string]interface{})
	for col, val := range columnMap {
		_, ignore := ignoredColumns[col]
		if !ignore {
			fields[col] = *val
		}
	}
	acc.AddFields("postgresql", fields, tags)

	return nil
}

func init() {
	inputs.Add("postgresql", func() telegraf.Input {
		return &Postgresql{
			Service: Service{
				MaxIdle:     1,
				MaxOpen:     1,
				MaxLifetime: config.Duration(0),
			},
			PreparedStatements: true,
		}
	})
}
