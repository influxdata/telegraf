package proxy

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

// httpConnectProxy proxies (only?) TCP over a HTTP tunnel using the CONNECT method
type httpConnectProxy struct {
	forward proxy.Dialer
	url     *url.URL
}

func (c *httpConnectProxy) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var proxyConn net.Conn
	var err error
	if dialer, ok := c.forward.(proxy.ContextDialer); ok {
		proxyConn, err = dialer.DialContext(ctx, "tcp", addr)
	} else {
		// TODO: still support timeout/cancellation w/o context
		proxyConn, err = c.forward.Dial("tcp", addr)
	}
	if err != nil {
		return nil, err
	}

	// Add and strip http:// to extract authority portion of the URL
	// since CONNECT doesn't use a full URL. The request header would
	// look something like: "CONNECT www.influxdata.com:443 HTTP/1.1"
	requestURL, err := url.Parse("http://" + addr)
	if err != nil {
		proxyConn.Close()
		return nil, err
	}
	requestURL.Scheme = ""

	// Build HTTP CONNECT request
	req, err := http.NewRequest("CONNECT", requestURL.String(), nil)
	if err != nil {
		proxyConn.Close()
		return nil, err
	}
	req.Close = false
	if password, hasAuth := c.url.User.Password(); hasAuth {
		req.SetBasicAuth(c.url.User.Username(), password)
	}

	err = req.Write(proxyConn)
	if err != nil {
		proxyConn.Close()
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), req)
	if err != nil {
		proxyConn.Close()
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to connect to proxy: %q", resp.Status)
	}

	return proxyConn, nil
}

func (c *httpConnectProxy) Dial(network, addr string) (net.Conn, error) {
	return c.DialContext(context.Background(), network, addr)
}

func newHTTPConnectProxy(url *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	return &httpConnectProxy{forward, url}, nil
}

func init() {
	// Register new proxy types
	proxy.RegisterDialerType("http", newHTTPConnectProxy)
	proxy.RegisterDialerType("https", newHTTPConnectProxy)
}
