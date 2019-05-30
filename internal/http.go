package internal

import (
	"crypto/subtle"
	"net/http"
)

// ErrorFunc is a callback for writing an error response.
type ErrorFunc func(rw http.ResponseWriter, code int)

// AuthHandler returns a http handler that requires HTTP basic auth
// credentials to match the given username and password.
func AuthHandler(username, password string, onError ErrorFunc) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &basicAuthHandler{
			username: username,
			password: password,
			onError:  onError,
			next:     h,
		}
	}

}

type basicAuthHandler struct {
	username string
	password string
	onError  ErrorFunc
	next     http.Handler
}

func (h *basicAuthHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.username != "" || h.password != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.password)) != 1 {

			h.onError(rw, http.StatusUnauthorized)
			return
		}
	}

	h.next.ServeHTTP(rw, req)
}
