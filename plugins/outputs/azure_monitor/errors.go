package azure_monitor

import "fmt"

// httpError represents an HTTP error with retry information.
type httpError struct {
	err        error
	statusCode int
	retryable  bool
}

func (e *httpError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("HTTP error: status %d", e.statusCode)
	}
	return e.err.Error()
}

func (e *httpError) Unwrap() error {
	return e.err
}
