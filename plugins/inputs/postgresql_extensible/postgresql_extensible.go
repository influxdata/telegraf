//go:generate ../../../tools/readme_config_includer/generator
package postgresql_extensible

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	// Required for SQL framework driver
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/postgresql"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Postgresql struct {
	Databases          []string `deprecated:"1.22.4;use the sqlquery option to specify database to use"`
	AdditionalTags     []string
	Timestamp          string
	Query              query
	Debug              bool
	PreparedStatements bool `toml:"prepared_statements"`
	Log                telegraf.Logger
	postgresql.Config

	service *postgresql.Service
}

type query []struct {
	Sqlquery    string
	Script      string
	Version     int  `deprecated:"1.28.0;use minVersion to specify minimal DB version this query supports"`
	MinVersion  int  `toml:"min_version"`
	MaxVersion  int  `toml:"max_version"`
	Withdbname  bool `deprecated:"1.22.4;use the sqlquery option to specify database to use"`
	Tagvalue    string
	Measurement string
	Timestamp   string
}

var ignoredColumns = map[string]bool{"stats_reset": true}

func (*Postgresql) SampleConfig() string {
	return sampleConfig
}

func (p *Postgresql) Init() error {
	var err error
	for i := range p.Query {
		if p.Query[i].Sqlquery == "" {
			p.Query[i].Sqlquery, err = ReadQueryFromFile(p.Query[i].Script)
			if err != nil {
				return err
			}
		}
		if p.Query[i].MinVersion == 0 {
			p.Query[i].MinVersion = p.Query[i].Version
		}
	}
	p.Config.IsPgBouncer = !p.PreparedStatements

	service, err := p.Config.CreateService()
	if err != nil {
		return err
	}
	p.service = service

	return nil
}

func (p *Postgresql) Start(_ telegraf.Accumulator) error {
	return p.service.Start()
}

func (p *Postgresql) Stop() {
	p.service.Stop()
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
		sqlQuery   string
		queryAddon string
		dbVersion  int
		query      string
		measName   string
	)

	// Retrieving the database version
	query = `SELECT setting::integer / 100 AS version FROM pg_settings WHERE name = 'server_version_num'`
	if err := p.service.DB.QueryRow(query).Scan(&dbVersion); err != nil {
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

		maxVer := p.Query[i].MaxVersion

		if p.Query[i].MinVersion <= dbVersion && (maxVer == 0 || maxVer > dbVersion) {
			p.gatherMetricsFromQuery(acc, sqlQuery, p.Query[i].Tagvalue, p.Query[i].Timestamp, measName)
		}
	}
	return nil
}

func (p *Postgresql) gatherMetricsFromQuery(acc telegraf.Accumulator, sqlQuery string, tagValue string, timestamp string, measName string) {
	var columns []string

	rows, err := p.service.DB.Query(sqlQuery)
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
		p.AdditionalTags = append(p.AdditionalTags, tagList...)
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
		dbname    bytes.Buffer
		timestamp time.Time
	)

	// this is where we'll store the column name with its *interface{}
	columnMap := make(map[string]*interface{})

	for _, column := range columns {
		columnMap[column] = new(interface{})
	}

	columnVars := make([]interface{}, 0, len(columnMap))
	// populate the array of interface{} with the pointers in the right order
	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[columns[i]])
	}

	// deconstruct array of variables and send to Scan
	if err := row.Scan(columnVars...); err != nil {
		return err
	}

	if c, ok := columnMap["datname"]; ok && *c != nil {
		// extract the database name from the column map
		switch datname := (*c).(type) {
		case string:
			dbname.WriteString(datname)
		default:
			dbname.WriteString(p.service.ConnectionDatabase)
		}
	} else {
		dbname.WriteString(p.service.ConnectionDatabase)
	}

	// Process the additional tags
	tags := map[string]string{
		"server": p.service.SanitizedAddress,
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
			case bool:
				tags[col] = strconv.FormatBool(v)
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
			Config: postgresql.Config{
				MaxIdle: 1,
				MaxOpen: 1,
			},
			PreparedStatements: true,
		}
	})
}
