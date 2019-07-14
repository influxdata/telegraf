package db

import (
	"log"

	"github.com/jackc/pgx"
	// pgx driver for sql connections
	_ "github.com/jackc/pgx/stdlib"
)

// Wrapper defines an interface that encapsulates communication with a DB.
type Wrapper interface {
	Exec(query string, args ...interface{}) (pgx.CommandTag, error)
	DoCopy(fullTableName *pgx.Identifier, colNames []string, batch [][]interface{}) error
	Query(query string, args ...interface{}) (*pgx.Rows, error)
	QueryRow(query string, args ...interface{}) *pgx.Row
	Close() error
}

type defaultDbWrapper struct {
	db *pgx.Conn
}

// NewWrapper returns an implementation of the db.Wrapper interface
// that issues queries to a PG database.
func NewWrapper(address string) (Wrapper, error) {
	connConfig, err := pgx.ParseConnectionString(address)
	if err != nil {
		log.Printf("E! Couldn't parse connection address: %s\n%v", address, err)
		return nil, err
	}
	db, err := pgx.Connect(connConfig)
	if err != nil {
		log.Printf("E! Couldn't connect to server\n%v", err)
		return nil, err
	}

	return &defaultDbWrapper{
		db: db,
	}, nil
}

func (d *defaultDbWrapper) Exec(query string, args ...interface{}) (pgx.CommandTag, error) {
	return d.db.Exec(query, args...)
}

func (d *defaultDbWrapper) DoCopy(fullTableName *pgx.Identifier, colNames []string, batch [][]interface{}) error {
	source := pgx.CopyFromRows(batch)
	_, err := d.db.CopyFrom(*fullTableName, colNames, source)
	if err != nil {
		log.Printf("E! Could not insert batch of rows in output db\n%v", err)
	}

	return err
}

func (d *defaultDbWrapper) Close() error { return d.db.Close() }

func (d *defaultDbWrapper) Query(query string, args ...interface{}) (*pgx.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *defaultDbWrapper) QueryRow(query string, args ...interface{}) *pgx.Row {
	return d.db.QueryRow(query, args...)
}
