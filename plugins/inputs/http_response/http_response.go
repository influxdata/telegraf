package http_response

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HttpResponses struct
type HttpResponse struct {
	Address         string
	Method          string
	ResponseTimeout int
}

func (_ *HttpResponse) Description() string {
	return "HTTP/HTTPS request given an address a method and a timeout"
}

var sampleConfig = `
  ## Server address (default http://localhost)
  address = "http://github.com:80"
  ## Set response_timeout (default 1 seconds)
  response_timeout = 1
  ## HTTP Method
  method = "GET"
`

func (_ *HttpResponse) SampleConfig() string {
	return sampleConfig
}

func (h *HttpResponse) HttpGather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})

	client := &http.Client{
		Timeout: time.Second * time.Duration(h.ResponseTimeout),
	}
	request, err := http.NewRequest(h.Method, h.Address, nil)
	if err != nil {
		return nil, err
	}
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

func (c *HttpResponse) Gather(acc telegraf.Accumulator) error {
	// Set default values
	if c.ResponseTimeout < 1 {
		c.ResponseTimeout = 1
	}
	// Check send and expected string
	if c.Method == "" {
		c.Method = "GET"
	}
	if c.Address == "" {
		c.Address = "http://localhost"
	}
	addr, err := url.Parse(c.Address)
	if err != nil {
		return err
	}
	if addr.Scheme != "http" && addr.Scheme != "https" {
		return errors.New("Only http and https are supported")
	}
	// Prepare data
	tags := map[string]string{"server": c.Address, "method": c.Method}
	var fields map[string]interface{}
	// Gather data
	fields, err = c.HttpGather()
	if err != nil {
		return err
	}
	// Add metrics
	acc.AddFields("http_response", fields, tags)
	return nil
}

func init() {
	inputs.Add("http_response", func() telegraf.Input {
		return &HttpResponse{}
	})
}
