package http_response

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HTTPResponse struct
type HTTPResponse struct {
	Address         string   // deprecated in 1.12
	URLs            []string `toml:"urls"`
	HTTPProxy       string   `toml:"http_proxy"`
	Body            string
	Method          string
	ResponseTimeout internal.Duration
	Headers         map[string]string
	FollowRedirects bool
	// Absolute path to file with Bearer token
	BearerToken         string `toml:"bearer_token"`
	ResponseStringMatch string
	Interface           string
	// HTTP Basic Auth Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`
	tls.ClientConfig

	Log telegraf.Logger

	compiledStringMatch *regexp.Regexp
	client              *http.Client
}

// Description returns the plugin Description
func (h *HTTPResponse) Description() string {
	return "HTTP/HTTPS request given an address a method and a timeout"
}

var sampleConfig = `
  ## Deprecated in 1.12, use 'urls'
  ## Server address (default http://localhost)
  # address = "http://localhost"

  ## List of urls to query.
  # urls = ["http://localhost"]

  ## Set http_proxy (telegraf uses the system wide proxy settings if it's is not set)
  # http_proxy = "http://localhost:8888"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## HTTP Request Method
  # method = "GET"

  ## Whether to follow redirects from the server (defaults to false)
  # follow_redirects = false

  ## Optional file with Bearer token
  ## file content is added as an Authorization header
  # bearer_token = "/path/to/file"

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional HTTP Request Body
  # body = '''
  # {'fake':'data'}
  # '''

  ## Optional substring or regex match in body of the response
  # response_string_match = "\"service_status\": \"up\""
  # response_string_match = "ok"
  # response_string_match = "\".*_status\".?:.?\"up\""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Request Headers (all values must be strings)
  # [inputs.http_response.headers]
  #   Host = "github.com"

  ## Interface to use when dialing an address
  # interface = "eth0"
`

// SampleConfig returns the plugin SampleConfig
func (h *HTTPResponse) SampleConfig() string {
	return sampleConfig
}

// ErrRedirectAttempted indicates that a redirect occurred
var ErrRedirectAttempted = errors.New("redirect")

// Set the proxy. A configured proxy overwrites the system wide proxy.
func getProxyFunc(http_proxy string) func(*http.Request) (*url.URL, error) {
	if http_proxy == "" {
		return http.ProxyFromEnvironment
	}
	proxyURL, err := url.Parse(http_proxy)
	if err != nil {
		return func(_ *http.Request) (*url.URL, error) {
			return nil, errors.New("bad proxy: " + err.Error())
		}
	}
	return func(r *http.Request) (*url.URL, error) {
		return proxyURL, nil
	}
}

// createHttpClient creates an http client which will timeout at the specified
// timeout period and can follow redirects if specified
func (h *HTTPResponse) createHttpClient() (*http.Client, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{}

	if h.Interface != "" {
		dialer.LocalAddr, err = localAddress(h.Interface)
		if err != nil {
			return nil, err
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:             getProxyFunc(h.HTTPProxy),
			DialContext:       dialer.DialContext,
			DisableKeepAlives: true,
			TLSClientConfig:   tlsCfg,
		},
		Timeout: h.ResponseTimeout.Duration,
	}

	if h.FollowRedirects == false {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client, nil
}

func localAddress(interfaceName string) (net.Addr, error) {
	i, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}

	addrs, err := i.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if naddr, ok := addr.(*net.IPNet); ok {
			// leaving port set to zero to let kernel pick
			return &net.TCPAddr{IP: naddr.IP}, nil
		}
	}

	return nil, fmt.Errorf("cannot create local address for interface %q", interfaceName)
}

func setResult(result_string string, fields map[string]interface{}, tags map[string]string) {
	result_codes := map[string]int{
		"success":                  0,
		"response_string_mismatch": 1,
		"body_read_error":          2,
		"connection_failed":        3,
		"timeout":                  4,
		"dns_error":                5,
	}

	tags["result"] = result_string
	fields["result_type"] = result_string
	fields["result_code"] = result_codes[result_string]
}

func setError(err error, fields map[string]interface{}, tags map[string]string) error {
	if timeoutError, ok := err.(net.Error); ok && timeoutError.Timeout() {
		setResult("timeout", fields, tags)
		return timeoutError
	}

	urlErr, isUrlErr := err.(*url.Error)
	if !isUrlErr {
		return nil
	}

	opErr, isNetErr := (urlErr.Err).(*net.OpError)
	if isNetErr {
		switch e := (opErr.Err).(type) {
		case (*net.DNSError):
			setResult("dns_error", fields, tags)
			return e
		case (*net.ParseError):
			// Parse error has to do with parsing of IP addresses, so we
			// group it with address errors
			setResult("address_error", fields, tags)
			return e
		}
	}

	return nil
}

