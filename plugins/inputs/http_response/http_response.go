package http_response

import (
	"bufio"
	"errors"
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
	Method          string
	ResponseTimeout int
	Headers         string
}

// Description returns the plugin Description
func (h *HTTPResponse) Description() string {
	return "HTTP/HTTPS request given an address a method and a timeout"
}

var sampleConfig = `
  ## Server address (default http://localhost)
  address = "https://github.com"
  ## Set response_timeout (default 1 seconds)
  response_timeout = 10
  ## HTTP Method
  method = "GET"
  ## HTTP Request Headers
  headers = '''
  Host: github.com
  '''
`

// SampleConfig returns the plugin SampleConfig
func (h *HTTPResponse) SampleConfig() string {
	return sampleConfig
}

// HTTPGather gathers all fields and returns any errors it encounters
func (h *HTTPResponse) HTTPGather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})

	client := &http.Client{
		Timeout: time.Second * time.Duration(h.ResponseTimeout),
	}
	request, err := http.NewRequest(h.Method, h.Address, nil)
	if err != nil {
		return nil, err
	}
	h.Headers = strings.TrimSpace(h.Headers) + "\n\n"
	reader := bufio.NewReader(strings.NewReader(h.Headers))
	tp := textproto.NewReader(reader)
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	request.Header = http.Header(mimeHeader)
	// Start Timer
	start := time.Now()
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	fields["response_time"] = time.Since(start).Seconds()
	fields["http_response_code"] = resp.StatusCode
	return fields, nil
}

// Gather gets all metric fields and tags and returns any errors it encounters
func (h *HTTPResponse) Gather(acc telegraf.Accumulator) error {
	// Set default values
	if h.ResponseTimeout < 1 {
		h.ResponseTimeout = 1
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
