// +build go1.10

package stdlib

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/jackc/pgx"
)

// OptionOpenDB options for configuring the driver when opening a new db pool.
type OptionOpenDB func(*connector)

// OptionAfterConnect provide a callback for after connect.
func OptionAfterConnect(ac func(*pgx.Conn) error) OptionOpenDB {
	return func(dc *connector) {
		dc.AfterConnect = ac
	}
}

func OpenDB(config pgx.ConnConfig, opts ...OptionOpenDB) *sql.DB {
	c := connector{
		ConnConfig:   config,
		AfterConnect: func(*pgx.Conn) error { return nil }, // noop after connect by default
		driver:       pgxDriver,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return sql.OpenDB(c)
}

type connector struct {
	pgx.ConnConfig
	AfterConnect func(*pgx.Conn) error // function to call on every new connection
	driver       *Driver
}

// Connect implement driver.Connector interface
func (c connector) Connect(ctx context.Context) (driver.Conn, error) {
	var (
		err  error
		conn *pgx.Conn
	)

	if conn, err = pgx.Connect(c.ConnConfig); err != nil {
		return nil, err
	}

	if err = c.AfterConnect(conn); err != nil {
		return nil, err
	}

	return &Conn{conn: conn, driver: c.driver, connConfig: c.ConnConfig}, nil
}

// Driver implement driver.Connector interface
func (c connector) Driver() driver.Driver {
	return c.driver
}
