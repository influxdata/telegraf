package proxy

import (
	"fmt"
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
