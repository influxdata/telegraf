package api

import "fmt"

type Error struct {
	StatusCode int
}

func (e *Error) Error() string {
	return fmt.Sprintf("api: HTTP status code %d", e.StatusCode)
}

func IsErrorWithStatusCode(err error, code int) bool {
	if err, ok := err.(*Error); ok {
		return err.StatusCode == code
	}
	return false
}
