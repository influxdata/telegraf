package proxy

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// serveOneCONNECT accepts a single connection, reads the request the dialer sent and
// answers with status. It returns the address to point a proxy URL at and a channel
// carrying the parsed request headers.
func serveOneCONNECT(t *testing.T, status string) (string, <-chan http.Header) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() })

	headers := make(chan http.Header, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Parsed rather than string-matched, so a header folded differently or cased
		// differently still registers, and so an assertion cannot pass on a substring
		// that happens to appear elsewhere in the request.
		req, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			close(headers)
			return
		}
		headers <- req.Header
		fmt.Fprintf(conn, "HTTP/1.1 %s\r\nContent-Length: 0\r\n\r\n", status)
	}()

	return listener.Addr().String(), headers
}

func awaitHeaders(t *testing.T, headers <-chan http.Header) http.Header {
	t.Helper()

	select {
	case got, ok := <-headers:
		require.True(t, ok, "proxy could not parse the CONNECT request")
		return got
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the CONNECT request")
		return nil
	}
}

// TestHTTPConnectProxyCredentialsUseProxyAuthorization covers the fix for issue #19885.
// Proxy credentials were sent with req.SetBasicAuth, which sets Authorization. That header
// authenticates against the origin resource; a proxy following RFC 9110 section 11.7.2
// looks for Proxy-Authorization, ignores Authorization and refuses the tunnel with
// "407 Proxy Authentication Required".
func TestHTTPConnectProxyCredentialsUseProxyAuthorization(t *testing.T) {
	addr, headers := serveOneCONNECT(t, "403 Forbidden")

	p := &TCPProxy{UseProxy: true, ProxyURL: "http://user:s3cret@" + addr}
	dialer, err := p.Proxy()
	require.NoError(t, err)

	// The 403 means the dial fails, which is deliberate: the assertion is about the
	// request that reached the proxy, and answering 200 would hand back a half-open
	// tunnel this test has no use for.
	conn, err := dialer.Dial("tcp", "example.com:443")
	if conn != nil {
		conn.Close()
	}
	require.Error(t, err)

	got := awaitHeaders(t, headers)
	expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:s3cret"))

	require.Equal(t, expected, got.Get("Proxy-Authorization"),
		"proxy credentials must be sent in Proxy-Authorization")
	require.Empty(t, got.Get("Authorization"),
		"proxy credentials must not also be sent in Authorization, which addresses the origin")
}

// TestHTTPConnectProxyWithoutCredentialsSendsNoAuth guards the other direction: a proxy URL
// with no password must not gain an empty credential header.
func TestHTTPConnectProxyWithoutCredentialsSendsNoAuth(t *testing.T) {
	addr, headers := serveOneCONNECT(t, "403 Forbidden")

	p := &TCPProxy{UseProxy: true, ProxyURL: "http://" + addr}
	dialer, err := p.Proxy()
	require.NoError(t, err)

	conn, err := dialer.Dial("tcp", "example.com:443")
	if conn != nil {
		conn.Close()
	}
	require.Error(t, err)

	got := awaitHeaders(t, headers)
	require.Empty(t, got.Get("Proxy-Authorization"))
	require.Empty(t, got.Get("Authorization"))
}

// TestHTTPConnectProxyUsernameOnlySendsNoAuth documents the existing behaviour that the fix
// deliberately leaves alone: the credential header is gated on Password() reporting a
// password being present, so a URL carrying only a username sends nothing.
func TestHTTPConnectProxyUsernameOnlySendsNoAuth(t *testing.T) {
	addr, headers := serveOneCONNECT(t, "403 Forbidden")

	p := &TCPProxy{UseProxy: true, ProxyURL: "http://user@" + addr}
	dialer, err := p.Proxy()
	require.NoError(t, err)

	conn, err := dialer.Dial("tcp", "example.com:443")
	if conn != nil {
		conn.Close()
	}
	require.Error(t, err)

	got := awaitHeaders(t, headers)
	require.Empty(t, got.Get("Proxy-Authorization"))
	require.Empty(t, got.Get("Authorization"))
}
