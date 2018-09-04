package gorethink

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

const (
	respHeaderLen          = 12
	defaultKeepAlivePeriod = time.Second * 30
)

// Response represents the raw response from a query, most of the time you
// should instead use a Cursor when reading from the database.
type Response struct {
	Token     int64
	Type      p.Response_ResponseType   `json:"t"`
	ErrorType p.Response_ErrorType      `json:"e"`
	Notes     []p.Response_ResponseNote `json:"n"`
	Responses []json.RawMessage         `json:"r"`
	Backtrace []interface{}             `json:"b"`
	Profile   interface{}               `json:"p"`
}

// Connection is a connection to a rethinkdb database. Connection is not thread
// safe and should only be accessed be a single goroutine
type Connection struct {
	net.Conn

	address string
	opts    *ConnectOpts

	_       [4]byte
	mu      sync.Mutex
	token   int64
	cursors map[int64]*Cursor
	bad     bool
	closed  bool
}

// NewConnection creates a new connection to the database server
func NewConnection(address string, opts *ConnectOpts) (*Connection, error) {
	var err error
	c := &Connection{
		address: address,
		opts:    opts,
		cursors: make(map[int64]*Cursor),
	}

	keepAlivePeriod := defaultKeepAlivePeriod
	if opts.KeepAlivePeriod > 0 {
		keepAlivePeriod = opts.KeepAlivePeriod
	}

	// Connect to Server
	nd := net.Dialer{Timeout: c.opts.Timeout, KeepAlive: keepAlivePeriod}
	if c.opts.TLSConfig == nil {
		c.Conn, err = nd.Dial("tcp", address)
	} else {
		c.Conn, err = tls.DialWithDialer(&nd, "tcp", address, c.opts.TLSConfig)
	}
	if err != nil {
		return nil, RQLConnectionError{rqlError(err.Error())}
	}

	// Send handshake
	handshake, err := c.handshake(opts.HandshakeVersion)
	if err != nil {
		return nil, err
	}

	if err = handshake.Send(); err != nil {
		return nil, err
	}

	return c, nil
}

// Close closes the underlying net.Conn
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error

	if !c.closed {
		err = c.Conn.Close()
		c.closed = true
		c.cursors = make(map[int64]*Cursor)
	}

	return err
}

// Query sends a Query to the database, returning both the raw Response and a
// Cursor which should be used to view the query's response.
//
// This function is used internally by Run which should be used for most queries.
func (c *Connection) Query(ctx context.Context, q Query) (*Response, *Cursor, error) {
	if ctx == nil {
		ctx = c.contextFromConnectionOpts()
	}

	if c == nil {
		return nil, nil, ErrConnectionClosed
	}
	c.mu.Lock()
	if c.Conn == nil {
		c.bad = true
		c.mu.Unlock()
		return nil, nil, ErrConnectionClosed
	}

	// Add token if query is a START/NOREPLY_WAIT
	if q.Type == p.Query_START || q.Type == p.Query_NOREPLY_WAIT || q.Type == p.Query_SERVER_INFO {
		q.Token = c.nextToken()
	}
	if q.Type == p.Query_START || q.Type == p.Query_NOREPLY_WAIT {
		if c.opts.Database != "" {
			var err error
			q.Opts["db"], err = DB(c.opts.Database).Build()
			if err != nil {
				c.mu.Unlock()
				return nil, nil, RQLDriverError{rqlError(err.Error())}
			}
		}
	}
	c.mu.Unlock()

	var response *Response
	var cursor *Cursor
	var errchan chan error = make(chan error, 1)
	go func() {
		err := c.sendQuery(q)
		if err != nil {
			errchan <- err
			return
		}

		if noreply, ok := q.Opts["noreply"]; ok && noreply.(bool) {
			errchan <- nil
			return
		}

		for {
			response, err := c.readResponse()
			if err != nil {
				errchan <- err
				return
			}

			if response.Token == q.Token {
				// If this was the requested response process and return
				response, cursor, err = c.processResponse(ctx, q, response)
				errchan <- err
				return
			} else if _, ok := c.cursors[response.Token]; ok {
				// If the token is in the cursor cache then process the response
				c.processResponse(ctx, q, response)
			} else {
				putResponse(response)
			}
		}
	}()

	select {
	case err := <-errchan:
		return response, cursor, err
	case <-ctx.Done():
		if q.Type != p.Query_STOP {
			stopQuery := newStopQuery(q.Token)
			c.Query(c.contextFromConnectionOpts(), stopQuery)
		}
		return nil, nil, ErrQueryTimeout
	}
}

type ServerResponse struct {
	ID   string `gorethink:"id"`
	Name string `gorethink:"name"`
}

// Server returns the server name and server UUID being used by a connection.
func (c *Connection) Server() (ServerResponse, error) {
	var response ServerResponse

	_, cur, err := c.Query(c.contextFromConnectionOpts(), Query{
		Type: p.Query_SERVER_INFO,
	})
	if err != nil {
		return response, err
	}

	if err = cur.One(&response); err != nil {
		return response, err
	}

	if err = cur.Close(); err != nil {
		return response, err
	}

	return response, nil
}

// sendQuery marshals the Query and sends the JSON to the server.
func (c *Connection) sendQuery(q Query) error {
	// Build query
	b, err := json.Marshal(q.Build())
	if err != nil {
		return RQLDriverError{rqlError(fmt.Sprintf("Error building query: %s", err.Error()))}
	}

	// Set timeout
	if c.opts.WriteTimeout == 0 {
		c.Conn.SetWriteDeadline(time.Time{})
	} else {
		c.Conn.SetWriteDeadline(time.Now().Add(c.opts.WriteTimeout))
	}

	// Send the JSON encoding of the query itself.
	if err = c.writeQuery(q.Token, b); err != nil {
		c.bad = true
		return RQLConnectionError{rqlError(err.Error())}
	}

	return nil
}

