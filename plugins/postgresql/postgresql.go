package postgresql

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	"github.com/influxdb/telegraf/plugins"

	_ "github.com/lib/pq"
)

type Server struct {
	Address        string
	Databases      []string
	OrderedColumns []string
}

type Postgresql struct {
	Servers []*Server
}

var ignoredColumns = map[string]bool{"datid": true, "datname": true, "stats_reset": true}

var sampleConfig = `
  # specify servers via an array of tables
  [[plugins.postgresql.servers]]

  # specify address via a url matching:
  #   postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]
  # or a simple string:
  #   host=localhost user=pqotest password=... sslmode=... dbname=app_production
  #
  # All connection parameters are optional. By default, the host is localhost
  # and the user is the currently running user. For localhost, we default
  # to sslmode=disable as well.
  #
  # Without the dbname parameter, the driver will default to a database
  # with the same name as the user. This dbname is just for instantiating a
  # connection with the server and doesn't restrict the databases we are trying
  # to grab metrics for.
  #

  address = "host=localhost user=postgres sslmode=disable"

  # A list of databases to pull metrics about. If not specified, metrics for all
  # databases are gathered.

  # databases = ["app_production", "blah_testing"]

  # [[plugins.postgresql.servers]]
  # address = "influx@remoteserver"
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

var localhost = &Server{Address: "sslmode=disable"}

func (p *Postgresql) Gather(acc plugins.Accumulator) error {
	if len(p.Servers) == 0 {
		p.gatherServer(localhost, acc)
		return nil
	}

	for _, serv := range p.Servers {
		err := p.gatherServer(serv, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Postgresql) gatherServer(serv *Server, acc plugins.Accumulator) error {
	var query string

	if serv.Address == "" || serv.Address == "localhost" {
		serv = localhost
	}

	db, err := sql.Open("postgres", serv.Address)
	if err != nil {
		return err
	}

	defer db.Close()

	if len(serv.Databases) == 0 {
		query = `SELECT * FROM pg_stat_database`
	} else {
		query = fmt.Sprintf(`SELECT * FROM pg_stat_database WHERE datname IN ('%s')`, strings.Join(serv.Databases, "','"))
	}

	rows, err := db.Query(query)
	if err != nil {
		return err
	}

	defer rows.Close()

	// grab the column information from the result
	serv.OrderedColumns, err = rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		err = p.accRow(rows, acc, serv)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (p *Postgresql) accRow(row scanner, acc plugins.Accumulator, serv *Server) error {
	var columnVars []interface{}
	var dbname bytes.Buffer

	// this is where we'll store the column name with its *interface{}
	columnMap := make(map[string]*interface{})

	for _, column := range serv.OrderedColumns {
		columnMap[column] = new(interface{})
	}

	// populate the array of interface{} with the pointers in the right order
	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[serv.OrderedColumns[i]])
	}

	// deconstruct array of variables and send to Scan
	err := row.Scan(columnVars...)

	if err != nil {
		return err
	}

	// extract the database name from the column map
	dbnameChars := (*columnMap["datname"]).([]uint8)
	for i := 0; i < len(dbnameChars); i++ {
		dbname.WriteString(string(dbnameChars[i]))
	}

	tags := map[string]string{"server": serv.Address, "db": dbname.String()}

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
	plugins.Add("postgresql", func() plugins.Plugin {
		return &Postgresql{}
	})
}
