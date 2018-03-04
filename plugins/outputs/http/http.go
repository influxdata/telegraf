package http

import (
	"bytes"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"net/http"
	"strings"
	"time"
)

var sampleConfig = `
  ## It requires a url name.
  ## Will be transmitted telegraf metrics to the HTTP Server using the below URL.
  ## Note that not support the HTTPS.
  url = "http://127.0.0.1:8080/metric"
  ## Configure dial timeout in seconds. Default : 3
  timeout = 3
  ## http_headers option can add a custom header to the request.
  ## Content-Type is required http header in http plugin.
  ## so content-type of HTTP specification (plain/text, application/json, etc...) must be filled out.
  [outputs.http.headers]
    Content-Type = "plain/text"
  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

const (
	POST = "POST"

	DEFAULT_TIME_OUT = 3

	CONTENT_TYPE     = "content-type"
	APPLICATION_JSON = "application/json"
	PLAIN_TEXT       = "plain/text"
)

type Http struct {
	// http required option
	URL     string `toml:"url"`
	Headers map[string]string

	// Option with http default value
	Timeout int `toml:"timeout"`

	client     http.Client
	serializer serializers.Serializer
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

	var contentType string
	var err error

	if contentType, err = getContentType(h.Headers); err != nil {
		return err
	}

	reqBody, err := makeReqBody(contentType, reqBodyBuf, mCount)

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

	for k, v := range h.Headers {
		req.Header.Set(k, v)
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

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d\n", resp.StatusCode)
	}

	return nil
}

func getContentType(headers map[string]string) (string, error) {
	var contentType string

	for k, v := range headers {
		if strings.ToLower(k) == CONTENT_TYPE {
			contentType = strings.ToLower(v)

			return contentType, nil
		}
	}

	return "", fmt.Errorf("E! httpHeader require content-type!")
}

// makeReqBody translates each serializer's converted metric into a request body by HTTP content-type format.
func makeReqBody(contentType string, reqBodyBuf []byte, mCount int) ([]byte, error) {
	switch contentType {
	case APPLICATION_JSON:
		var arrayJsonObj []byte
		arrayJsonObj = append(arrayJsonObj, []byte("[")...)
		arrayJsonObj = append(arrayJsonObj, reqBodyBuf...)
		arrayJsonObj = append(arrayJsonObj, []byte("]")...)
		return bytes.Replace(arrayJsonObj, []byte("\n"), []byte(","), mCount-1), nil
	case PLAIN_TEXT:
		return reqBodyBuf, nil
	default:
		return nil, fmt.Errorf("E! HTTP %s content-type is not supported!", contentType)
	}
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &Http{
			Timeout: DEFAULT_TIME_OUT,
		}
	})
}
