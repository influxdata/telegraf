package http_response

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HTTPResponse struct
type HTTPResponse struct {
	Address             string
	Body                string
	Method              string
	ResponseTimeout     internal.Duration
	Headers             map[string]string
	FollowRedirects     bool
	ResponseStringMatch string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

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

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
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

// CreateHttpClient creates an http client which will timeout at the specified
// timeout period and can follow redirects if specified
func (h *HTTPResponse) createHttpClient() (*http.Client, error) {
	tlsCfg, err := internal.GetTLSConfig(
		h.SSLCert, h.SSLKey, h.SSLCA, h.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
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

// HTTPGather gathers all fields and returns any errors it encounters
func (h *HTTPResponse) httpGather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})

	var body io.Reader
	if h.Body != "" {
		body = strings.NewReader(h.Body)
	}
	request, err := http.NewRequest(h.Method, h.Address, body)
	if err != nil {
		return nil, err
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

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fields["result_type"] = "timeout"
			return fields, nil
		}
		fields["result_type"] = "connection_failed"
		if h.FollowRedirects {
			return fields, nil
		}
		if urlError, ok := err.(*url.Error); ok &&
			urlError.Err == ErrRedirectAttempted {
			err = nil
		} else {
			return fields, nil
		}
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	fields["response_time"] = time.Since(start).Seconds()
	fields["http_response_code"] = resp.StatusCode

	// Check the response for a regex match.
	if h.ResponseStringMatch != "" {

		// Compile once and reuse
		if h.compiledStringMatch == nil {
			h.compiledStringMatch = regexp.MustCompile(h.ResponseStringMatch)
			if err != nil {
				log.Printf("E! Failed to compile regular expression %s : %s", h.ResponseStringMatch, err)
				fields["result_type"] = "response_string_mismatch"
				return fields, nil
			}
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("E! Failed to read body of HTTP Response : %s", err)
			fields["result_type"] = "response_string_mismatch"
			fields["response_string_match"] = 0
			return fields, nil
		}

		if h.compiledStringMatch.Match(bodyBytes) {
			fields["result_type"] = "success"
			fields["response_string_match"] = 1
		} else {
			fields["result_type"] = "response_string_mismatch"
			fields["response_string_match"] = 0
		}
	} else {
		fields["result_type"] = "success"
	}

	return fields, nil
}

// Gather gets all metric fields and tags and returns any errors it encounters
func (h *HTTPResponse) Gather(acc telegraf.Accumulator) error {
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
	tags := map[string]string{"server": h.Address, "method": h.Method}
	var fields map[string]interface{}

	if h.client == nil {
		client, err := h.createHttpClient()
		if err != nil {
			return err
		}
		h.client = client
	}

	// Gather data
	fields, err = h.httpGather()
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
