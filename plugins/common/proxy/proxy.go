package proxy

import (
	"fmt"
	"golang.org/x/net/proxy"
	"net/http"
	"net/url"
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

func (p *TCPProxy) Proxy() (proxy.Dialer, error) {
	if len(p.ProxyURL) > 0 {
		parsed, err := url.Parse(p.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing proxy url %q: %w", p.ProxyURL, err)
		}

		return proxy.FromURL(parsed, proxy.Direct)
	}
	return proxy.FromEnvironment(), nil
}
