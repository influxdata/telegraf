package internal

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/url"
	"strings"
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

type TokenAuthErrorFunc func(rw http.ResponseWriter)

// TokenAuthHandler returns a http handler that requires `Authorization: Token <token>`
// Introduced to support InfluxDB 2.x style authentication
// https://v2.docs.influxdata.com/v2.0/reference/api/#authentication
func TokenAuthHandler(token string, onError TokenAuthErrorFunc) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &schemeAuthHandler{
			scheme:      "Token",
			credentials: token,
			onError:     onError,
			next:        h,
		}
	}
}

// General auth scheme handler - match `Authorization: <scheme> <credentials>`
type schemeAuthHandler struct {
	scheme      string
	credentials string
	onError     TokenAuthErrorFunc
	next        http.Handler
}

func (h *schemeAuthHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.scheme != "" || h.credentials != "" {
		// Scheme checking
		authHeader := req.Header.Get("Authorization")
		authParts := strings.SplitN(authHeader, " ", 2)
		if len(authParts) != 2 ||
			subtle.ConstantTimeCompare(
				[]byte(strings.ToLower(strings.TrimSpace(authParts[0]))),
				[]byte(strings.ToLower(h.scheme))) != 1 ||
			subtle.ConstantTimeCompare([]byte(strings.TrimSpace(authParts[1])), []byte(h.credentials)) != 1 {

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

func OnClientError(client *http.Client, err error) {
	// Close connection after a timeout error. If this is a HTTP2
	// connection this ensures that next interval a new connection will be
	// used and name lookup will be performed.
	//   https://github.com/golang/go/issues/36026
	if err, ok := err.(*url.Error); ok && err.Timeout() {
		client.CloseIdleConnections()
	}
}
