package postgresql

import (
	"database/sql"

	"github.com/influxdb/tivan/plugins"

	_ "github.com/lib/pq"
)

type Server struct {
	Address   string
	Databases []string
}

type Postgresql struct {
	Servers []*Server
}

func (p *Postgresql) Gather(acc plugins.Accumulator) error {
	for _, serv := range p.Servers {
		err := p.gatherServer(serv, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Postgresql) gatherServer(serv *Server, acc plugins.Accumulator) error {
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
			err := p.accRow(rows, acc)
			if err != nil {
				return err
			}
		}

		return rows.Err()
	} else {
		for _, name := range serv.Databases {
			row := db.QueryRow(`SELECT * FROM pg_stat_database WHERE datname=$1`, name)

			err := p.accRow(row, acc)
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

func (p *Postgresql) accRow(row scanner, acc plugins.Accumulator) error {
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

	tags := map[string]string{"db": name}

	acc.Add("postgresql_xact_commit", commit, tags)
	acc.Add("postgresql_xact_rollback", rollback, tags)
	acc.Add("postgresql_blks_read", read, tags)
	acc.Add("postgresql_blks_hit", hit, tags)
	acc.Add("postgresql_tup_returned", returned, tags)
	acc.Add("postgresql_tup_fetched", fetched, tags)
	acc.Add("postgresql_tup_inserted", inserted, tags)
	acc.Add("postgresql_tup_updated", updated, tags)
	acc.Add("postgresql_tup_deleted", deleted, tags)
	acc.Add("postgresql_conflicts", conflicts, tags)
	acc.Add("postgresql_temp_files", temp_files, tags)
	acc.Add("postgresql_temp_bytes", temp_bytes, tags)
	acc.Add("postgresql_deadlocks", deadlocks, tags)
	acc.Add("postgresql_blk_read_time", read_time, tags)
	acc.Add("postgresql_blk_write_time", read_time, tags)

	return nil
}

func init() {
	plugins.Add("postgresql", func() plugins.Plugin {
		return &Postgresql{}
	})
}
