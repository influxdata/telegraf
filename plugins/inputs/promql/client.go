package promql

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promcfg "github.com/prometheus/common/config"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
)

type client struct {
	url      string
	username config.Secret
	password config.Secret
	token    config.Secret
	cfg      common_http.TransportConfig

	client api.Client
	apiv1.API
}

func (c *client) init() (*client, error) {
	// Create a round-tripper suitable for the given configuration based on the
	// http-client transport...
	transport, err := c.cfg.CreateTransport()
	if err != nil {
		return nil, fmt.Errorf("creating transport failed: %w", err)
	}
	rt := promcfg.NewUserAgentRoundTripper(internal.ProductToken(), transport)
	if !c.username.Empty() {
		rt = promcfg.NewBasicAuthRoundTripper(
			&secretReader{"username", &c.username},
			&secretReader{"password", &c.password},
			rt,
		)
	} else if !c.token.Empty() {
		rt = promcfg.NewAuthorizationCredentialsRoundTripper(
			"Bearer",
			&secretReader{"token", &c.token},
			rt,
		)
	}

	// Create API client
	apiClient, err := api.NewClient(api.Config{
		Address:      c.url,
		RoundTripper: rt,
	})
	if err != nil {
		return nil, fmt.Errorf("creating API client failed: %w", err)
	}
	c.client = apiClient
	c.API = apiv1.NewAPI(c.client)

	return c, nil
}

func (c *client) close() {
	if c.client != nil {
		if c, ok := c.client.(api.CloseIdler); ok {
			c.CloseIdleConnections()
		}
	}
}

// Wrapper for reading secrets from Prometheus API client
type secretReader struct {
	desc   string
	secret *config.Secret
}

// Fetch implements the Prometheus secret-reader API
func (r *secretReader) Fetch(context.Context) (string, error) {
	raw, err := r.secret.Get()
	if err != nil {
		return "", fmt.Errorf("getting %s failed: %w", r.desc, err)
	}
	s := raw.String()
	raw.Destroy()

	return s, nil
}

// Description implements the Prometheus secret-reader API
func (r *secretReader) Description() string {
	return r.desc
}

// Immutable implements the Prometheus secret-reader API
func (*secretReader) Immutable() bool {
	return true
}
