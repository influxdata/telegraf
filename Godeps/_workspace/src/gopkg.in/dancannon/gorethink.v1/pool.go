package gorethink

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
)

const defaultMaxIdleConns = 1

// maxBadConnRetries is the number of maximum retries if the driver returns
// driver.ErrBadConn to signal a broken connection.
const maxBadConnRetries = 10

var (
	connectionRequestQueueSize = 1000000

	errPoolClosed = errors.New("gorethink: pool is closed")
	errConnClosed = errors.New("gorethink: conn is closed")
	errConnBusy   = errors.New("gorethink: conn is busy")
)

// depSet is a finalCloser's outstanding dependencies
type depSet map[interface{}]bool // set of true bools
// The finalCloser interface is used by (*Pool).addDep and related
// dependency reference counting.
type finalCloser interface {
	// finalClose is called when the reference count of an object
	// goes to zero. (*Pool).mu is not held while calling it.
	finalClose() error
}

// A Pool is used to store a pool of connections to a single RethinkDB server
type Pool struct {
	host Host
	opts *ConnectOpts

	mu           sync.Mutex // protects following fields
	freeConn     []*poolConn
	connRequests []chan connRequest
	numOpen      int
	pendingOpens int
	// Used to signal the need for new connections
	// a goroutine running connectionOpener() reads on this chan and
	// maybeOpenNewConnections sends on the chan (one send per needed connection)
	// It is closed during p.Close(). The close tells the connectionOpener
	// goroutine to exit.
	openerCh chan struct{}
	closed   bool
	dep      map[finalCloser]depSet
	lastPut  map[*poolConn]string // stacktrace of last conn's put; debug only
	maxIdle  int                  // zero means defaultMaxIdleConns; negative means 0
	maxOpen  int                  // <= 0 means unlimited
}

