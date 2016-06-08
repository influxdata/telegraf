package errchan

import (
	"fmt"
	"strings"
)

type ErrChan struct {
	C chan error
}

// New returns an error channel of max length 'n'
// errors can be sent to the ErrChan.C channel, and will be returned when
// ErrChan.Error() is called.
func New(n int) *ErrChan {
	return &ErrChan{
		C: make(chan error, n),
	}
}

// Error closes the ErrChan.C channel and returns an error if there are any
// non-nil errors, otherwise returns nil.
func (e *ErrChan) Error() error {
	close(e.C)

	var out string
	for err := range e.C {
		if err != nil {
			out += "[" + err.Error() + "], "
		}
	}

	if out != "" {
		return fmt.Errorf("Errors encountered: " + strings.TrimRight(out, ", "))
	}
	return nil
}
