package httpconfig

import (
	"context"
	"net/http"
	"time"

	"github.com/influxdata/telegraf/config"
	oauthConfig "github.com/influxdata/telegraf/plugins/common/oauth"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

// Common HTTP client struct.
type HTTPClientConfig struct {
	Timeout         config.Duration `toml:"timeout"`
	IdleConnTimeout config.Duration `toml:"idle_conn_timeout"`

	proxy.HTTPProxy
	tls.ClientConfig
	oauthConfig.OAuth2Config
}

func (h *HTTPClientConfig) CreateClient(ctx context.Context) (*http.Client, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	prox, err := h.HTTPProxy.Proxy()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
		Proxy:           prox,
		IdleConnTimeout: time.Duration(h.IdleConnTimeout),
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

	return client, nil
}
