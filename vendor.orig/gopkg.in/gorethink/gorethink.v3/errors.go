package gorethink

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	p "gopkg.in/gorethink/gorethink.v3/ql2"
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
	// ErrNoConnections is returned when there are no active connections in the
	// clusters connection pool.
	ErrNoConnections = errors.New("gorethink: no connections were available")
	// ErrConnectionClosed is returned when trying to send a query with a closed
	// connection.
	ErrConnectionClosed = errors.New("gorethink: the connection is closed")
	// ErrQueryTimeout is returned when query context deadline exceeded.
	ErrQueryTimeout = errors.New("gorethink: query timeout")
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
type rqlServerError struct {
	response *Response
	term     *Term
}

func (e rqlServerError) Error() string {
	var err = "An error occurred"
	if e.response != nil {
		json.Unmarshal(e.response.Responses[0], &err)
	}

	if e.term == nil {
		return fmt.Sprintf("gorethink: %s", err)
	}

	return fmt.Sprintf("gorethink: %s in:\n%s", err, e.term.String())

}

func (e rqlServerError) String() string {
	return e.Error()
}

type rqlError string

func (e rqlError) Error() string {
	return fmt.Sprintf("gorethink: %s", string(e))
}

func (e rqlError) String() string {
	return e.Error()
}

// Exported Error "Implementations"

type RQLClientError struct{ rqlServerError }
type RQLCompileError struct{ rqlServerError }
type RQLDriverCompileError struct{ RQLCompileError }
type RQLServerCompileError struct{ RQLCompileError }
type RQLAuthError struct{ RQLDriverError }
type RQLRuntimeError struct{ rqlServerError }

type RQLQueryLogicError struct{ RQLRuntimeError }
type RQLNonExistenceError struct{ RQLQueryLogicError }
type RQLResourceLimitError struct{ RQLRuntimeError }
type RQLUserError struct{ RQLRuntimeError }
type RQLInternalError struct{ RQLRuntimeError }
type RQLTimeoutError struct{ rqlServerError }
type RQLAvailabilityError struct{ RQLRuntimeError }
type RQLOpFailedError struct{ RQLAvailabilityError }
type RQLOpIndeterminateError struct{ RQLAvailabilityError }

// RQLDriverError represents an unexpected error with the driver, if this error
// persists please create an issue.
type RQLDriverError struct {
	rqlError
}

// RQLConnectionError represents an error when communicating with the database
// server.
type RQLConnectionError struct {
	rqlError
}

func createRuntimeError(errorType p.Response_ErrorType, response *Response, term *Term) error {
	serverErr := rqlServerError{response, term}

	switch errorType {
	case p.Response_QUERY_LOGIC:
		return RQLQueryLogicError{RQLRuntimeError{serverErr}}
	case p.Response_NON_EXISTENCE:
		return RQLNonExistenceError{RQLQueryLogicError{RQLRuntimeError{serverErr}}}
	case p.Response_RESOURCE_LIMIT:
		return RQLResourceLimitError{RQLRuntimeError{serverErr}}
	case p.Response_USER:
		return RQLUserError{RQLRuntimeError{serverErr}}
	case p.Response_INTERNAL:
		return RQLInternalError{RQLRuntimeError{serverErr}}
	case p.Response_OP_FAILED:
		return RQLOpFailedError{RQLAvailabilityError{RQLRuntimeError{serverErr}}}
	case p.Response_OP_INDETERMINATE:
		return RQLOpIndeterminateError{RQLAvailabilityError{RQLRuntimeError{serverErr}}}
	default:
		return RQLRuntimeError{serverErr}
	}
}

// Error type helpers

// IsConflictErr returns true if the error is non-nil and the query failed
// due to a duplicate primary key.
func IsConflictErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.HasPrefix(err.Error(), "Duplicate primary key")
}

// IsTypeErr returns true if the error is non-nil and the query failed due
// to a type error.
func IsTypeErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.HasPrefix(err.Error(), "Expected type")
}
