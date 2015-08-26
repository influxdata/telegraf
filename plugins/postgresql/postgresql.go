package postgresql

import (
	"database/sql"

	"github.com/influxdb/telegraf/plugins"

	_ "github.com/lib/pq"
)

type Server struct {
	Address   string
	Databases []string
}

type Postgresql struct {
	Servers []*Server
}

var sampleConfig = `
	# specify servers via an array of tables
	[[postgresql.servers]]

	# specify address via a url matching:
	#   postgres://[pqgotest[:password]]@localhost?sslmode=[disable|verify-ca|verify-full]
	# or a simple string:
	#   host=localhost user=pqotest password=... sslmode=...
	#
	# All connection parameters are optional. By default, the host is localhost
	# and the user is the currently running user. For localhost, we default
	# to sslmode=disable as well.
	#

	address = "sslmode=disable"

	# A list of databases to pull metrics about. If not specified, metrics for all
	# databases are gathered.

	# databases = ["app_production", "blah_testing"]

	# [[postgresql.servers]]
	# address = "influx@remoteserver"
`

func (p *Postgresql) SampleConfig() string {
	return sampleConfig
}

func (p *Postgresql) Description() string {
	return "Read metrics from one or many postgresql servers"
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
	if serv.Address == "" || serv.Address == "localhost" {
		serv = localhost
	}

	db, err := sql.Open("postgres", serv.Address)
	if err != nil {
		return err
	}

	defer db.Close()

	if len(serv.Databases) == 0 {
		rows, err := db.Query(`SELECT * FROM pg_stat_database`)
		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			err := p.accRow(rows, acc, serv.Address)
			if err != nil {
				return err
			}
		}

		return rows.Err()
	} else {
		for _, name := range serv.Databases {
			row := db.QueryRow(`SELECT * FROM pg_stat_database WHERE datname=$1`, name)

			err := p.accRow(row, acc, serv.Address)
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

func (p *Postgresql) accRow(row scanner, acc plugins.Accumulator, server string) error {
	var ignore interface{}
	var name string
	var commit, rollback, read, hit int64
	var returned, fetched, inserted, updated, deleted int64
	var conflicts, temp_files, temp_bytes, deadlocks int64
	var read_time, write_time float64

	err := row.Scan(&ignore, &name, &ignore,
		&commit, &rollback,
		&read, &hit,
		&returned, &fetched, &inserted, &updated, &deleted,
		&conflicts, &temp_files, &temp_bytes,
		&deadlocks, &read_time, &write_time,
		&ignore,
	)

	if err != nil {
		return err
	}

	tags := map[string]string{"server": server, "db": name}

	acc.Add("xact_commit", commit, tags)
	acc.Add("xact_rollback", rollback, tags)
	acc.Add("blks_read", read, tags)
	acc.Add("blks_hit", hit, tags)
	acc.Add("tup_returned", returned, tags)
	acc.Add("tup_fetched", fetched, tags)
	acc.Add("tup_inserted", inserted, tags)
	acc.Add("tup_updated", updated, tags)
	acc.Add("tup_deleted", deleted, tags)
	acc.Add("conflicts", conflicts, tags)
	acc.Add("temp_files", temp_files, tags)
	acc.Add("temp_bytes", temp_bytes, tags)
	acc.Add("deadlocks", deadlocks, tags)
	acc.Add("blk_read_time", read_time, tags)
	acc.Add("blk_write_time", read_time, tags)

	return nil
}

func init() {
	plugins.Add("postgresql", func() plugins.Plugin {
		return &Postgresql{}
	})
}
