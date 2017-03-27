package pgbouncer

import (
	"bytes"
	"database/sql"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/lib/pq"
)

type Pgbouncer struct {
	Address          string
	Databases        []string
	OrderedColumns   []string
	AllColumns       []string
	sanitizedAddress string
}

var ignoredColumns = map[string]bool{"pool_mode": true, "database": true, "user": true}

var sampleConfig = `
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost:port[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqotest port=6432 password=... sslmode=... dbname=pgbouncer
  ##
  ## All connection parameters are optional, except for dbname,
  ## you need to set it always as pgbouncer.
  address = "host=localhost user=postgres port=6432 sslmode=disable dbname=pgbouncer"

  ## A list of databases to pull metrics about. If not specified, metrics for all
  ## databases are gathered.
  # databases = ["app_production", "testing"]
`

func (p *Pgbouncer) SampleConfig() string {
	return sampleConfig
}

func (p *Pgbouncer) Description() string {
	return "Read metrics from one or many pgbouncer servers"
}

func (p *Pgbouncer) IgnoredColumns() map[string]bool {
	return ignoredColumns
}

var localhost = "host=localhost port=6432 sslmode=disable dbname=pgbouncer"

func (p *Pgbouncer) Gather(acc telegraf.Accumulator) error {
	if p.Address == "" || p.Address == "localhost" {
		p.Address = localhost
	}

	db, err := sql.Open("postgres", p.Address)
	if err != nil {
		return err
	}

	defer db.Close()

	queries := map[string]string{"pools": "SHOW POOLS", "stats": "SHOW STATS"}

	for metric, query := range queries {
		rows, err := db.Query(query)
		if err != nil {
			return err
		}

		defer rows.Close()

		// grab the column information from the result
		p.OrderedColumns, err = rows.Columns()
		if err != nil {
			return err
		} else {
			p.AllColumns = make([]string, len(p.OrderedColumns))
			copy(p.AllColumns, p.OrderedColumns)
		}

		for rows.Next() {
			err = p.accRow(rows, metric, acc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

var passwordKVMatcher, _ = regexp.Compile("password=\\S+ ?")

func (p *Pgbouncer) SanitizedAddress() (_ string, err error) {
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

func (p *Pgbouncer) accRow(row scanner, metric string, acc telegraf.Accumulator) error {
	var columnVars []interface{}
	var tags = make(map[string]string)
	var dbname, user, poolMode bytes.Buffer

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

	// extract the database name from the column map
	dbnameChars := (*columnMap["database"]).([]uint8)
	for i := 0; i < len(dbnameChars); i++ {
		dbname.WriteString(string(dbnameChars[i]))
	}

	if p.ignoreDatabase(dbname.String()) {
		return nil
	}

	tags["db"] = dbname.String()

	if columnMap["user"] != nil {
		userChars := (*columnMap["user"]).([]uint8)
		for i := 0; i < len(userChars); i++ {
			user.WriteString(string(userChars[i]))
		}
		tags["user"] = user.String()
	}

	if columnMap["pool_mode"] != nil {
		poolChars := (*columnMap["pool_mode"]).([]uint8)
		for i := 0; i < len(poolChars); i++ {
			poolMode.WriteString(string(poolChars[i]))
		}
		tags["pool_mode"] = poolMode.String()
	}

	var tagAddress string
	tagAddress, err = p.SanitizedAddress()
	if err != nil {
		return err
	} else {
		tags["server"] = tagAddress
	}

	fields := make(map[string]interface{})
	for col, val := range columnMap {
		_, ignore := ignoredColumns[col]
		if !ignore {
			fields[col] = *val
		}
	}
	acc.AddFields("pgbouncer_"+metric, fields, tags)

	return nil
}

func (p *Pgbouncer) ignoreDatabase(db string) bool {
	if len(p.Databases) == 0 {
		return false
	}

	for _, dbName := range p.Databases {
		if db == dbName {
			return false
		}
	}
	return true
}

func init() {
	inputs.Add("pgbouncer", func() telegraf.Input {
		return &Pgbouncer{}
	})
}