// HTTPGather gathers all fields and returns any errors it encounters
func (h *HTTPResponse) httpGather(u string) (map[string]interface{}, map[string]string, error) {
	// Prepare fields and tags
	fields := make(map[string]interface{})
	tags := map[string]string{"server": u, "method": h.Method}

	var body io.Reader
	if h.Body != "" {
		body = strings.NewReader(h.Body)
	}
	request, err := http.NewRequest(h.Method, u, body)
	if err != nil {
		return nil, nil, err
	}

	if h.BearerToken != "" {
		token, err := ioutil.ReadFile(h.BearerToken)
		if err != nil {
			return nil, nil, err
		}
		bearer := "Bearer " + strings.Trim(string(token), "\n")
		request.Header.Add("Authorization", bearer)
	}

	for key, val := range h.Headers {
		request.Header.Add(key, val)
		if key == "Host" {
			request.Host = val
		}
	}

	if h.Username != "" || h.Password != "" {
		request.SetBasicAuth(h.Username, h.Password)
	}

	// Start Timer
	start := time.Now()
	resp, err := h.client.Do(request)
	response_time := time.Since(start).Seconds()

	// If an error in returned, it means we are dealing with a network error, as
	// HTTP error codes do not generate errors in the net/http library
	if err != nil {
		// Log error
		h.Log.Debugf("Network error while polling %s: %s", u, err.Error())

		// Get error details
		netErr := setError(err, fields, tags)

		// If recognize the returned error, get out
		if netErr != nil {
			return fields, tags, nil
		}

		// Any error not recognized by `set_error` is considered a "connection_failed"
		setResult("connection_failed", fields, tags)
		return fields, tags, nil
	}

	if _, ok := fields["response_time"]; !ok {
		fields["response_time"] = response_time
	}

	// This function closes the response body, as
	// required by the net/http library
	defer resp.Body.Close()

	// Set log the HTTP response code
	tags["status_code"] = strconv.Itoa(resp.StatusCode)
	fields["http_response_code"] = resp.StatusCode

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.Log.Debugf("Failed to read body of HTTP Response : %s", err.Error())
		setResult("body_read_error", fields, tags)
		fields["content_length"] = len(bodyBytes)
		if h.ResponseStringMatch != "" {
			fields["response_string_match"] = 0
		}
		return fields, tags, nil
	}

	fields["content_length"] = len(bodyBytes)

	// Check the response for a regex match.
	if h.ResponseStringMatch != "" {
		if h.compiledStringMatch.Match(bodyBytes) {
			setResult("success", fields, tags)
			fields["response_string_match"] = 1
		} else {
			setResult("response_string_mismatch", fields, tags)
			fields["response_string_match"] = 0
		}
	} else {
		setResult("success", fields, tags)
	}

	return fields, tags, nil
}

// Gather gets all metric fields and tags and returns any errors it encounters
func (h *HTTPResponse) Gather(acc telegraf.Accumulator) error {
	// Compile the body regex if it exist
	if h.compiledStringMatch == nil {
		var err error
		h.compiledStringMatch, err = regexp.Compile(h.ResponseStringMatch)
		if err != nil {
			return fmt.Errorf("Failed to compile regular expression %s : %s", h.ResponseStringMatch, err)
		}
	}

	// Set default values
	if h.ResponseTimeout.Duration < time.Second {
		h.ResponseTimeout.Duration = time.Second * 5
	}
	// Check send and expected string
	if h.Method == "" {
		h.Method = "GET"
	}

	if len(h.URLs) == 0 {
		if h.Address == "" {
			h.URLs = []string{"http://localhost"}
		} else {
			h.Log.Warn("'address' deprecated in telegraf 1.12, please use 'urls'")
			h.URLs = []string{h.Address}
		}
	}

	if h.client == nil {
		client, err := h.createHttpClient()
		if err != nil {
			return err
		}
		h.client = client
	}

	for _, u := range h.URLs {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(err)
			continue
		}

		if addr.Scheme != "http" && addr.Scheme != "https" {
			acc.AddError(errors.New("Only http and https are supported"))
			continue
		}

		// Prepare data
		var fields map[string]interface{}
		var tags map[string]string

		// Gather data
		fields, tags, err = h.httpGather(u)
		if err != nil {
			acc.AddError(err)
			continue
		}

		// Add metrics
		acc.AddFields("http_response", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("http_response", func() telegraf.Input {
		return &HTTPResponse{}
	})
}
