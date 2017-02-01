package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"net"
	"net/http"
	"strings"
	"time"
)

var sampleConfig = `
  ## It requires a url name.
  ## Will be transmitted telegraf metrics to the HTTP Server using the below URL.
  url = "http://127.0.0.1:8080/metric"
  ## http_headers option can add a custom header to the request.
  ## The value is written as a delimiter(:).
  ## Content-Type is required http header in http plugin.
  ## so content-type of HTTP specification (plain/text, application/json, etc...) must be filled out.
  http_headers = [ "Content-Type:application/json" ]
  ## With this HTTP status code, the http plugin checks that the HTTP request is completed normally.
  ## As a result, any status code that is not a specified status code is considered to be an error condition and processed.
  expected_status_codes = [ 200, 204 ]
  ## Configure TLS handshake timeout. Default : 10
  tls_handshake_timeout = 10
  ## Configure response header timeout in seconds. Default : 3
  response_header_timeout = 3
  ## Configure dial timeout in seconds. Default : 3
  dial_timeout = 3
  ## Configure HTTP Keep-Alive. Default : 0
  keepalive = 0
  ## Configure HTTP expect continue timeout in seconds. Default : 0
  expect_continue_timeout = 3
  ## Configure idle connection timeout in seconds. Default : 0
  idle_conn_timeout = 3

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

const (
	POST = "POST"

	DEFAULT_RESPONSE_HEADER_TIMEOUT = 3
	DEFAULT_DIAL_TIME_OUT           = 3
	DEFAULT_KEEP_ALIVE              = 3
)

type Http struct {
	// http required option
	URL                 string   `toml:"url"`
	HttpHeaders         []string `toml:"http_headers"`
	ExpectedStatusCodes []int    `toml:"expected_status_codes"`

	// Option with http default value
	ResponseHeaderTimeout int `toml:"response_header_timeout"`
	DialTimeOut           int `toml:"dial_timeout"`
	KeepAlive             int `toml:"keepalive"`

	client                http.Client
	serializer            serializers.Serializer
	expectedStatusCodeMap map[int]bool

	// Context for request cancel of client
	cancelContext context.Context
	cancel        context.CancelFunc
}

func (h *Http) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

// Connect to the Output
func (h *Http) Connect() error {
	h.client = http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   time.Duration(h.DialTimeOut) * time.Second,
				KeepAlive: time.Duration(h.KeepAlive) * time.Second,
			}).Dial,
			ResponseHeaderTimeout: time.Duration(h.ResponseHeaderTimeout) * time.Second,
		},
	}

	h.cancelContext, h.cancel = context.WithCancel(context.TODO())

	return nil
}

// Close the Client Connection using Context.
func (h Http) Close() error {
	h.cancel()

	return nil
}

// Description A plugin that can transmit metrics over HTTP
func (h Http) Description() string {
	return "A plugin that can transmit metrics over HTTP"
}

// SampleConfig provides sample example for developer
func (h Http) SampleConfig() string {
	return sampleConfig
}

// Writes metrics over HTTP POST
func (h Http) Write(metrics []telegraf.Metric) error {
	if err := validate(h); err != nil {
		return err
	}

	for _, metric := range metrics {
		buf, err := h.serializer.Serialize(metric)

		if err != nil {
			return fmt.Errorf("E! Error serializing some metrics: %s", err.Error())
		}

		response, err := h.write(buf)

		if err := h.isOk(response, err); err != nil {
			return err
		}

		defer response.Body.Close()
	}

	return nil
}

func (h Http) isOk(response *http.Response, err error) error {
	if response == nil || err != nil {
		return fmt.Errorf("E! %s request failed! %s.", h.URL, err.Error())
	}

	if !h.isExpectedStatusCode(response.StatusCode) {
		return fmt.Errorf("E! %s response is unexpected status code : %d.", h.URL, response.StatusCode)
	}

	return nil
}

func (h Http) isExpectedStatusCode(responseStatusCode int) bool {
	if h.expectedStatusCodeMap == nil {
		h.expectedStatusCodeMap = make(map[int]bool)

		for _, expectedStatusCode := range h.ExpectedStatusCodes {
			h.expectedStatusCodeMap[expectedStatusCode] = true
		}
	}

	if h.expectedStatusCodeMap[responseStatusCode] {
		return true
	}

	return false
}

// required option validate
func validate(h Http) error {
	if h.URL == "" || len(h.HttpHeaders) == 0 || len(h.ExpectedStatusCodes) == 0 {
		return errors.New("E! Http ouput plugin is not working. Because your configuration omits the required option. Please check url, http_headers, expected_status_codes is empty!")
	}

	return nil
}

func (h Http)write(buf []byte) (*http.Response, error) {
	req, err := http.NewRequest(POST, h.URL, bytes.NewBuffer(buf))

	for _, httpHeader := range h.HttpHeaders {
		keyAndValue := strings.Split(httpHeader, ":")
		req.Header.Set(keyAndValue[0], keyAndValue[1])
	}

	req.Close = false
	req.WithContext(h.cancelContext)

	response, err := h.client.Do(req)

	return response, err
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &Http{
			ResponseHeaderTimeout: DEFAULT_RESPONSE_HEADER_TIMEOUT,
			DialTimeOut:           DEFAULT_DIAL_TIME_OUT,
			KeepAlive:             DEFAULT_KEEP_ALIVE,
		}
	})
}