// NewPool creates a new connection pool for the given host
func NewPool(host Host, opts *ConnectOpts) (*Pool, error) {
	p := &Pool{
		host:     host,
		opts:     opts,
		openerCh: make(chan struct{}, connectionRequestQueueSize),
		lastPut:  make(map[*poolConn]string),
	}

	p.SetMaxIdleConns(opts.MaxIdle)
	p.SetMaxOpenConns(opts.MaxOpen)

	go p.connectionOpener()
	return p, nil
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (p *Pool) Ping() error {
	pc, err := p.conn()
	if err != nil {
		return err
	}
	p.putConn(pc, nil)
	return nil
}

// Close closes the database, releasing any open resources.
//
// It is rare to Close a Pool, as the Pool handle is meant to be
// long-lived and shared between many goroutines.
func (p *Pool) Close() error {
	p.mu.Lock()
	if p.closed { // Make Pool.Close idempotent
		p.mu.Unlock()
		return nil
	}
	close(p.openerCh)
	var err error
	fns := make([]func() error, 0, len(p.freeConn))
	for _, pc := range p.freeConn {
		fns = append(fns, pc.closePoolLocked())
	}
	p.freeConn = nil
	p.closed = true
	for _, req := range p.connRequests {
		close(req)
	}
	p.mu.Unlock()
	for _, fn := range fns {
		err1 := fn()
		if err1 != nil {
			err = err1
		}
	}
	return err
}

func (p *Pool) maxIdleConnsLocked() int {
	n := p.maxIdle
	switch {
	case n == 0:
		return defaultMaxIdleConns
	case n < 0:
		return 0
	default:
		return n
	}
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool.
//
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns
// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit
//
// If n <= 0, no idle connections are retained.
func (p *Pool) SetMaxIdleConns(n int) {
	p.mu.Lock()
	if n > 0 {
		p.maxIdle = n
	} else {
		// No idle connections.
		p.maxIdle = -1
	}
	// Make sure maxIdle doesn't exceed maxOpen
	if p.maxOpen > 0 && p.maxIdleConnsLocked() > p.maxOpen {
		p.maxIdle = p.maxOpen
	}
	var closing []*poolConn
	idleCount := len(p.freeConn)
	maxIdle := p.maxIdleConnsLocked()
	if idleCount > maxIdle {
		closing = p.freeConn[maxIdle:]
		p.freeConn = p.freeConn[:maxIdle]
	}
	p.mu.Unlock()
	for _, c := range closing {
		c.Close()
	}
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
//
// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
// MaxIdleConns, then MaxIdleConns will be reduced to match the new
// MaxOpenConns limit
//
// If n <= 0, then there is no limit on the number of open connections.
// The default is 0 (unlimited).
func (p *Pool) SetMaxOpenConns(n int) {
	p.mu.Lock()
	p.maxOpen = n
	if n < 0 {
		p.maxOpen = 0
	}
	syncMaxIdle := p.maxOpen > 0 && p.maxIdleConnsLocked() > p.maxOpen
	p.mu.Unlock()
	if syncMaxIdle {
		p.SetMaxIdleConns(n)
	}
}

// Assumes p.mu is locked.
// If there are connRequests and the connection limit hasn't been reached,
// then tell the connectionOpener to open new connections.
func (p *Pool) maybeOpenNewConnections() {
	numRequests := len(p.connRequests) - p.pendingOpens
	if p.maxOpen > 0 {
		numCanOpen := p.maxOpen - (p.numOpen + p.pendingOpens)
		if numRequests > numCanOpen {
			numRequests = numCanOpen
		}
	}
	for numRequests > 0 {
		p.pendingOpens++
		numRequests--
		p.openerCh <- struct{}{}
	}
}

// Runs in a separate goroutine, opens new connections when requested.
func (p *Pool) connectionOpener() {
	for _ = range p.openerCh {
		p.openNewConnection()
	}
}

// Open one new connection
func (p *Pool) openNewConnection() {
	ci, err := NewConnection(p.host.String(), p.opts)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		if err == nil {
			ci.Close()
		}
		return
	}
	p.pendingOpens--
	if err != nil {
		p.putConnPoolLocked(nil, err)
		return
	}
	pc := &poolConn{
		p:  p,
		ci: ci,
	}
	if p.putConnPoolLocked(pc, err) {
		p.addDepLocked(pc, pc)
		p.numOpen++
	} else {
		ci.Close()
	}
}

// connRequest represents one request for a new connection
// When there are no idle connections available, Pool.conn will create
// a new connRequest and put it on the p.connRequests list.
type connRequest struct {
	conn *poolConn
	err  error
}

// conn returns a newly-opened or cached *poolConn
func (p *Pool) conn() (*poolConn, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errPoolClosed
	}
	// If p.maxOpen > 0 and the number of open connections is over the limit
	// and there are no free connection, make a request and wait.
	if p.maxOpen > 0 && p.numOpen >= p.maxOpen && len(p.freeConn) == 0 {
		// Make the connRequest channel. It's buffered so that the
		// connectionOpener doesn't block while waiting for the req to be read.
		req := make(chan connRequest, 1)
		p.connRequests = append(p.connRequests, req)
		p.mu.Unlock()
		ret := <-req
		return ret.conn, ret.err
	}
	if c := len(p.freeConn); c > 0 {
		conn := p.freeConn[0]
		copy(p.freeConn, p.freeConn[1:])
		p.freeConn = p.freeConn[:c-1]
		conn.inUse = true
		p.mu.Unlock()
		return conn, nil
	}
	p.numOpen++ // optimistically
	p.mu.Unlock()
	ci, err := NewConnection(p.host.String(), p.opts)
	if err != nil {
		p.mu.Lock()
		p.numOpen-- // correct for earlier optimism
		p.mu.Unlock()
		return nil, err
	}
	p.mu.Lock()
	pc := &poolConn{
		p:  p,
		ci: ci,
	}
	p.addDepLocked(pc, pc)
	pc.inUse = true
	p.mu.Unlock()
	return pc, nil
}

// connIfFree returns (wanted, nil) if wanted is still a valid conn and
// isn't in use.
//
// The error is errConnClosed if the connection if the requested connection
// is invalid because it's been closed.
//
// The error is errConnBusy if the connection is in use.
func (p *Pool) connIfFree(wanted *poolConn) (*poolConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if wanted.pmuClosed {
		return nil, errConnClosed
	}
	if wanted.inUse {
		return nil, errConnBusy
	}
	idx := -1
	for ii, v := range p.freeConn {
		if v == wanted {
			idx = ii
			break
		}
	}
	if idx >= 0 {
		p.freeConn = append(p.freeConn[:idx], p.freeConn[idx+1:]...)
		wanted.inUse = true
		return wanted, nil
	}

	panic("connIfFree call requested a non-closed, non-busy, non-free conn")
}

// noteUnusedCursor notes that si is no longer used and should
// be closed whenever possible (when c is next not in use), unless c is
// already closed.
func (p *Pool) noteUnusedCursor(c *poolConn, ci *Cursor) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if c.inUse {
		c.onPut = append(c.onPut, func() {
			ci.Close()
		})
	} else {
		c.Lock()
		defer c.Unlock()
		if !c.finalClosed {
			ci.Close()
		}
	}
}

// debugGetPut determines whether getConn & putConn calls' stack traces
// are returned for more verbose crashes.
const debugGetPut = false

// putConn adds a connection to the free pool.
// err is optionally the last error that occurred on this connection.
func (p *Pool) putConn(pc *poolConn, err error) {
	p.mu.Lock()
	if !pc.inUse {
		if debugGetPut {
			fmt.Printf("putConn(%v) DUPLICATE was: %s\n\nPREVIOUS was: %s", pc, stack(), p.lastPut[pc])
		}
		panic("gorethink: connection returned that was never out")
	}
	if debugGetPut {
		p.lastPut[pc] = stack()
	}
	pc.inUse = false
	for _, fn := range pc.onPut {
		fn()
	}
	pc.onPut = nil
	if err != nil && pc.ci.bad {
		// Don't reuse bad connections.
		// Since the conn is considered bad and is being discarded, treat it
		// as closed. Don't decrement the open count here, finalClose will
		// take care of that.
		p.maybeOpenNewConnections()
		p.mu.Unlock()
		pc.Close()
		return
	}
	added := p.putConnPoolLocked(pc, nil)
	p.mu.Unlock()
	if !added {
		pc.Close()
	}
}

