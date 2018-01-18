package postgresql_extensible

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	// register in driver.
	_ "github.com/jackc/pgx/stdlib"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
)

type Postgresql struct {
	Address          string
	Outputaddress    string
	Databases        []string
	OrderedColumns   []string
	AllColumns       []string
	AdditionalTags   []string
	sanitizedAddress string
	Query            []struct {
		Sqlquery    string
		Version     int
		Withdbname  bool
		Tagvalue    string
		Measurement string
	}
	Debug bool
}

type query []struct {
	Sqlquery    string
	Version     int
	Withdbname  bool
	Tagvalue    string
	Measurement string
}

var ignoredColumns = map[string]bool{"stats_reset": true}

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
  # outputaddress = "db01"
  ## A custom name for the database that will be used as the "server" tag in the
  ## measurement output. If not specified, a default one generated from
  ## the connection address is used.
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
  ## field is used to define custom tags (separated by commas)
  ## The optional "measurement" value can be used to override the default
  ## output measurement name ("postgresql").
  #
  ## Structure :
  ## [[inputs.postgresql_extensible.query]]
  ##   sqlquery string
  ##   version string
  ##   withdbname boolean
  ##   tagvalue string (comma separated)
  ##   measurement string
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_database"
    version=901
    withdbname=false
    tagvalue=""
    measurement=""
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_bgwriter"
    version=901
    withdbname=false
    tagvalue="postgresql.stats"
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
	var (
		err         error
		db          *sql.DB
		sql_query   string
		query_addon string
		db_version  int
		query       string
		tag_value   string
		meas_name   string
	)

	if p.Address == "" || p.Address == "localhost" {
		p.Address = localhost
	}

	if db, err = sql.Open("pgx", p.Address); err != nil {
		return err
	}
	defer db.Close()

	// Retreiving the database version

	query = `select substring(setting from 1 for 3) as version from pg_settings where name='server_version_num'`
	err = db.QueryRow(query).Scan(&db_version)
	if err != nil {
		db_version = 0
	}
	// We loop in order to process each query
	// Query is not run if Database version does not match the query version.

	for i := range p.Query {
		sql_query = p.Query[i].Sqlquery
		tag_value = p.Query[i].Tagvalue
		if p.Query[i].Measurement != "" {
			meas_name = p.Query[i].Measurement
		} else {
			meas_name = "postgresql"
		}

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
				acc.AddError(err)
				continue
			}

			defer rows.Close()

			// grab the column information from the result
			p.OrderedColumns, err = rows.Columns()
			if err != nil {
				acc.AddError(err)
				continue
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
				err = p.accRow(meas_name, rows, acc)
				if err != nil {
					acc.AddError(err)
					break
				}
			}
		}
	}
	return nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

var KVMatcher, _ = regexp.Compile("(password|sslcert|sslkey|sslmode|sslrootcert)=\\S+ ?")

func (p *Postgresql) SanitizedAddress() (_ string, err error) {
	if p.Outputaddress != "" {
		return p.Outputaddress, nil
	}
	var canonicalizedAddress string
	if strings.HasPrefix(p.Address, "postgres://") || strings.HasPrefix(p.Address, "postgresql://") {
		canonicalizedAddress, err = postgresql.ParseURL(p.Address)
		if err != nil {
			return p.sanitizedAddress, err
		}
	} else {
		canonicalizedAddress = p.Address
	}
	p.sanitizedAddress = KVMatcher.ReplaceAllString(canonicalizedAddress, "")

	return p.sanitizedAddress, err
}

func (p *Postgresql) accRow(meas_name string, row scanner, acc telegraf.Accumulator) error {
	var (
		err        error
		columnVars []interface{}
		dbname     bytes.Buffer
		tagAddress string
	)

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
	if err = row.Scan(columnVars...); err != nil {
		return err
	}

	if columnMap["datname"] != nil {
		// extract the database name from the column map
		dbname.WriteString((*columnMap["datname"]).(string))
	} else {
		dbname.WriteString("postgres")
	}

	if tagAddress, err = p.SanitizedAddress(); err != nil {
		return err
	}

	// Process the additional tags
	tags := map[string]string{
		"server": tagAddress,
		"db":     dbname.String(),
	}

	fields := make(map[string]interface{})
COLUMN:
	for col, val := range columnMap {
		log.Printf("D! postgresql_extensible: column: %s = %T: %s\n", col, *val, *val)
		_, ignore := ignoredColumns[col]
		if ignore || *val == nil {
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
				log.Println("failed to add additional tag", col)
			}
			continue COLUMN
		}

		if v, ok := (*val).([]byte); ok {
			fields[col] = string(v)
		} else {
			fields[col] = *val
		}
	}
	acc.AddFields(meas_name, fields, tags)
	return nil
}

func init() {
	inputs.Add("postgresql_extensible", func() telegraf.Input {
		return &Postgresql{}
	})
}
