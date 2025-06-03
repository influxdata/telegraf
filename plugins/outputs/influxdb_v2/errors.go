package influxdb_v2

import (
	"fmt"
	"strings"
	"time"
)

type genericRespError struct {
	Code      string
	Message   string
	Line      *int32
	MaxLength *int32
}

func (g genericRespError) Error() string {
	errString := fmt.Sprintf("%s: %s", g.Code, g.Message)
	var details []string
	if g.Line != nil {
		details = append(details, fmt.Sprintf("line[%d]", *g.Line))
	}
	if g.MaxLength != nil {
		details = append(details, fmt.Sprintf("maxlen[%d]", *g.MaxLength))
	}
	if len(details) > 0 {
		errString += " - " + strings.Join(details, ", ")
	}
	return errString
}

type APIError struct {
	Err        error
	StatusCode int
	Retryable  bool
}

func (e APIError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("API error: status %d", e.StatusCode)
	}
	return e.Err.Error()
}

func (e APIError) Unwrap() error {
	return e.Err
}

type ThrottleError struct {
	Err        error
	StatusCode int
	RetryAfter time.Duration
}

func (e ThrottleError) Error() string {
	return e.Err.Error()
}

func (e ThrottleError) Unwrap() error {
	return e.Err
}
