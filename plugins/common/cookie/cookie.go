package cookie

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	clockutil "github.com/benbjohnson/clock"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

type CookieAuthConfig struct {
	URL    string `toml:"cookie_auth_url"`
	Method string `toml:"cookie_auth_method"`

	Headers map[string]string `toml:"cookie_auth_headers"`

	// HTTP Basic Auth Credentials
	Username string `toml:"cookie_auth_username"`
	Password string `toml:"cookie_auth_password"`

	Body    string          `toml:"cookie_auth_body"`
	Renewal config.Duration `toml:"cookie_auth_renewal"`

	client *http.Client
	wg     sync.WaitGroup
}

func (c *CookieAuthConfig) Start(client *http.Client, log telegraf.Logger, clock clockutil.Clock) (err error) {
	if err = c.initializeClient(client); err != nil {
		return err
	}

	// continual auth renewal if set
	if c.Renewal > 0 {
		ticker := clock.Ticker(time.Duration(c.Renewal))
		// this context is used in the tests only, it is to cancel the goroutine
		go c.authRenewal(context.Background(), ticker, log)
	}

	return nil
}

func (c *CookieAuthConfig) initializeClient(client *http.Client) (err error) {
	c.client = client

	if c.Method == "" {
		c.Method = http.MethodPost
	}

	// add cookie jar to HTTP client
	if c.client.Jar, err = cookiejar.New(nil); err != nil {
		return err
	}

	return c.auth()
}

func (c *CookieAuthConfig) authRenewal(ctx context.Context, ticker *clockutil.Ticker, log telegraf.Logger) {
	for {
		select {
		case <-ctx.Done():
			c.wg.Done()
			return
		case <-ticker.C:
			if err := c.auth(); err != nil && log != nil {
				log.Errorf("renewal failed for %q: %v", c.URL, err)
			}
		}
	}
}

func (c *CookieAuthConfig) auth() error {
	var body io.ReadCloser
	if c.Body != "" {
		body = io.NopCloser(strings.NewReader(c.Body))
		defer body.Close()
	}

	req, err := http.NewRequest(c.Method, c.URL, body)
	if err != nil {
		return err
	}

	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	for k, v := range c.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		} else {
			req.Header.Add(k, v)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cookie auth renewal received status code: %v (%v)",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
		)
	}

	return nil
}
