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

var sampleConfig = `
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqgotest password=... sslmode=... dbname=app_production
  ##
  ## All connection parameters are optional.
  ##
  ## Without the dbname parameter, the driver will default to a database
  ## with the same name as the user. This dbname is just for instantiating a
  ## connection with the server and doesn't restrict the databases we are trying
  ## to grab metrics for.
  ##
  address = "host=localhost user=postgres sslmode=disable"
  ## A custom name for the database that will be used as the "server" tag in the
  ## measurement output. If not specified, a default one generated from
  ## the connection address is used.
  # outputaddress = "db01"

  ## connection configuration.
  ## maxlifetime - specify the maximum lifetime of a connection.
  ## default is forever (0s)
  max_lifetime = "0s"

  ## A  list of databases to explicitly ignore.  If not specified, metrics for all
  ## databases are gathered.  Do NOT use with the 'databases' option.
  # ignored_databases = ["postgres", "template0", "template1"]

  ## A list of databases to pull metrics about. If not specified, metrics for all
  ## databases are gathered.  Do NOT use with the 'ignored_databases' option.
  # databases = ["app_production", "testing"]

  ## Whether to use prepared statements when connecting to the database.
  ## This should be set to false when connecting through a PgBouncer instance
  ## with pool_mode set to transaction.
  # prepared_statements = true
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
    
	// total_table_count
	query = `SELECT count(*) as total_table_count from information_schema.tables where table_schema not in ('sc_toolkit','information_schema','pg_catalog');`

	totalTableCountRow, err := p.DB.Query(query)
	if err != nil {
		return err
	}

	defer totalTableCountRow.Close()
	if columns, err = totalTableCountRow.Columns(); err != nil {
		return err
	}

	for totalTableCountRow.Next() {
		err = p.accRow(totalTableCountRow, acc, columns)
		if err != nil {
			return err
		}
	}

	// total_connect
	query = `select count(*) total_connect, 
	count(*) filter(where state='idle') idle_connect, 
	count(*) filter(where state<>'idle') active_connect,
	count(*) filter(where state='active') running_connect,
	count(*) filter(where state like 'wait%') waiting_connect
	from pg_stat_activity where pid <> pg_backend_pid();`
	totalConnectRow, err := p.DB.Query(query)
	if err != nil {
		return err
	}

	defer totalConnectRow.Close()
	if columns, err = totalConnectRow.Columns(); err != nil {
		return err
	}

	for totalConnectRow.Next() {
		err = p.accRow(totalConnectRow, acc, columns)
		if err != nil {
			return err
		}
	}
  
	// connect_by_name sql
	query = `select usename, 
	count(*) total, 
	count(*) filter(where query='<IDLE>') idle, 
	count(*) filter(where query<>'<IDLE>') active 
	from pg_stat_activity group by 1;`
	connectByNameRow, err := p.DB.Query(query)
	if err != nil {
		return err
	}

	defer connectByNameRow.Close()
	if columns, err = connectByNameRow.Columns(); err != nil {
		return err
	}

	for connectByNameRow.Next() {
		err = p.accRow(connectByNameRow, acc, columns)
		if err != nil {
			return err
		}
	}

	// connect_by_client sql
	query = `select client_addr,
             count(*) total,
             count(*) filter(where query='<IDLE>') idle,
             count(*) filter(where query<>'<IDLE>') active
             from pg_stat_activity where pid <> pg_backend_pid() group by 1;`
	connectByClientRow, err := p.DB.Query(query)
	if err != nil {
		return err
	}

	defer connectByClientRow.Close()
	if columns, err = connectByClientRow.Columns(); err != nil {
		return err
	}

	for connectByClientRow.Next() {
		err = p.accRow(connectByClientRow, acc, columns)
		if err != nil {
			return err
		}
	}

	// up_day sql
	query = `select extract(day FROM(age(now()::date, pg_start_time()::date))) as up_day`;
	upDaytRow, err := p.DB.Query(query)
	if err != nil {
	return err
	}

	defer upDaytRow.Close()
	if columns, err = upDaytRow.Columns(); err != nil {
	return err
	}

	for upDaytRow.Next() {
	err = p.accRow(upDaytRow, acc, columns)
	if err != nil {
		return err
	}
	}
	return upDaytRow.Err()
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
