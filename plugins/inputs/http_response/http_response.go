package http_response

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HTTPResponse struct
type HTTPResponse struct {
	Address         string
	Body            string
	Method          string
	ResponseTimeout internal.Duration
	Headers         map[string]string
	FollowRedirects bool
}

// Description returns the plugin Description
func (h *HTTPResponse) Description() string {
	return "HTTP/HTTPS request given an address a method and a timeout"
}

var sampleConfig = `
  ## Server address (default http://localhost)
  address = "http://github.com"
  ## Set response_timeout (default 5 seconds)
  response_timeout = "5s"
  ## HTTP Request Method
  method = "GET"
  ## Whether to follow redirects from the server (defaults to false)
  follow_redirects = true
  ## HTTP Request Headers (all values must be strings)
  # [inputs.http_response.headers]
  #   Host = "github.com"
  ## Optional HTTP Request Body
  # body = '''
  # {'fake':'data'}
  # '''
`

// SampleConfig returns the plugin SampleConfig
func (h *HTTPResponse) SampleConfig() string {
	return sampleConfig
}

// ErrRedirectAttempted indicates that a redirect occurred
var ErrRedirectAttempted = errors.New("redirect")

// CreateHttpClient creates an http client which will timeout at the specified
// timeout period and can follow redirects if specified
func CreateHttpClient(followRedirects bool, ResponseTimeout time.Duration) *http.Client {
	client := &http.Client{
		Timeout: ResponseTimeout,
	}

	if followRedirects == false {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return ErrRedirectAttempted
		}
	}
	return client
}

// HTTPGather gathers all fields and returns any errors it encounters
func (h *HTTPResponse) HTTPGather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})

	client := CreateHttpClient(h.FollowRedirects, h.ResponseTimeout.Duration)

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
	resp, err := client.Do(request)
	if err != nil {
		if h.FollowRedirects {
			return nil, err
		}
		if urlError, ok := err.(*url.Error); ok &&
			urlError.Err == ErrRedirectAttempted {
			err = nil
		} else {
			return nil, err
		}
	}
	fields["response_time"] = time.Since(start).Seconds()
	fields["http_response_code"] = resp.StatusCode
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
	// Gather data
	fields, err = h.HTTPGather()
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