// getToken generates the next query token, used to number requests and match
// responses with requests.
func (c *Connection) nextToken() int64 {
	// requires c.token to be 64-bit aligned on ARM
	return atomic.AddInt64(&c.token, 1)
}

// readResponse attempts to read a Response from the server, if no response
// could be read then an error is returned.
func (c *Connection) readResponse() (*Response, error) {
	// Set timeout
	if c.opts.ReadTimeout == 0 {
		c.Conn.SetReadDeadline(time.Time{})
	} else {
		c.Conn.SetReadDeadline(time.Now().Add(c.opts.ReadTimeout))
	}

	// Read response header (token+length)
	headerBuf := [respHeaderLen]byte{}
	if _, err := c.read(headerBuf[:], respHeaderLen); err != nil {
		c.bad = true
		return nil, RQLConnectionError{rqlError(err.Error())}
	}

	responseToken := int64(binary.LittleEndian.Uint64(headerBuf[:8]))
	messageLength := binary.LittleEndian.Uint32(headerBuf[8:])

	// Read the JSON encoding of the Response itself.
	b := make([]byte, int(messageLength))

	if _, err := c.read(b, int(messageLength)); err != nil {
		c.bad = true
		return nil, RQLConnectionError{rqlError(err.Error())}
	}

	// Decode the response
	var response = newCachedResponse()
	if err := json.Unmarshal(b, response); err != nil {
		c.bad = true
		return nil, RQLDriverError{rqlError(err.Error())}
	}
	response.Token = responseToken

	return response, nil
}

func (c *Connection) processResponse(ctx context.Context, q Query, response *Response) (*Response, *Cursor, error) {
	switch response.Type {
	case p.Response_CLIENT_ERROR:
		return c.processErrorResponse(q, response, RQLClientError{rqlServerError{response, q.Term}})
	case p.Response_COMPILE_ERROR:
		return c.processErrorResponse(q, response, RQLCompileError{rqlServerError{response, q.Term}})
	case p.Response_RUNTIME_ERROR:
		return c.processErrorResponse(q, response, createRuntimeError(response.ErrorType, response, q.Term))
	case p.Response_SUCCESS_ATOM, p.Response_SERVER_INFO:
		return c.processAtomResponse(ctx, q, response)
	case p.Response_SUCCESS_PARTIAL:
		return c.processPartialResponse(ctx, q, response)
	case p.Response_SUCCESS_SEQUENCE:
		return c.processSequenceResponse(ctx, q, response)
	case p.Response_WAIT_COMPLETE:
		return c.processWaitResponse(q, response)
	default:
		putResponse(response)
		return nil, nil, RQLDriverError{rqlError("Unexpected response type")}
	}
}

func (c *Connection) processErrorResponse(q Query, response *Response, err error) (*Response, *Cursor, error) {
	c.mu.Lock()
	cursor := c.cursors[response.Token]

	delete(c.cursors, response.Token)
	c.mu.Unlock()

	return response, cursor, err
}

func (c *Connection) processAtomResponse(ctx context.Context, q Query, response *Response) (*Response, *Cursor, error) {
	// Create cursor
	cursor := newCursor(ctx, c, "Cursor", response.Token, q.Term, q.Opts)
	cursor.profile = response.Profile

	cursor.extend(response)

	return response, cursor, nil
}

func (c *Connection) processPartialResponse(ctx context.Context, q Query, response *Response) (*Response, *Cursor, error) {
	cursorType := "Cursor"
	if len(response.Notes) > 0 {
		switch response.Notes[0] {
		case p.Response_SEQUENCE_FEED:
			cursorType = "Feed"
		case p.Response_ATOM_FEED:
			cursorType = "AtomFeed"
		case p.Response_ORDER_BY_LIMIT_FEED:
			cursorType = "OrderByLimitFeed"
		case p.Response_UNIONED_FEED:
			cursorType = "UnionedFeed"
		case p.Response_INCLUDES_STATES:
			cursorType = "IncludesFeed"
		}
	}

	c.mu.Lock()
	cursor, ok := c.cursors[response.Token]
	if !ok {
		// Create a new cursor if needed
		cursor = newCursor(ctx, c, cursorType, response.Token, q.Term, q.Opts)
		cursor.profile = response.Profile

		c.cursors[response.Token] = cursor
	}
	c.mu.Unlock()

	cursor.extend(response)

	return response, cursor, nil
}

func (c *Connection) processSequenceResponse(ctx context.Context, q Query, response *Response) (*Response, *Cursor, error) {
	c.mu.Lock()
	cursor, ok := c.cursors[response.Token]
	if !ok {
		// Create a new cursor if needed
		cursor = newCursor(ctx, c, "Cursor", response.Token, q.Term, q.Opts)
		cursor.profile = response.Profile
	}

	delete(c.cursors, response.Token)
	c.mu.Unlock()

	cursor.extend(response)

	return response, cursor, nil
}

func (c *Connection) processWaitResponse(q Query, response *Response) (*Response, *Cursor, error) {
	c.mu.Lock()
	delete(c.cursors, response.Token)
	c.mu.Unlock()

	return response, nil, nil
}

func (c *Connection) isBad() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.bad
}

var responseCache = make(chan *Response, 16)

func newCachedResponse() *Response {
	select {
	case r := <-responseCache:
		return r
	default:
		return new(Response)
	}
}

func putResponse(r *Response) {
	*r = Response{} // zero it
	select {
	case responseCache <- r:
	default:
	}
}
