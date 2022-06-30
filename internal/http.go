package internal

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/url"
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

type GenericAuthErrorFunc func(rw http.ResponseWriter)

// GenericAuthHandler returns a http handler that requires `Authorization: <credentials>`
func GenericAuthHandler(credentials string, onError GenericAuthErrorFunc) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &genericAuthHandler{
			credentials: credentials,
			onError:     onError,
			next:        h,
		}
	}
}

// Generic auth scheme handler - exact match on `Authorization: <credentials>`
type genericAuthHandler struct {
	credentials string
	onError     GenericAuthErrorFunc
	next        http.Handler
}

func (h *genericAuthHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.credentials != "" {
		// Scheme checking
		authorization := req.Header.Get("Authorization")
		if subtle.ConstantTimeCompare([]byte(authorization), []byte(h.credentials)) != 1 {
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
func IPRangeHandler(networks []*net.IPNet, onError ErrorFunc) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &ipRangeHandler{
			networks: networks,
			onError:  onError,
			next:     h,
		}
	}
}

type ipRangeHandler struct {
	networks []*net.IPNet
	onError  ErrorFunc
	next     http.Handler
}

func (h *ipRangeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(h.networks) == 0 {
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

	for _, network := range h.networks {
		if network.Contains(remoteIP) {
			h.next.ServeHTTP(rw, req)
			return
		}
	}

	h.onError(rw, http.StatusForbidden)
}

func OnClientError(client *http.Client, err error) {
	// Close connection after a timeout error. If this is a HTTP2
	// connection this ensures that next interval a new connection will be
	// used and name lookup will be performed.
	//   https://github.com/golang/go/issues/36026
	if err, ok := err.(*url.Error); ok && err.Timeout() {
		client.CloseIdleConnections()
	}
}
