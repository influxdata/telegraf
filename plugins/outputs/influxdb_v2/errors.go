package influxdb_v2

import (
	"fmt"
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
	if g.Line != nil {
		return fmt.Sprintf("%s - line[%d]", errString, g.Line)
	} else if g.MaxLength != nil {
		return fmt.Sprintf("%s - maxlen[%d]", errString, g.MaxLength)
	}
	return errString
}

type APIError struct {
	Err        error
	StatusCode int
	Retryable  bool
}

func (e APIError) Error() string {
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
