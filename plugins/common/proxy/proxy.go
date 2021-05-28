package proxy

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

type HTTPProxy struct {
	HTTPProxyURL string `toml:"http_proxy_url"`
}

type proxyFunc func(req *http.Request) (*url.URL, error)

func (p *HTTPProxy) Proxy() (proxyFunc, error) {
	if len(p.HTTPProxyURL) > 0 {
		address, err := url.Parse(p.HTTPProxyURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing proxy url %q: %w", p.HTTPProxyURL, err)
		}
		return http.ProxyURL(address), nil
	}
	return http.ProxyFromEnvironment, nil
}

type TCPProxy struct {
	ProxyURL string `toml:"proxy_url"`
}

func (p *TCPProxy) Proxy() (*ProxiedDialer, error) {
	var dialer proxy.Dialer
	if len(p.ProxyURL) > 0 {
		parsed, err := url.Parse(p.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing proxy url %q: %w", p.ProxyURL, err)
		}

		if dialer, err = proxy.FromURL(parsed, proxy.Direct); err != nil {
			return nil, err
		}
	} else {
		dialer = proxy.FromEnvironment()
	}

	return &ProxiedDialer{dialer}, nil
}
