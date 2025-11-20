package logql

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

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
	org      string
	timeout  time.Duration
	cfg      common_http.TransportConfig

	client *http.Client
}

func (c *client) init() error {
	// Create a round-tripper suitable for the given configuration based on the
	// http-client transport...
	transport, err := c.cfg.CreateTransport()
	if err != nil {
		return fmt.Errorf("creating transport failed: %w", err)
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

	// Create HTTP client for API requests
	c.client = &http.Client{Transport: rt, Timeout: c.timeout}

	return nil
}

func (c *client) ready(ctx context.Context) (bool, string, error) {
	// Construct the URL
	u, err := url.Parse(c.url)
	if err != nil {
		return false, "", fmt.Errorf("parsing URL %q failed: %w", c.url, err)
	}
	u = u.JoinPath("ready")

	// Issue the request and check the returned status
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request failed: %w", err)
	}
	if c.org != "" {
		req.Header.Set("X-Scope-OrgID", c.org)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the body and if the query was successful
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("reading response body failed: %w", err)
	}

	return resp.StatusCode == 200, string(body), nil
}

func (c *client) execute(ctx context.Context, u string) (interface{}, error) {
	// Prepare the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}
	if c.org != "" {
		req.Header.Set("X-Scope-OrgID", c.org)
	}
	req.Header.Add("Content-Type", "application/json")

	// Execute the query
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the body and if the query was successful
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf(
			"querying failed with status %d (%s): %s",
			resp.StatusCode, http.StatusText(resp.StatusCode), string(body),
		)
	}

	// Parse the response
	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("decoding query response failed: %w", err)
	}

	return r.parse()
}

func (c *client) close() {
	if c.client != nil {
		c.client.CloseIdleConnections()
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
