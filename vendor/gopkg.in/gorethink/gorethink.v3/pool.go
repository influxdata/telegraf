package gorethink

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"golang.org/x/net/context"
	"gopkg.in/fatih/pool.v2"
)

var (
	errPoolClosed = errors.New("gorethink: pool is closed")
)

// A Pool is used to store a pool of connections to a single RethinkDB server
type Pool struct {
	host Host
	opts *ConnectOpts

	pool pool.Pool

	mu     sync.RWMutex // protects following fields
	closed bool
}

// NewPool creates a new connection pool for the given host
func NewPool(host Host, opts *ConnectOpts) (*Pool, error) {
	initialCap := opts.InitialCap
	if initialCap <= 0 {
		// Fallback to MaxIdle if InitialCap is zero, this should be removed
		// when MaxIdle is removed
		initialCap = opts.MaxIdle
	}

	maxOpen := opts.MaxOpen
	if maxOpen <= 0 {
		maxOpen = 2
	}

	p, err := pool.NewChannelPool(initialCap, maxOpen, func() (net.Conn, error) {
		conn, err := NewConnection(host.String(), opts)
		if err != nil {
			return nil, err
		}

		return conn, err
	})
	if err != nil {
		return nil, err
	}

	return &Pool{
		pool: p,
		host: host,
		opts: opts,
	}, nil
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (p *Pool) Ping() error {
	_, pc, err := p.conn()
	if err != nil {
		return err
	}
	return pc.Close()
}

// Close closes the database, releasing any open resources.
//
// It is rare to Close a Pool, as the Pool handle is meant to be
// long-lived and shared between many goroutines.
func (p *Pool) Close() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return nil
	}

	p.pool.Close()

	return nil
}

func (p *Pool) conn() (*Connection, *pool.PoolConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, nil, errPoolClosed
	}

	nc, err := p.pool.Get()
	if err != nil {
		return nil, nil, err
	}

	pc, ok := nc.(*pool.PoolConn)
	if !ok {
		// This should never happen!
		return nil, nil, fmt.Errorf("Invalid connection in pool")
	}

	conn, ok := pc.Conn.(*Connection)
	if !ok {
		// This should never happen!
		return nil, nil, fmt.Errorf("Invalid connection in pool")
	}

	return conn, pc, nil
}

// SetInitialPoolCap sets the initial capacity of the connection pool.
//
// Deprecated: This value should only be set when connecting
func (p *Pool) SetInitialPoolCap(n int) {
	return
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool.
//
// Deprecated: This value should only be set when connecting
func (p *Pool) SetMaxIdleConns(n int) {
	return
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
//
// Deprecated: This value should only be set when connecting
func (p *Pool) SetMaxOpenConns(n int) {
	return
}

// Query execution functions

// Exec executes a query without waiting for any response.
func (p *Pool) Exec(ctx context.Context, q Query) error {
	c, pc, err := p.conn()
	if err != nil {
		return err
	}
	defer pc.Close()

	_, _, err = c.Query(ctx, q)

	if c.isBad() {
		pc.MarkUnusable()
	}

	return err
}

// Query executes a query and waits for the response
func (p *Pool) Query(ctx context.Context, q Query) (*Cursor, error) {
	c, pc, err := p.conn()
	if err != nil {
		return nil, err
	}

	_, cursor, err := c.Query(ctx, q)

	if err == nil {
		cursor.releaseConn = releaseConn(c, pc)
	} else if c.isBad() {
		pc.MarkUnusable()
	}

	return cursor, err
}

// Server returns the server name and server UUID being used by a connection.
func (p *Pool) Server() (ServerResponse, error) {
	var response ServerResponse

	c, pc, err := p.conn()
	if err != nil {
		return response, err
	}
	defer pc.Close()

	response, err = c.Server()

	if c.isBad() {
		pc.MarkUnusable()
	}

	return response, err
}

func releaseConn(c *Connection, pc *pool.PoolConn) func() error {
	return func() error {
		if c.isBad() {
			pc.MarkUnusable()
		}

		return pc.Close()
	}
}
