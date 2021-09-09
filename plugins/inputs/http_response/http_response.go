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
	"unicode/utf8"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// defaultResponseBodyMaxSize is the default maximum response body size, in bytes.
	// if the response body is over this size, we will raise a body_read_error.
	defaultResponseBodyMaxSize = 32 * 1024 * 1024
)

// HTTPResponse struct
type HTTPResponse struct {
	Address         string   // deprecated in 1.12
	URLs            []string `toml:"urls"`
	HTTPProxy       string   `toml:"http_proxy"`
	Body            string
	Method          string
	ResponseTimeout config.Duration
	HTTPHeaderTags  map[string]string `toml:"http_header_tags"`
	Headers         map[string]string
	FollowRedirects bool
	// Absolute path to file with Bearer token
	BearerToken         string      `toml:"bearer_token"`
	ResponseBodyField   string      `toml:"response_body_field"`
	ResponseBodyMaxSize config.Size `toml:"response_body_max_size"`
	ResponseStringMatch string
	ResponseStatusCode  int
	Interface           string
	// HTTP Basic Auth Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`
	tls.ClientConfig

	Log telegraf.Logger

	compiledStringMatch *regexp.Regexp
	client              httpClient
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
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

  ## Optional name of the field that will contain the body of the response.
  ## By default it is set to an empty String indicating that the body's content won't be added
  # response_body_field = ''

  ## Maximum allowed HTTP response body size in bytes.
  ## 0 means to use the default of 32MiB.
  ## If the response body size exceeds this limit a "body_read_error" will be raised
  # response_body_max_size = "32MiB"

  ## Optional substring or regex match in body of the response (case sensitive)
  # response_string_match = "\"service_status\": \"up\""
  # response_string_match = "ok"
  # response_string_match = "\".*_status\".?:.?\"up\""

  ## Expected response status code.
  ## The status code of the response is compared to this value. If they match, the field
  ## "response_status_code_match" will be 1, otherwise it will be 0. If the
  ## expected status code is 0, the check is disabled and the field won't be added.
  # response_status_code = 0

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Request Headers (all values must be strings)
  # [inputs.http_response.headers]
  #   Host = "github.com"

  ## Optional setting to map response http headers into tags
  ## If the http header is not present on the request, no corresponding tag will be added
  ## If multiple instances of the http header are present, only the first value will be used
  # http_header_tags = {"HTTP_HEADER" = "TAG_NAME"}

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
func getProxyFunc(httpProxy string) func(*http.Request) (*url.URL, error) {
	if httpProxy == "" {
		return http.ProxyFromEnvironment
	}
	proxyURL, err := url.Parse(httpProxy)
	if err != nil {
		return func(_ *http.Request) (*url.URL, error) {
			return nil, errors.New("bad proxy: " + err.Error())
		}
	}
	return func(r *http.Request) (*url.URL, error) {
		return proxyURL, nil
	}
}

// createHTTPClient creates an http client which will timeout at the specified
// timeout period and can follow redirects if specified
func (h *HTTPResponse) createHTTPClient() (*http.Client, error) {
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
		Timeout: time.Duration(h.ResponseTimeout),
	}

	if !h.FollowRedirects {
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

func setResult(resultString string, fields map[string]interface{}, tags map[string]string) {
	resultCodes := map[string]int{
		"success":                       0,
		"response_string_mismatch":      1,
		"body_read_error":               2,
		"connection_failed":             3,
		"timeout":                       4,
		"dns_error":                     5,
		"response_status_code_mismatch": 6,
	}

	tags["result"] = resultString
	fields["result_type"] = resultString
	fields["result_code"] = resultCodes[resultString]
}

func setError(err error, fields map[string]interface{}, tags map[string]string) error {
	if timeoutError, ok := err.(net.Error); ok && timeoutError.Timeout() {
		setResult("timeout", fields, tags)
		return timeoutError
	}

	urlErr, isURLErr := err.(*url.Error)
	if !isURLErr {
		return nil
	}

	opErr, isNetErr := (urlErr.Err).(*net.OpError)
	if isNetErr {
		switch e := (opErr.Err).(type) {
		case *net.DNSError:
			setResult("dns_error", fields, tags)
			return e
		case *net.ParseError:
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
	responseTime := time.Since(start).Seconds()

	// If an error in returned, it means we are dealing with a network error, as
	// HTTP error codes do not generate errors in the net/http library
	if err != nil {
		// Log error
		h.Log.Debugf("Network error while polling %s: %s", u, err.Error())

		// Get error details
		if setError(err, fields, tags) == nil {
			// Any error not recognized by `set_error` is considered a "connection_failed"
			setResult("connection_failed", fields, tags)
		}

		return fields, tags, nil
	}

	if _, ok := fields["response_time"]; !ok {
		fields["response_time"] = responseTime
	}

	// This function closes the response body, as
	// required by the net/http library
	defer resp.Body.Close()

	// Add the response headers
	for headerName, tag := range h.HTTPHeaderTags {
		headerValues, foundHeader := resp.Header[headerName]
		if foundHeader && len(headerValues) > 0 {
			tags[tag] = headerValues[0]
		}
	}

	// Set log the HTTP response code
	tags["status_code"] = strconv.Itoa(resp.StatusCode)
	fields["http_response_code"] = resp.StatusCode

	if h.ResponseBodyMaxSize == 0 {
		h.ResponseBodyMaxSize = config.Size(defaultResponseBodyMaxSize)
	}
	bodyBytes, err := ioutil.ReadAll(io.LimitReader(resp.Body, int64(h.ResponseBodyMaxSize)+1))
	// Check first if the response body size exceeds the limit.
	if err == nil && int64(len(bodyBytes)) > int64(h.ResponseBodyMaxSize) {
		h.setBodyReadError("The body of the HTTP Response is too large", bodyBytes, fields, tags)
		return fields, tags, nil
	} else if err != nil {
		h.setBodyReadError(fmt.Sprintf("Failed to read body of HTTP Response : %s", err.Error()), bodyBytes, fields, tags)
		return fields, tags, nil
	}

	// Add the body of the response if expected
	if len(h.ResponseBodyField) > 0 {
		// Check that the content of response contains only valid utf-8 characters.
		if !utf8.Valid(bodyBytes) {
			h.setBodyReadError("The body of the HTTP Response is not a valid utf-8 string", bodyBytes, fields, tags)
			return fields, tags, nil
		}
		fields[h.ResponseBodyField] = string(bodyBytes)
	}
	fields["content_length"] = len(bodyBytes)

	var success = true

	// Check the response for a regex
	if h.ResponseStringMatch != "" {
		if h.compiledStringMatch.Match(bodyBytes) {
			fields["response_string_match"] = 1
		} else {
			success = false
			setResult("response_string_mismatch", fields, tags)
			fields["response_string_match"] = 0
		}
	}

	// Check the response status code
	if h.ResponseStatusCode > 0 {
		if resp.StatusCode == h.ResponseStatusCode {
			fields["response_status_code_match"] = 1
		} else {
			success = false
			setResult("response_status_code_mismatch", fields, tags)
			fields["response_status_code_match"] = 0
		}
	}

	if success {
		setResult("success", fields, tags)
	}

	return fields, tags, nil
}

// Set result in case of a body read error
func (h *HTTPResponse) setBodyReadError(errorMsg string, bodyBytes []byte, fields map[string]interface{}, tags map[string]string) {
	h.Log.Debugf(errorMsg)
	setResult("body_read_error", fields, tags)
	fields["content_length"] = len(bodyBytes)
	if h.ResponseStringMatch != "" {
		fields["response_string_match"] = 0
	}
}

// Gather gets all metric fields and tags and returns any errors it encounters
func (h *HTTPResponse) Gather(acc telegraf.Accumulator) error {
	// Compile the body regex if it exist
	if h.compiledStringMatch == nil {
		var err error
		h.compiledStringMatch, err = regexp.Compile(h.ResponseStringMatch)
		if err != nil {
			return fmt.Errorf("failed to compile regular expression %s : %s", h.ResponseStringMatch, err)
		}
	}

	// Set default values
	if h.ResponseTimeout < config.Duration(time.Second) {
		h.ResponseTimeout = config.Duration(time.Second * 5)
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
		client, err := h.createHTTPClient()
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
			acc.AddError(errors.New("only http and https are supported"))
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
