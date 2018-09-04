package ts3

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// ErrInvalidConnectHeader is returned by NewClient if the server doesn't respond with the required connection header.
	ErrInvalidConnectHeader = errors.New("invalid connect header")

	// ErrNilOption is returned by NewClient if an option is nil.
	ErrNilOption = errors.New("nil option")
)

// Error represents a error returned from the TeamSpeak 3 server.
type Error struct {
	ID      int
	Msg     string
	Details map[string]interface{}
}

// NewError returns a new Error parsed from TeamSpeak 3 server response.
func NewError(matches []string) *Error {
	e := &Error{Msg: Decode(matches[2])}

	var err error
	if e.ID, err = strconv.Atoi(matches[1]); err != nil {
		// This should be impossible given it matched \d+ in the regexp.
		e.ID = -1
	}

	if rem := strings.TrimSpace(matches[3]); rem != "" {
		e.Details = make(map[string]interface{})
		for _, s := range strings.Split(rem, " ") {
			d := strings.SplitN(s, "=", 2)
			v := Decode(d[0])
			if i, err := strconv.Atoi(d[1]); err == nil {
				e.Details[v] = i
			} else {
				e.Details[v] = Decode(d[1])
			}
		}
	}

	return e
}

func (e *Error) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("%v %v (%v)", e.Msg, e.Details, e.ID)
	}
	return fmt.Sprintf("%v (%v)", e.Msg, e.ID)
}

// InvalidResponseError is the error returned when the response data was invalid.
type InvalidResponseError struct {
	Reason string
	Data   []string
}

// NewInvalidResponseError returns a new InvalidResponseError from lines.
func NewInvalidResponseError(reason string, lines []string) *InvalidResponseError {
	return &InvalidResponseError{Reason: reason, Data: lines}
}

func (e *InvalidResponseError) Error() string {
	return fmt.Sprintf("%v (%+v)", e.Reason, e.Data)
}
