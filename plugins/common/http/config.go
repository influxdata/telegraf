package httpconfig

import (
	"context"
	"fmt"
	"github.com/benbjohnson/clock"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/cookie"
	oauthConfig "github.com/influxdata/telegraf/plugins/common/oauth"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"net/http"
	"time"
)

// Common HTTP client struct.
type HTTPClientConfig struct {
	Timeout             config.Duration `toml:"timeout"`
	IdleConnTimeout     config.Duration `toml:"idle_conn_timeout"`
	MaxIdleConns        int             `toml:"max_idle_conn"`
	MaxIdleConnsPerHost int             `toml:"max_idle_conn_per_host"`

	proxy.HTTPProxy
	tls.ClientConfig
	oauthConfig.OAuth2Config
	cookie.CookieAuthConfig
}

func (h *HTTPClientConfig) CreateClient(ctx context.Context, log telegraf.Logger) (*http.Client, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to set TLS config: %w", err)
	}

	prox, err := h.HTTPProxy.Proxy()
	if err != nil {
		return nil, fmt.Errorf("failed to set proxy: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig:     tlsCfg,
		Proxy:               prox,
		IdleConnTimeout:     time.Duration(h.IdleConnTimeout),
		MaxIdleConns:        h.MaxIdleConns,
		MaxIdleConnsPerHost: h.MaxIdleConnsPerHost,
	}

	timeout := h.Timeout
	if timeout == 0 {
		timeout = config.Duration(time.Second * 5)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout),
	}

	client, err = h.OAuth2Config.CreateOauth2Client(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 client: %w", err)
	}

	if h.CookieAuthConfig.URL != "" {
		if err := h.CookieAuthConfig.Start(client, log, clock.New()); err != nil {
			return nil, err
		}
	}

	return client, nil
}
