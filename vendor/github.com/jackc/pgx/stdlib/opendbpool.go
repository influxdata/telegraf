// +build go1.10

package stdlib

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/jackc/pgx"
)

// OptionOpenDB options for configuring the driver when opening a new db pool.
type OptionOpenDBFromPool func(*poolConnector)

// OptionAfterConnect provide a callback for after connect.
func OptionPreferSimpleProtocol(preferSimpleProtocol bool) OptionOpenDBFromPool {
	return func(dc *poolConnector) {
		dc.preferSimpleProtocol = preferSimpleProtocol
	}
}

// OpenDBFromPool create a sql.DB connection from a pgx.ConnPool
func OpenDBFromPool(pool *pgx.ConnPool, opts ...OptionOpenDBFromPool) *sql.DB {
	c := poolConnector{
		pool:   pool,
		driver: pgxDriver,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return sql.OpenDB(c)
}

type poolConnector struct {
	pool                 *pgx.ConnPool
	driver               *Driver
	preferSimpleProtocol bool
}

// Connect implement driver.Connector interface
func (pc poolConnector) Connect(ctx context.Context) (driver.Conn, error) {
	var (
		err  error
		conn *pgx.Conn
	)

	if conn, err = pc.pool.Acquire(); err != nil {
		return nil, err
	}

	return &Conn{conn: conn, driver: pc.driver, connConfig: pgx.ConnConfig{PreferSimpleProtocol: pc.preferSimpleProtocol}}, nil
}

// Driver implement driver.Connector interface
func (pc poolConnector) Driver() driver.Driver {
	return pc.driver
}
