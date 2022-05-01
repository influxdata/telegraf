package postgresql_extensible

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" //to register stdlib from PostgreSQL Driver and Toolkit

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
)

type Postgresql struct {
	postgresql.Service
	Databases          []string
	AdditionalTags     []string
	Timestamp          string
	Query              query
	Debug              bool
	PreparedStatements bool `toml:"prepared_statements"`

	Log telegraf.Logger
}

type query []struct {
	Sqlquery    string
	Script      string
	Version     int
	Withdbname  bool
	Tagvalue    string
	Measurement string
	Timestamp   string
}

var ignoredColumns = map[string]bool{"stats_reset": true}

func (p *Postgresql) Init() error {
	var err error
	for i := range p.Query {
		if p.Query[i].Sqlquery == "" {
			p.Query[i].Sqlquery, err = ReadQueryFromFile(p.Query[i].Script)
			if err != nil {
				return err
			}
		}
	}
	p.Service.IsPgBouncer = !p.PreparedStatements
	return nil
}

func (p *Postgresql) IgnoredColumns() map[string]bool {
	return ignoredColumns
}

func ReadQueryFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	query, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(query), err
}

func (p *Postgresql) Gather(acc telegraf.Accumulator) error {
	var (
		err        error
		sqlQuery   string
		queryAddon string
		dbVersion  int
		query      string
		measName   string
	)

	// Retrieving the database version
	query = `SELECT setting::integer / 100 AS version FROM pg_settings WHERE name = 'server_version_num'`
	if err = p.DB.QueryRow(query).Scan(&dbVersion); err != nil {
		dbVersion = 0
	}

	// We loop in order to process each query
	// Query is not run if Database version does not match the query version.
	for i := range p.Query {
		sqlQuery = p.Query[i].Sqlquery

		if p.Query[i].Measurement != "" {
			measName = p.Query[i].Measurement
		} else {
			measName = "postgresql"
		}

		if p.Query[i].Withdbname {
			if len(p.Databases) != 0 {
				queryAddon = fmt.Sprintf(` IN ('%s')`, strings.Join(p.Databases, "','"))
			} else {
				queryAddon = " is not null"
			}
		} else {
			queryAddon = ""
		}
		sqlQuery += queryAddon

		if p.Query[i].Version <= dbVersion {
			p.gatherMetricsFromQuery(acc, sqlQuery, p.Query[i].Tagvalue, p.Query[i].Timestamp, measName)
		}
	}
	return nil
}

func (p *Postgresql) gatherMetricsFromQuery(acc telegraf.Accumulator, sqlQuery string, tagValue string, timestamp string, measName string) {
	var columns []string

	rows, err := p.DB.Query(sqlQuery)
	if err != nil {
		acc.AddError(err)
		return
	}

	defer rows.Close()

	// grab the column information from the result
	if columns, err = rows.Columns(); err != nil {
		acc.AddError(err)
		return
	}

	p.AdditionalTags = nil
	if tagValue != "" {
		tagList := strings.Split(tagValue, ",")
		for t := range tagList {
			p.AdditionalTags = append(p.AdditionalTags, tagList[t])
		}
	}

	p.Timestamp = timestamp

	for rows.Next() {
		err = p.accRow(measName, rows, acc, columns)
		if err != nil {
			acc.AddError(err)
			break
		}
	}
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (p *Postgresql) accRow(measName string, row scanner, acc telegraf.Accumulator, columns []string) error {
	var (
		err        error
		columnVars []interface{}
		dbname     bytes.Buffer
		tagAddress string
		timestamp  time.Time
	)

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
	if err = row.Scan(columnVars...); err != nil {
		return err
	}

	if c, ok := columnMap["datname"]; ok && *c != nil {
		// extract the database name from the column map
		switch datname := (*c).(type) {
		case string:
			if _, err := dbname.WriteString(datname); err != nil {
				return err
			}
		default:
			if _, err := dbname.WriteString("postgres"); err != nil {
				return err
			}
		}
	} else {
		if _, err := dbname.WriteString("postgres"); err != nil {
			return err
		}
	}

	if tagAddress, err = p.SanitizedAddress(); err != nil {
		return err
	}

	// Process the additional tags
	tags := map[string]string{
		"server": tagAddress,
		"db":     dbname.String(),
	}

	// set default timestamp to Now
	timestamp = time.Now()

	fields := make(map[string]interface{})
COLUMN:
	for col, val := range columnMap {
		p.Log.Debugf("Column: %s = %T: %v\n", col, *val, *val)
		_, ignore := ignoredColumns[col]
		if ignore || *val == nil {
			continue
		}

		if col == p.Timestamp {
			if v, ok := (*val).(time.Time); ok {
				timestamp = v
			}
			continue
		}

		for _, tag := range p.AdditionalTags {
			if col != tag {
				continue
			}
			switch v := (*val).(type) {
			case string:
				tags[col] = v
			case []byte:
				tags[col] = string(v)
			case int64, int32, int:
				tags[col] = fmt.Sprintf("%d", v)
			default:
				p.Log.Debugf("Failed to add %q as additional tag", col)
			}
			continue COLUMN
		}

		if v, ok := (*val).([]byte); ok {
			fields[col] = string(v)
		} else {
			fields[col] = *val
		}
	}
	acc.AddFields(measName, fields, tags, timestamp)
	return nil
}

func init() {
	inputs.Add("postgresql_extensible", func() telegraf.Input {
		return &Postgresql{
			Service: postgresql.Service{
				MaxIdle:     1,
				MaxOpen:     1,
				MaxLifetime: config.Duration(0),
				IsPgBouncer: false,
			},
			PreparedStatements: true,
		}
	})
}
