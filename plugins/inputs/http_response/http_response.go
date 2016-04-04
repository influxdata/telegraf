package http_response

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HTTPResponse struct
type HTTPResponse struct {
	Address         string
	Body            string
	Method          string
	ResponseTimeout int
	Headers         string
	FollowRedirects bool
}

// Description returns the plugin Description
func (h *HTTPResponse) Description() string {
	return "HTTP/HTTPS request given an address a method and a timeout"
}

var sampleConfig = `
  ## Server address (default http://localhost)
  address = "http://github.com"
  ## Set response_timeout (default 10 seconds)
  response_timeout = 10
  ## HTTP Request Method
  method = "GET"
  ## HTTP Request Headers
  headers = '''
  Host: github.com
  '''
  ## Whether to follow redirects from the server (defaults to false)
  follow_redirects = true
  ## Optional HTTP Request Body
  body = '''
  {'fake':'data'}
  '''
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
		Timeout: time.Second * ResponseTimeout,
	}

	if followRedirects == false {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return ErrRedirectAttempted
		}
	}
	return client
}

// ParseHeaders takes a string of newline seperated http headers and returns a
// http.Header object. An error is returned if the headers cannot be parsed.
func ParseHeaders(headers string) (http.Header, error) {
	headers = strings.TrimSpace(headers) + "\n\n"
	reader := bufio.NewReader(strings.NewReader(headers))
	tp := textproto.NewReader(reader)
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	return http.Header(mimeHeader), nil
}

// HTTPGather gathers all fields and returns any errors it encounters
func (h *HTTPResponse) HTTPGather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})

	client := CreateHttpClient(h.FollowRedirects, time.Duration(h.ResponseTimeout))

	var body io.Reader
	if h.Body != "" {
		body = strings.NewReader(h.Body)
	}
	request, err := http.NewRequest(h.Method, h.Address, body)
	if err != nil {
		return nil, err
	}
	request.Header, err = ParseHeaders(h.Headers)
	if err != nil {
		return nil, err
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
	if h.ResponseTimeout < 1 {
		h.ResponseTimeout = 10
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
