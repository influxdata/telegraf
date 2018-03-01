package http

import (
	"bytes"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"net/http"
	"strings"
	"time"
)

var sampleConfig = `
  ## It requires a url name.
  ## Will be transmitted telegraf metrics to the HTTP Server using the below URL.
  ## Note that not support the HTTPS.
  url = "http://127.0.0.1:8080/metric"
  ## http_headers option can add a custom header to the request.
  ## The value is written as a delimiter(:).
  ## Content-Type is required http header in http plugin.
  ## so content-type of HTTP specification (plain/text, application/json, etc...) must be filled out.
  http_headers = [ "Content-Type:application/json" ]
  ## With this HTTP status code, the http plugin checks that the HTTP request is completed normally.
  ## As a result, any status code that is not a specified status code is considered to be an error condition and processed.
  success_status_codes = [ 200, 201, 204 ]
  ## Configure http.Client.Timeout in seconds. Default : 3
  timeout = 3

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

const (
	POST = "POST"

	DEFAULT_TIME_OUT = 3
)

type Http struct {
	// http required option
	URL                string   `toml:"url"`
	HttpHeaders        []string `toml:"http_headers"`
	SuccessStatusCodes []int    `toml:"success_status_codes"`

	// Option with http default value
	Timeout int `toml:"timeout"`

	client     http.Client
	serializer serializers.Serializer
	// SuccessStatusCode that stores option values received with expected_status_codes
	SuccessStatusCode map[int]bool
}

func (h *Http) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

// Connect to the Output
func (h *Http) Connect() error {
	h.client = http.Client{
		Timeout: time.Duration(h.Timeout) * time.Second,
	}

	return nil
}

// Close is not implemented. Because http.Client not provided connection close policy. Instead, uses the response.Body.Close() pattern.
func (h *Http) Close() error {
	return nil
}

// Description A plugin that can transmit metrics over HTTP
func (h *Http) Description() string {
	return "A plugin that can transmit metrics over HTTP"
}

// SampleConfig provides sample example for developer
func (h *Http) SampleConfig() string {
	return sampleConfig
}

// Writes metrics over HTTP POST
func (h *Http) Write(metrics []telegraf.Metric) error {
	var mCount int
	var reqBodyBuf []byte

	for _, m := range metrics {
		buf, err := h.serializer.Serialize(m)

		if err != nil {
			return fmt.Errorf("E! Error serializing some metrics: %s", err.Error())
		}

		reqBodyBuf = append(reqBodyBuf, buf...)
		mCount++
	}

	reqBody, err := makeReqBody(h.serializer, reqBodyBuf, mCount)

	if err != nil {
		return fmt.Errorf("E! Error serialized metric is not assembled : %s", err.Error())
	}

	if err := h.write(reqBody); err != nil {
		return err
	}

	return nil
}

func (h *Http) write(reqBody []byte) error {
	req, err := http.NewRequest(POST, h.URL, bytes.NewBuffer(reqBody))

	for _, httpHeader := range h.HttpHeaders {
		keyAndValue := strings.Split(httpHeader, ":")
		req.Header.Set(keyAndValue[0], keyAndValue[1])
	}

	resp, err := h.client.Do(req)

	if err := h.isOk(resp, err); err != nil {
		return err
	}

	defer resp.Body.Close()

	return err
}

func (h *Http) isOk(resp *http.Response, err error) error {
	if resp == nil || err != nil {
		return fmt.Errorf("E! %s request failed! %s.", h.URL, err.Error())
	}

	if !h.isExpStatusCode(resp.StatusCode) {
		return fmt.Errorf("E! %s response is unexpected status code : %d.", h.URL, resp.StatusCode)
	}

	return nil
}

func (h *Http) isExpStatusCode(resStatusCode int) bool {
	if h.SuccessStatusCode == nil {
		h.SuccessStatusCode = make(map[int]bool)

		for _, expectedStatusCode := range h.SuccessStatusCodes {
			h.SuccessStatusCode[expectedStatusCode] = true
		}
	}

	if h.SuccessStatusCode[resStatusCode] {
		return true
	}

	return false
}

// makeReqBody translates each serializer's converted metric into a request body.
func makeReqBody(serializer serializers.Serializer, reqBodyBuf []byte, mCount int) ([]byte, error) {
	switch serializer.(type) {
	case *json.JsonSerializer:
		var arrayJsonObj []byte
		arrayJsonObj = append(arrayJsonObj, []byte("[")...)
		arrayJsonObj = append(arrayJsonObj, reqBodyBuf...)
		arrayJsonObj = append(arrayJsonObj, []byte("]")...)
		return bytes.Replace(arrayJsonObj, []byte("\n"), []byte(","), mCount-1), nil
	default:
		return reqBodyBuf, nil
	}
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &Http{
			Timeout: DEFAULT_TIME_OUT,
		}
	})
}
