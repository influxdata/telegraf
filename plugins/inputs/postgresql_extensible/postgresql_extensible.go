package postgresql_extensible

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/lib/pq"
)

type Postgresql struct {
	Address          string
	Databases        []string
	OrderedColumns   []string
	AllColumns       []string
	AdditionalTags   []string
	sanitizedAddress string
	Query            []struct {
		Sqlquery   string
		Version    int
		Withdbname bool
		Tagvalue   string
	}
}

type query []struct {
	Sqlquery   string
	Version    int
	Withdbname bool
	Tagvalue   string
}

var ignoredColumns = map[string]bool{"datid": true, "datname": true, "stats_reset": true}

var sampleConfig = `
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
  #
  ## All connection parameters are optional.  #
  ## Without the dbname parameter, the driver will default to a database
  ## with the same name as the user. This dbname is just for instantiating a
  ## connection with the server and doesn't restrict the databases we are trying
  ## to grab metrics for.
  #
  address = "host=localhost user=postgres sslmode=disable"
  ## A list of databases to pull metrics about. If not specified, metrics for all
  ## databases are gathered.
  ## databases = ["app_production", "testing"]
  #
  ## Define the toml config where the sql queries are stored
  ## New queries can be added, if the withdbname is set to true and there is no
  ## databases defined in the 'databases field', the sql query is ended by a
  ## 'is not null' in order to make the query succeed.
  ## Example :
  ## The sqlquery : "SELECT * FROM pg_stat_database where datname" become
  ## "SELECT * FROM pg_stat_database where datname IN ('postgres', 'pgbench')"
  ## because the databases variable was set to ['postgres', 'pgbench' ] and the
  ## withdbname was true. Be careful that if the withdbname is set to false you
  ## don't have to define the where clause (aka with the dbname) the tagvalue
  ## field is used to define custom tags (separated by comas)
  #
  ## Structure :
  ## [[inputs.postgresql_extensible.query]]
  ##   sqlquery string
  ##   version string
  ##   withdbname boolean
  ##   tagvalue string (coma separated)
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_database"
    version=901
    withdbname=false
    tagvalue=""
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_bgwriter"
    version=901
    withdbname=false
    tagvalue=""
`

func (p *Postgresql) SampleConfig() string {
	return sampleConfig
}

func (p *Postgresql) Description() string {
	return "Read metrics from one or many postgresql servers"
}

func (p *Postgresql) IgnoredColumns() map[string]bool {
	return ignoredColumns
}

var localhost = "host=localhost sslmode=disable"

func (p *Postgresql) Gather(acc telegraf.Accumulator) error {

	var sql_query string
	var query_addon string
	var db_version int
	var query string
	var tag_value string

	if p.Address == "" || p.Address == "localhost" {
		p.Address = localhost
	}

	db, err := sql.Open("postgres", p.Address)
	if err != nil {
		return err
	}

	defer db.Close()

	// Retreiving the database version

	query = `select substring(setting from 1 for 3) as version from pg_settings where name='server_version_num'`
	err = db.QueryRow(query).Scan(&db_version)
	if err != nil {
		return err
	}
	// We loop in order to process each query
	// Query is not run if Database version does not match the query version.

	for i := range p.Query {
		sql_query = p.Query[i].Sqlquery
		tag_value = p.Query[i].Tagvalue

		if p.Query[i].Withdbname {
			if len(p.Databases) != 0 {
				query_addon = fmt.Sprintf(` IN ('%s')`,
					strings.Join(p.Databases, "','"))
			} else {
				query_addon = " is not null"
			}
		} else {
			query_addon = ""
		}
		sql_query += query_addon

		if p.Query[i].Version <= db_version {
			rows, err := db.Query(sql_query)
			if err != nil {
				return err
			}

			defer rows.Close()

			// grab the column information from the result
			p.OrderedColumns, err = rows.Columns()
			if err != nil {
				return err
			} else {
				for _, v := range p.OrderedColumns {
					p.AllColumns = append(p.AllColumns, v)
				}
			}
			p.AdditionalTags = nil
			if tag_value != "" {
				tag_list := strings.Split(tag_value, ",")
				for t := range tag_list {
					p.AdditionalTags = append(p.AdditionalTags, tag_list[t])
				}
			}

			for rows.Next() {
				err = p.accRow(rows, acc)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

var passwordKVMatcher, _ = regexp.Compile("password=\\S+ ?")

func (p *Postgresql) SanitizedAddress() (_ string, err error) {
	var canonicalizedAddress string
	if strings.HasPrefix(p.Address, "postgres://") || strings.HasPrefix(p.Address, "postgresql://") {
		canonicalizedAddress, err = pq.ParseURL(p.Address)
		if err != nil {
			return p.sanitizedAddress, err
		}
	} else {
		canonicalizedAddress = p.Address
	}
	p.sanitizedAddress = passwordKVMatcher.ReplaceAllString(canonicalizedAddress, "")

	return p.sanitizedAddress, err
}

func (p *Postgresql) accRow(row scanner, acc telegraf.Accumulator) error {
	var columnVars []interface{}
	var dbname bytes.Buffer

	// this is where we'll store the column name with its *interface{}
	columnMap := make(map[string]*interface{})

	for _, column := range p.OrderedColumns {
		columnMap[column] = new(interface{})
	}

	// populate the array of interface{} with the pointers in the right order
	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[p.OrderedColumns[i]])
	}

	// deconstruct array of variables and send to Scan
	err := row.Scan(columnVars...)

	if err != nil {
		return err
	}
	if columnMap["datname"] != nil {
		// extract the database name from the column map
		dbnameChars := (*columnMap["datname"]).([]uint8)
		for i := 0; i < len(dbnameChars); i++ {
			dbname.WriteString(string(dbnameChars[i]))
		}
	} else {
		dbname.WriteString("postgres")
	}

	var tagAddress string
	tagAddress, err = p.SanitizedAddress()
	if err != nil {
		return err
	}

	// Process the additional tags

	tags := map[string]string{}
	tags["server"] = tagAddress
	tags["db"] = dbname.String()
	var isATag int
	fields := make(map[string]interface{})
	for col, val := range columnMap {
		_, ignore := ignoredColumns[col]
		//if !ignore && *val != "" {
		if !ignore {
			isATag = 0
			for tag := range p.AdditionalTags {
				if col == p.AdditionalTags[tag] {
					isATag = 1
					value_type_p := fmt.Sprintf(`%T`, *val)
					if value_type_p == "[]uint8" {
						tags[col] = fmt.Sprintf(`%s`, *val)
					} else if value_type_p == "int64" {
						tags[col] = fmt.Sprintf(`%v`, *val)
					}
				}
			}
			if isATag == 0 {
				fields[col] = *val
			}
		}
	}
	acc.AddFields("postgresql", fields, tags)
	return nil
}

func init() {
	inputs.Add("postgresql_extensible", func() telegraf.Input {
		return &Postgresql{}
	})
}
