package httpconfig

import (
	"context"
	"net/http"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Common HTTP client struct.
type HTTPClientConfig struct {
	// OAuth2 Credentials
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	TokenURL     string   `toml:"token_url"`
	Scopes       []string `toml:"scopes"`

	Timeout config.Duration `toml:"timeout"`

	proxy.HTTPProxy
	tls.ClientConfig
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
	}

	timeout := h.Timeout
	if timeout == 0 {
		timeout = config.Duration(time.Second * 5)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout),
	}

	if h.ClientID != "" && h.ClientSecret != "" && h.TokenURL != "" {
		oauthConfig := clientcredentials.Config{
			ClientID:     h.ClientID,
			ClientSecret: h.ClientSecret,
			TokenURL:     h.TokenURL,
			Scopes:       h.Scopes,
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
		client = oauthConfig.Client(ctx)
	}

	return client, nil
}
