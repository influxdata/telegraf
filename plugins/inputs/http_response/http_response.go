package http_response

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	Address             string
	HTTPProxy           string `toml:"http_proxy"`
	Body                string
	Method              string
	ResponseTimeout     internal.Duration
	Headers             map[string]string
	FollowRedirects     bool
	ResponseStringMatch string
	tls.ClientConfig

	compiledStringMatch *regexp.Regexp
	client              *http.Client
}

// Description returns the plugin Description
func (h *HTTPResponse) Description() string {
	return "HTTP/HTTPS request given an address a method and a timeout"
}

var sampleConfig = `
  ## Server address (default http://localhost)
  # address = "http://localhost"

  ## Set http_proxy (telegraf uses the system wide proxy settings if it's is not set)
  # http_proxy = "http://localhost:8888"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## HTTP Request Method
  # method = "GET"

  ## Whether to follow redirects from the server (defaults to false)
  # follow_redirects = false

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

// CreateHttpClient creates an http client which will timeout at the specified
// timeout period and can follow redirects if specified
func (h *HTTPResponse) createHttpClient() (*http.Client, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:             getProxyFunc(h.HTTPProxy),
			DisableKeepAlives: true,
			TLSClientConfig:   tlsCfg,
		},
		Timeout: h.ResponseTimeout.Duration,
	}

	if h.FollowRedirects == false {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return ErrRedirectAttempted
		}
	}
	return client, nil
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
func (h *HTTPResponse) httpGather() (map[string]interface{}, map[string]string, error) {
	// Prepare fields and tags
	fields := make(map[string]interface{})
	tags := map[string]string{"server": h.Address, "method": h.Method}

	var body io.Reader
	if h.Body != "" {
		body = strings.NewReader(h.Body)
	}
	request, err := http.NewRequest(h.Method, h.Address, body)
	if err != nil {
		return nil, nil, err
	}

	for key, val := range h.Headers {
		request.Header.Add(key, val)
		if key == "Host" {
			request.Host = val
		}
	}

	// Start Timer
	start := time.Now()
	resp, err := h.client.Do(request)
	response_time := time.Since(start).Seconds()

	// If an error in returned, it means we are dealing with a network error, as
	// HTTP error codes do not generate errors in the net/http library
	if err != nil {
		// Log error
		log.Printf("D! Network error while polling %s: %s", h.Address, err.Error())

		// Get error details
		netErr := setError(err, fields, tags)

		// If recognize the returnded error, get out
		if netErr != nil {
			return fields, tags, nil
		}

		// Any error not recognized by `set_error` is considered a "connection_failed"
		setResult("connection_failed", fields, tags)

		// If the error is a redirect we continue processing and log the HTTP code
		urlError, isUrlError := err.(*url.Error)
		if !h.FollowRedirects && isUrlError && urlError.Err == ErrRedirectAttempted {
			err = nil
		} else {
			// If the error isn't a timeout or a redirect stop
			// processing the request
			return fields, tags, nil
		}
	}

	if _, ok := fields["response_time"]; !ok {
		fields["response_time"] = response_time
	}

	// This function closes the response body, as
	// required by the net/http library
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	// Set log the HTTP response code
	tags["status_code"] = strconv.Itoa(resp.StatusCode)
	fields["http_response_code"] = resp.StatusCode

	// Check the response for a regex match.
	if h.ResponseStringMatch != "" {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("D! Failed to read body of HTTP Response : %s", err)
			setResult("body_read_error", fields, tags)
			fields["response_string_match"] = 0
			return fields, tags, nil
		}

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
	if h.Address == "" {
		h.Address = "http://localhost"
	}
	addr, err := url.Parse(h.Address)
	if err != nil {
		return err
	}
	if addr.Scheme != "http" && addr.Scheme != "https" {
		return errors.New("Only http and https are supported")
	}

	// Prepare data
	var fields map[string]interface{}
	var tags map[string]string

	if h.client == nil {
		client, err := h.createHttpClient()
		if err != nil {
			return err
		}
		h.client = client
	}

	// Gather data
	fields, tags, err = h.httpGather()
	if err != nil {
		return err
	}

	// Add metrics
	acc.AddFields("http_response", fields, tags)
	return nil
}

func init() {
	inputs.Add("http_response", func() telegraf.Input {
		return &HTTPResponse{}
	})
}
