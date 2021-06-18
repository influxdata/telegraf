package http

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

var (
	defaultCookieAuthRenewal = 5 * time.Minute
	defaultCookieAuthMethod  = http.MethodPost
)

type HTTP struct {
	URLs            []string `toml:"urls"`
	Method          string   `toml:"method"`
	Body            string   `toml:"body"`
	ContentEncoding string   `toml:"content_encoding"`

	Headers map[string]string `toml:"headers"`

	// HTTP Basic Auth Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`

	// Cookie authentication
	CookieAuthURL     string        `toml:"cookie_auth_url"`
	CookieAuthMethod  string        `toml:"cookie_auth_method"`
	CookieAuthBody    string        `toml:"cookie_auth_body"`
	CookieAuthRenewal time.Duration `toml:"cookie_auth_renewal"`

	// Absolute path to file with Bearer token
	BearerToken string `toml:"bearer_token"`

	SuccessStatusCodes []int `toml:"success_status_codes"`

	client *http.Client
	httpconfig.HTTPClientConfig

	// The parser will automatically be set by Telegraf core code because
	// this plugin implements the ParserInput interface (i.e. the SetParser method)
	parser parsers.Parser
}

var sampleConfig = `
  ## One or more URLs from which to read formatted metrics
  urls = [
    "http://localhost/metrics"
  ]

  ## HTTP method
  # method = "GET"

  ## Optional HTTP headers
  # headers = {"X-Special-Header" = "Special-Value"}

  ## Optional file with Bearer token
  ## file content is added as an Authorization header
  # bearer_token = "/path/to/file"

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## HTTP entity-body to send with POST/PUT requests.
  # body = ""

  ## Cookie-based authentication
  # cookie_auth_url = "https://localhost/authMe"
  # cookie_auth_method = "POST" [default: POST]
  # cookie_auth_body = {"username": "user", "password": "pa$$word", "authenticate": "me"}
  # cookie_auth_renewal = 8h [default: 5m]

  ## HTTP Content-Encoding for write request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "identity"

  ## HTTP Proxy support
  # http_proxy_url = ""

  ## OAuth2 Client Credentials Grant
  # client_id = "clientid"
  # client_secret = "secret"
  # token_url = "https://indentityprovider/oauth2/v1/token"
  # scopes = ["urn:opc:idm:__myscopes__"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"

  ## List of success status codes
  # success_status_codes = [200]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"
`

// SampleConfig returns the default configuration of the Input
func (*HTTP) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the Input
func (*HTTP) Description() string {
	return "Read formatted metrics from one or more HTTP endpoints"
}

func (h *HTTP) Init() error {
	ctx := context.Background()
	client, err := h.HTTPClientConfig.CreateClient(ctx)
	if err != nil {
		return err
	}

	h.client = client

	if h.CookieAuthURL != "" {
		// cookie auth defaults
		if h.CookieAuthRenewal == 0 {
			h.CookieAuthRenewal = defaultCookieAuthRenewal
		}
		if h.CookieAuthMethod == "" {
			h.CookieAuthMethod = defaultCookieAuthMethod
		}
		// add cookie jar to HTTP client
		if h.client.Jar, err = cookiejar.New(nil); err != nil {
			return err
		}
		// start auth ticker
		if authErr := h.startCookieAuth(); authErr != nil {
			return authErr
		}
	}

	// Set default as [200]
	if len(h.SuccessStatusCodes) == 0 {
		h.SuccessStatusCodes = []int{200}
	}
	return nil
}

func (h *HTTP) startCookieAuth() error {
	// continual auth ticker
	ticker := time.NewTicker(h.CookieAuthRenewal)
	go func() {
		for range ticker.C {
			_ = h.doCookieAuth()
		}
	}()

	// initial auth will immediately error out the Init() if auth fails
	return h.doCookieAuth()
}

func (h *HTTP) doCookieAuth() error {
	request, err := h.newRequest(h.CookieAuthURL, h.CookieAuthMethod, h.CookieAuthBody)
	if err != nil {
		return err
	}
	defer request.Body.Close()

	errBadRespCode := fmt.Errorf("bad response code")
	resp, respErr := h.client.Do(request)
	if respErr != nil {
		return respErr
	}
	if _, err = io.Copy(ioutil.Discard, resp.Body); err != nil {
		return err
	}
	if err = resp.Body.Close(); err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %v", errBadRespCode, resp.StatusCode)
	}
	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval"
func (h *HTTP) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, u := range h.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := h.gatherURL(acc, url); err != nil {
				acc.AddError(fmt.Errorf("[url=%s]: %s", url, err))
			}
		}(u)
	}

	wg.Wait()

	return nil
}

// SetParser takes the data_format from the config and finds the right parser for that format
func (h *HTTP) SetParser(parser parsers.Parser) {
	h.parser = parser
}

// Gathers data from a particular URL
// Parameters:
//     acc    : The telegraf Accumulator to use
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (h *HTTP) gatherURL(
	acc telegraf.Accumulator,
	url string,
) error {
	request, err := h.newRequest(url, h.Method, h.Body)
	if err != nil {
		return err
	}
	defer request.Body.Close()

	resp, err := h.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseHasSuccessCode := false
	for _, statusCode := range h.SuccessStatusCodes {
		if resp.StatusCode == statusCode {
			responseHasSuccessCode = true
			break
		}
	}

	if !responseHasSuccessCode {
		return fmt.Errorf("received status code %d (%s), expected any value out of %v",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			h.SuccessStatusCodes)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	metrics, err := h.parser.Parse(b)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		if !metric.HasTag("url") {
			metric.AddTag("url", url)
		}
		acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}

	return nil
}

func (h *HTTP) newRequest(url, method, body string) (*http.Request, error) {
	bodyRC, err := makeRequestBodyReader(h.ContentEncoding, body)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(method, url, bodyRC)
	if err != nil {
		return nil, err
	}

	if h.BearerToken != "" {
		token, err := ioutil.ReadFile(h.BearerToken)
		if err != nil {
			return nil, err
		}
		bearer := "Bearer " + strings.Trim(string(token), "\n")
		request.Header.Set("Authorization", bearer)
	}

	if h.ContentEncoding == "gzip" {
		request.Header.Set("Content-Encoding", "gzip")
	}

	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			request.Host = v
		} else {
			request.Header.Add(k, v)
		}
	}

	if h.Username != "" || h.Password != "" {
		request.SetBasicAuth(h.Username, h.Password)
	}
	return request, nil
}

func makeRequestBodyReader(contentEncoding, body string) (io.ReadCloser, error) {
	reader := strings.NewReader(body)
	if contentEncoding == "gzip" {
		rc, err := internal.CompressWithGzip(reader)
		if err != nil {
			return nil, err
		}
		return rc, nil
	}
	return ioutil.NopCloser(reader), nil
}

func init() {
	inputs.Add("http", func() telegraf.Input {
		return &HTTP{
			Method: "GET",
		}
	})
}
