package proxy

import (
	"bufio"
	"crypto/tls"
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
