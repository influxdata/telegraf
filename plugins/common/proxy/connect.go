package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"

	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
)

// httpConnectProxy proxies (only?) TCP over a HTTP tunnel using the CONNECT method
type httpConnectProxy struct {
	forward   proxy.Dialer
	url       *url.URL
	tlsConfig *tls.Config
}

func (c *httpConnectProxy) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	// Prevent using UDP
	if network == "udp" {
		return nil, fmt.Errorf("cannot proxy %q traffic over HTTP CONNECT", network)
	}

	var proxyConn net.Conn
	var err error
	if dialer, ok := c.forward.(proxy.ContextDialer); ok {
		proxyConn, err = dialer.DialContext(ctx, "tcp", c.url.Host)
	} else {
		shim := contextDialerShim{c.forward}
		proxyConn, err = shim.DialContext(ctx, "tcp", c.url.Host)
	}
	if err != nil {
		return nil, err
	}

	// Secure the connection to an https proxy before the CONNECT request
	// and the credentials it may carry are sent.
	if c.tlsConfig != nil {
		tlsConn := tls.Client(proxyConn, c.tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			if err := proxyConn.Close(); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("tls handshake with proxy failed: %w", err)
		}
		proxyConn = tlsConn
	}

	// Add and strip http:// to extract authority portion of the URL
	// since CONNECT doesn't use a full URL. The request header would
	// look something like: "CONNECT www.influxdata.com:443 HTTP/1.1"
	requestURL, err := url.Parse("http://" + addr)
	if err != nil {
		if err := proxyConn.Close(); err != nil {
			return nil, err
		}
		return nil, err
	}
	requestURL.Scheme = ""

	// Build HTTP CONNECT request
	req, err := http.NewRequest(http.MethodConnect, requestURL.String(), nil)
	if err != nil {
		if err := proxyConn.Close(); err != nil {
			return nil, err
		}
		return nil, err
	}
	req.Close = false
	if password, hasAuth := c.url.User.Password(); hasAuth {
		req.SetBasicAuth(c.url.User.Username(), password)
	}

	err = req.Write(proxyConn)
	if err != nil {
		if err := proxyConn.Close(); err != nil {
			return nil, err
		}
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), req)
	if err != nil {
		if err := proxyConn.Close(); err != nil {
			return nil, err
		}
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		if err := proxyConn.Close(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to connect to proxy: %q", resp.Status)
	}

	return proxyConn, nil
}

func (c *httpConnectProxy) Dial(network, addr string) (net.Conn, error) {
	return c.DialContext(context.Background(), network, addr)
}

func newHTTPConnectProxy(proxyURL *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	c := &httpConnectProxy{forward: forward, url: proxyURL}
	if proxyURL.Scheme == "https" {
		c.tlsConfig = &tls.Config{
			ServerName: proxyURL.Hostname(),
			MinVersion: common_tls.TLSMinVersionDefault,
		}
	}
	return c, nil
}

func init() {
	// Register new proxy types
	proxy.RegisterDialerType("http", newHTTPConnectProxy)
	proxy.RegisterDialerType("https", newHTTPConnectProxy)
}

// contextDialerShim allows cancellation of the dial from a context even if the underlying
// dialer does not implement `proxy.ContextDialer`. Arguably, this shouldn't actually get run,
// unless a new proxy type is added that doesn't implement `proxy.ContextDialer`, as all the
// standard library dialers implement `proxy.ContextDialer`.
type contextDialerShim struct {
	dialer proxy.Dialer
}

func (cd *contextDialerShim) Dial(network, addr string) (net.Conn, error) {
	return cd.dialer.Dial(network, addr)
}

func (cd *contextDialerShim) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var (
		conn net.Conn
		done = make(chan struct{}, 1)
		err  error
	)

	go func() {
		conn, err = cd.dialer.Dial(network, addr)
		close(done)
		if conn != nil && ctx.Err() != nil {
			_ = conn.Close()
		}
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-done:
	}

	return conn, err
}
