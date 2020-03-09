package internal

import (
	"crypto/subtle"
	"net"
	"net/http"
)

type BasicAuthErrorFunc func(rw http.ResponseWriter)

// AuthHandler returns a http handler that requires HTTP basic auth
// credentials to match the given username and password.
func AuthHandler(username, password, realm string, onError BasicAuthErrorFunc) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &basicAuthHandler{
			username: username,
			password: password,
			realm:    realm,
			onError:  onError,
			next:     h,
		}
	}
}

type basicAuthHandler struct {
	username string
	password string
	realm    string
	onError  BasicAuthErrorFunc
	next     http.Handler
}

func (h *basicAuthHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.username != "" || h.password != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.password)) != 1 {

			rw.Header().Set("WWW-Authenticate", "Basic realm=\""+h.realm+"\"")
			h.onError(rw)
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	h.next.ServeHTTP(rw, req)
}

// ErrorFunc is a callback for writing an error response.
type ErrorFunc func(rw http.ResponseWriter, code int)

// IPRangeHandler returns a http handler that requires the remote address to be
// in the specified network.
func IPRangeHandler(network []*net.IPNet, onError ErrorFunc) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &ipRangeHandler{
			network: network,
			onError: onError,
			next:    h,
		}
	}
}

type ipRangeHandler struct {
	network []*net.IPNet
	onError ErrorFunc
	next    http.Handler
}

func (h *ipRangeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(h.network) == 0 {
		h.next.ServeHTTP(rw, req)
		return
	}

	remoteIPString, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		h.onError(rw, http.StatusForbidden)
		return
	}

	remoteIP := net.ParseIP(remoteIPString)
	if remoteIP == nil {
		h.onError(rw, http.StatusForbidden)
		return
	}

	for _, net := range h.network {
		if net.Contains(remoteIP) {
			h.next.ServeHTTP(rw, req)
			return
		}
	}

	h.onError(rw, http.StatusForbidden)
}