// Satisfy a connRequest or put the poolConn in the idle pool and return true
// or return false.
// putConnPoolLocked will satisfy a connRequest if there is one, or it will
// return the *poolConn to the freeConn list if err == nil and the idle
// connection limit will not be exceeded.
// If err != nil, the value of pc is ignored.
// If err == nil, then pc must not equal nil.
// If a connRequest was fulfilled or the *poolConn was placed in the
// freeConn list, then true is returned, otherwise false is returned.
func (p *Pool) putConnPoolLocked(pc *poolConn, err error) bool {
	if p.maxOpen > 0 && p.numOpen > p.maxOpen {
		return false
	}
	if c := len(p.connRequests); c > 0 {
		req := p.connRequests[0]
		// This copy is O(n) but in practice faster than a linked list.
		// TODO: consider compacting it down less often and
		// moving the base instead?
		copy(p.connRequests, p.connRequests[1:])
		p.connRequests = p.connRequests[:c-1]
		if err == nil {
			pc.inUse = true
		}
		req <- connRequest{
			conn: pc,
			err:  err,
		}
		return true
	} else if err == nil && !p.closed && p.maxIdleConnsLocked() > len(p.freeConn) {
		p.freeConn = append(p.freeConn, pc)
		return true
	}
	return false
}

// addDep notes that x now depends on dep, and x's finalClose won't be
// called until all of x's dependencies are removed with removeDep.
func (p *Pool) addDep(x finalCloser, dep interface{}) {
	//println(fmt.Sprintf("addDep(%T %p, %T %p)", x, x, dep, dep))
	p.mu.Lock()
	defer p.mu.Unlock()
	p.addDepLocked(x, dep)
}

func (p *Pool) addDepLocked(x finalCloser, dep interface{}) {
	if p.dep == nil {
		p.dep = make(map[finalCloser]depSet)
	}
	xdep := p.dep[x]
	if xdep == nil {
		xdep = make(depSet)
		p.dep[x] = xdep
	}
	xdep[dep] = true
}

// removeDep notes that x no longer depends on dep.
// If x still has dependencies, nil is returned.
// If x no longer has any dependencies, its finalClose method will be
// called and its error value will be returned.
func (p *Pool) removeDep(x finalCloser, dep interface{}) error {
	p.mu.Lock()
	fn := p.removeDepLocked(x, dep)
	p.mu.Unlock()
	return fn()
}

func (p *Pool) removeDepLocked(x finalCloser, dep interface{}) func() error {
	//println(fmt.Sprintf("removeDep(%T %p, %T %p)", x, x, dep, dep))
	xdep, ok := p.dep[x]
	if !ok {
		panic(fmt.Sprintf("unpaired removeDep: no deps for %T", x))
	}
	l0 := len(xdep)
	delete(xdep, dep)
	switch len(xdep) {
	case l0:
		// Nothing removed. Shouldn't happen.
		panic(fmt.Sprintf("unpaired removeDep: no %T dep on %T", dep, x))
	case 0:
		// No more dependencies.
		delete(p.dep, x)
		return x.finalClose
	default:
		// Dependencies remain.
		return func() error { return nil }
	}
}

// Query execution functions

// Exec executes a query without waiting for any response.
func (p *Pool) Exec(q Query) error {
	var err error
	for i := 0; i < maxBadConnRetries; i++ {
		err = p.exec(q)
		if err != ErrBadConn {
			break
		}
	}
	return err
}
func (p *Pool) exec(q Query) (err error) {
	pc, err := p.conn()
	if err != nil {
		return err
	}
	defer func() {
		p.putConn(pc, err)
	}()

	pc.Lock()
	_, _, err = pc.ci.Query(q)
	pc.Unlock()

	if err != nil {
		return err
	}
	return nil
}

// Query executes a query and waits for the response
func (p *Pool) Query(q Query) (*Cursor, error) {
	var cursor *Cursor
	var err error
	for i := 0; i < maxBadConnRetries; i++ {
		cursor, err = p.query(q)
		if err != ErrBadConn {
			break
		}
	}
	return cursor, err
}
func (p *Pool) query(query Query) (*Cursor, error) {
	ci, err := p.conn()
	if err != nil {
		return nil, err
	}
	return p.queryConn(ci, ci.releaseConn, query)
}

// queryConn executes a query on the given connection.
// The connection gets released by the releaseConn function.
func (p *Pool) queryConn(pc *poolConn, releaseConn func(error), q Query) (*Cursor, error) {
	pc.Lock()
	_, cursor, err := pc.ci.Query(q)
	pc.Unlock()
	if err != nil {
		releaseConn(err)
		return nil, err
	}

	cursor.releaseConn = releaseConn

	return cursor, nil
}

// Helper functions

func stack() string {
	var buf [2 << 10]byte
	return string(buf[:runtime.Stack(buf[:], false)])
}
