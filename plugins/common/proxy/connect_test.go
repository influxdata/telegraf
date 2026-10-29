package proxy

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/proxy"

	"github.com/influxdata/telegraf/testutil"
)

func TestHTTPConnectProxy(t *testing.T) {
	pki := testutil.NewPKI("../../../testutil/pki")
	serverConfig, err := pki.TLSServerConfig().TLSConfig()
	require.NoError(t, err)
	clientConfig, err := pki.TLSClientConfig().TLSConfig()
	require.NoError(t, err)

	tests := []struct {
		name   string
		scheme string
		listen func() (net.Listener, error)
	}{
		{
			name:   "http",
			scheme: "http",
			listen: func() (net.Listener, error) {
				return net.Listen("tcp", "127.0.0.1:0")
			},
		},
		{
			name:   "https",
			scheme: "https",
			listen: func() (net.Listener, error) {
				return tls.Listen("tcp", "127.0.0.1:0", serverConfig)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := tt.listen()
			require.NoError(t, err)
			defer listener.Close()

			// Minimal CONNECT proxy echoing everything sent through the tunnel
			go func() {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				defer conn.Close()

				reader := bufio.NewReader(conn)
				req, err := http.ReadRequest(reader)
				if err != nil || req.Method != http.MethodConnect {
					return
				}
				if _, err := io.WriteString(conn, "HTTP/1.1 200 Connection established\r\n\r\n"); err != nil {
					return
				}
				//nolint:errcheck // the client closes the tunnel when it is done
				io.Copy(conn, reader)
			}()

			proxyURL := &url.URL{Scheme: tt.scheme, Host: listener.Addr().String()}
			dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
			require.NoError(t, err)

			// Trust the test CA and present the client certificate the test
			// server requires, everything else stays as in production.
			if tt.scheme == "https" {
				tlsConfig := dialer.(*httpConnectProxy).tlsConfig
				tlsConfig.RootCAs = clientConfig.RootCAs
				tlsConfig.Certificates = clientConfig.Certificates
			}

			conn, err := dialer.Dial("tcp", "example.com:443")
			require.NoError(t, err)
			defer conn.Close()
			require.NoError(t, conn.SetDeadline(time.Now().Add(5*time.Second)))

			_, err = conn.Write([]byte("ping"))
			require.NoError(t, err)
			buf := make([]byte, 4)
			_, err = io.ReadFull(conn, buf)
			require.NoError(t, err)
			require.Equal(t, "ping", string(buf))
		})
	}
}

// TestHTTPConnectProxyAuthorizationHeader covers the fix for issue #19885.
// Proxy credentials were sent with req.SetBasicAuth, which sets Authorization. That header
// authenticates against the origin resource; a proxy following RFC 9110 section 11.7.2
// looks for Proxy-Authorization, ignores Authorization and refuses the tunnel with
// "407 Proxy Authentication Required".
func TestHTTPConnectProxyAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name     string
		userinfo string
		expected string
	}{
		{
			name:     "username and password",
			userinfo: "user:s3cret@",
			expected: "Basic " + base64.StdEncoding.EncodeToString([]byte("user:s3cret")),
		},
		{
			// net/http's own connectMethod.proxyAuth sends the header whenever the
			// userinfo part is present, password or not, so a username-only proxy
			// URL still authenticates.
			name:     "username only",
			userinfo: "user@",
			expected: "Basic " + base64.StdEncoding.EncodeToString([]byte("user:")),
		},
		{
			name:     "no credentials",
			userinfo: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			defer listener.Close()

			headers := make(chan http.Header, 1)
			go func() {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				defer conn.Close()

				// Parsed rather than string-matched, so a header folded differently or
				// cased differently still registers, and so an assertion cannot pass on
				// a substring that happens to appear elsewhere in the request.
				req, err := http.ReadRequest(bufio.NewReader(conn))
				if err != nil {
					close(headers)
					return
				}
				headers <- req.Header
				// 403 rather than 200 on purpose: the assertion is about the request that
				// reached the proxy, and a 200 would hand back a half-open tunnel this
				// test has no use for.
				fmt.Fprint(conn, "HTTP/1.1 403 Forbidden\r\nContent-Length: 0\r\n\r\n")
			}()

			p := &TCPProxy{UseProxy: true, ProxyURL: "http://" + tt.userinfo + listener.Addr().String()}
			dialer, err := p.Proxy()
			require.NoError(t, err)

			conn, err := dialer.Dial("tcp", "example.com:443")
			if conn != nil {
				conn.Close()
			}
			// Asserted so a future change that stops dialling altogether cannot make
			// these pass vacuously.
			require.Error(t, err)

			var got http.Header
			select {
			case h, ok := <-headers:
				require.True(t, ok, "proxy could not parse the CONNECT request")
				got = h
			case <-time.After(10 * time.Second):
				require.FailNow(t, "timed out waiting for the CONNECT request")
			}

			require.Equal(t, tt.expected, got.Get("Proxy-Authorization"), "proxy credentials not set")
			require.Empty(t, got.Get("Authorization"), "unexpected Authorization header")
		})
	}
}
