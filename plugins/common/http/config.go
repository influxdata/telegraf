package httpconfig

import (
	"context"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/cookie"
	oauthConfig "github.com/influxdata/telegraf/plugins/common/oauth"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
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
		return nil, err
	}

	prox, err := h.HTTPProxy.Proxy()
	if err != nil {
		return nil, err
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

	client = h.OAuth2Config.CreateOauth2Client(ctx, client)

	if h.CookieAuthConfig.URL != "" {
		if err := h.CookieAuthConfig.Start(client, log, clock.New()); err != nil {
			return nil, err
		}
	}

	return client, nil
}
