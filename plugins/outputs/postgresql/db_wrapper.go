package postgresql

import (
	"database/sql"
	// pgx driver for sql connections
	_ "github.com/jackc/pgx/stdlib"
)

type dbWrapper interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Close() error
}

type defaultDbWrapper struct {
	db *sql.DB
}

func newDbWrapper(address string) (dbWrapper, error) {
	db, err := sql.Open("pgx", address)
	if err != nil {
		return nil, err
	}

	return &defaultDbWrapper{
		db: db,
	}, nil
}

func (d *defaultDbWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *defaultDbWrapper) Close() error { return d.db.Close() }

func (d *defaultDbWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *defaultDbWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}
