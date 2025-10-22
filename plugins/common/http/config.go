//go:generate ../../../tools/config_includer/generator "common.http" "transport.conf.in"
package httpconfig

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/peterbourgon/unixtransport"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/cookie"
	"github.com/influxdata/telegraf/plugins/common/oauth"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

// TransportConfig is configuration structure for HTTP transports
type TransportConfig struct {
	IdleConnTimeout       config.Duration `toml:"idle_conn_timeout"`
	MaxIdleConns          int             `toml:"max_idle_conn"`
	MaxIdleConnsPerHost   int             `toml:"max_idle_conn_per_host"`
	ResponseHeaderTimeout config.Duration `toml:"response_timeout"`
	proxy.HTTPProxy
	tls.ClientConfig
}

func (h *TransportConfig) CreateTransport() (*http.Transport, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("creating TLS configuration failed: %w", err)
	}

	prox, err := h.HTTPProxy.Proxy()
	if err != nil {
		return nil, fmt.Errorf("setting up proxy failed: %w", err)
	}

	return &http.Transport{
		TLSClientConfig:       tlsCfg,
		Proxy:                 prox,
		IdleConnTimeout:       time.Duration(h.IdleConnTimeout),
		MaxIdleConns:          h.MaxIdleConns,
		MaxIdleConnsPerHost:   h.MaxIdleConnsPerHost,
		ResponseHeaderTimeout: time.Duration(h.ResponseHeaderTimeout),
	}, nil
}

// HTTPClientConfig is a common HTTP client struct.
type HTTPClientConfig struct {
	Timeout               config.Duration `toml:"timeout"`
	IdleConnTimeout       config.Duration `toml:"idle_conn_timeout"`
	MaxIdleConns          int             `toml:"max_idle_conn"`
	MaxIdleConnsPerHost   int             `toml:"max_idle_conn_per_host"`
	ResponseHeaderTimeout config.Duration `toml:"response_timeout"`
	proxy.HTTPProxy
	tls.ClientConfig
	oauth.OAuth2Config
	cookie.CookieAuthConfig
}

func (h *HTTPClientConfig) CreateClient(ctx context.Context, log telegraf.Logger) (*http.Client, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("creating TLS configuration failed: %w", err)
	}

	prox, err := h.HTTPProxy.Proxy()
	if err != nil {
		return nil, fmt.Errorf("setting up proxy failed: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig:       tlsCfg,
		Proxy:                 prox,
		IdleConnTimeout:       time.Duration(h.IdleConnTimeout),
		MaxIdleConns:          h.MaxIdleConns,
		MaxIdleConnsPerHost:   h.MaxIdleConnsPerHost,
		ResponseHeaderTimeout: time.Duration(h.ResponseHeaderTimeout),
	}

	// Register "http+unix" and "https+unix" protocol handler.
	unixtransport.Register(transport)

	client := &http.Client{
		Transport: transport,
	}

	// While CreateOauth2Client returns a http.Client keeping the Transport configuration,
	// it does not keep other http.Client parameters (e.g. Timeout).
	client = h.OAuth2Config.CreateOauth2Client(ctx, client)

	if h.CookieAuthConfig.URL != "" {
		if err := h.CookieAuthConfig.Start(client, log, clock.New()); err != nil {
			return nil, err
		}
	}

	timeout := h.Timeout
	if timeout == 0 {
		timeout = config.Duration(time.Second * 5)
	}
	client.Timeout = time.Duration(timeout)

	return client, nil
}
