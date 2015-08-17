package gorethink

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	p "github.com/dancannon/gorethink/ql2"
)

var (
	// ErrNoHosts is returned when no hosts to the Connect method.
	ErrNoHosts = errors.New("no hosts provided")
	// ErrNoConnectionsStarted is returned when the driver couldn't to any of
	// the provided hosts.
	ErrNoConnectionsStarted = errors.New("no connections were made when creating the session")
	// ErrInvalidNode is returned when attempting to connect to a node which
	// returns an invalid response.
	ErrInvalidNode = errors.New("invalid node")
	// ErrClusterClosed is returned when a query is executed after the connection
	// to the cluster has been closed.
	ErrClusterClosed = errors.New("cluster closed")
	// ErrNoConnections is returned when there are no active connections in the
	// clusters connection pool.
	ErrNoConnections = errors.New("gorethink: no connections were available")
	// ErrConnectionClosed is returned when trying to send a query with a closed
	// connection.
	ErrConnectionClosed = errors.New("gorethink: the connection is closed")
)

func printCarrots(t Term, frames []*p.Frame) string {
	var frame *p.Frame
	if len(frames) > 1 {
		frame, frames = frames[0], frames[1:]
	} else if len(frames) == 1 {
		frame, frames = frames[0], []*p.Frame{}
	}

	for i, arg := range t.args {
		if frame.GetPos() == int64(i) {
			t.args[i] = Term{
				termType: p.Term_DATUM,
				data:     printCarrots(arg, frames),
			}
		}
	}

	for k, arg := range t.optArgs {
		if frame.GetOpt() == k {
			t.optArgs[k] = Term{
				termType: p.Term_DATUM,
				data:     printCarrots(arg, frames),
			}
		}
	}

	b := &bytes.Buffer{}
	for _, c := range t.String() {
		if c != '^' {
			b.WriteString(" ")
		} else {
			b.WriteString("^")
		}
	}

	return b.String()
}

// Error constants
var ErrEmptyResult = errors.New("The result does not contain any more rows")

// Connection/Response errors

// rqlResponseError is the base type for all errors, it formats both
// for the response and query if set.
type rqlResponseError struct {
	response *Response
	term     *Term
}

func (e rqlResponseError) Error() string {
	var err = "An error occurred"
	if e.response != nil {
		json.Unmarshal(e.response.Responses[0], &err)
	}

	if e.term == nil {
		return fmt.Sprintf("gorethink: %s", err)
	}

	return fmt.Sprintf("gorethink: %s in: \n%s", err, e.term.String())

}

func (e rqlResponseError) String() string {
	return e.Error()
}

// RQLCompileError represents an error that occurs when compiling a query on
// the database server.
type RQLCompileError struct {
	rqlResponseError
}

// RQLRuntimeError represents an error when executing an error on the database
// server, this is also returned by the database when using the `Error` term.
type RQLRuntimeError struct {
	rqlResponseError
}

// RQLClientError represents a client error returned from the database.
type RQLClientError struct {
	rqlResponseError
}

// RQLDriverError represents an unexpected error with the driver, if this error
// persists please create an issue.
type RQLDriverError struct {
	message string
}

func (e RQLDriverError) Error() string {
	return fmt.Sprintf("gorethink: %s", e.message)
}

func (e RQLDriverError) String() string {
	return e.Error()
}

// RQLConnectionError represents an error when communicating with the database
// server.
type RQLConnectionError struct {
	message string
}

func (e RQLConnectionError) Error() string {
	return fmt.Sprintf("gorethink: %s", e.message)
}

func (e RQLConnectionError) String() string {
	return e.Error()
}
