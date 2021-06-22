package cookie

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/sirupsen/logrus"
)

type CookieAuthConfig struct {
	URL    string `toml:"cookie_auth_url"`
	Method string `toml:"cookie_auth_method"`

	// HTTP Basic Auth Credentials
	Username string `toml:"cookie_auth_username"`
	Password string `toml:"cookie_auth_password"`

	Body    string          `toml:"cookie_auth_body"`
	Renewal config.Duration `toml:"cookie_auth_renewal"`

	client *http.Client
}

func (c *CookieAuthConfig) Start(client *http.Client) (err error) {
	c.client = client

	if c.Method == "" {
		c.Method = http.MethodPost
	}

	// add cookie jar to HTTP client
	if c.client.Jar, err = cookiejar.New(nil); err != nil {
		return err
	}

	// continual auth renewal if set
	if c.Renewal > 0 {
		ticker := time.NewTicker(time.Duration(c.Renewal))
		go func() {
			for range ticker.C {
				if err := c.auth(); err != nil {
					logrus.WithError(err).Error("cookie auth renewal failure")
				}
			}
		}()
	}

	// initial auth will immediately error out the Init() if auth fails
	return c.auth()
}

func (c *CookieAuthConfig) auth() error {
	var body io.ReadCloser
	if c.Body != "" {
		body = ioutil.NopCloser(strings.NewReader(c.Body))
		defer body.Close()
	}

	req, err := http.NewRequest(c.Method, c.URL, body)
	if err != nil {
		return err
	}

	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err = io.Copy(ioutil.Discard, resp.Body); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response code: %v", resp.StatusCode)
	}

	return nil
}
