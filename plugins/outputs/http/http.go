package http

import (
	"bytes"
	"context"
	ejson "encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"net"
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
  expected_status_codes = [ 200, 204 ]
  ## Configure response header timeout in seconds. Default : 3
  response_header_timeout = 3
  ## Configure dial timeout in seconds. Default : 3
  dial_timeout = 3
  ## max_bulk_limit defines how much of the metrics will be sent.
  ## Max_bulk_limit = 0   => Write all metrics collected during flush_interval.
  ## Max_bulk_limit = 100 => Write 100 of all metrics collected during flush_interval.
  ## Note that If the amount of metric collected during flush_interval is less than max_bulk_limit, then all of the stacked metrics are sent.
  max_bulk_limit = 0

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
	DEFAULT_MAX_BULK_LIMIT          = 0
)

type Http struct {
	// http required option
	URL            string   `toml:"url"`
	HttpHeaders    []string `toml:"http_headers"`
	ExpStatusCodes []int    `toml:"expected_status_codes"`

	// Option with http default value
	ResHeaderTimeout int `toml:"response_header_timeout"`
	DialTimeOut      int `toml:"dial_timeout"`
	MaxBulkLimit     int `toml:"max_bulk_limit"`

	client     http.Client
	serializer serializers.Serializer
	// expStatusCode that stores option values received with expected_status_codes
	expStatusCode map[int]bool

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
				Timeout: time.Duration(h.DialTimeOut) * time.Second,
			}).Dial,
			ResponseHeaderTimeout: time.Duration(h.ResHeaderTimeout) * time.Second,
		},
	}

	h.cancelContext, h.cancel = context.WithCancel(context.TODO())

	return nil
}

// Close the Client Connection using Context.
func (h *Http) Close() error {
	h.cancel()

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
	if err := validate(h); err != nil {
		return err
	}

	var reqBodyBuf [][]byte

	for _, metric := range metrics {
		buf, err := h.serializer.Serialize(metric)

		if err != nil {
			return fmt.Errorf("E! Error serializing some metrics: %s", err.Error())
		}

		reqBodyBuf = append(reqBodyBuf, buf)
	}

	if h.MaxBulkLimit == 0 || len(reqBodyBuf) <= h.MaxBulkLimit {
		if err := h.write(reqBodyBuf); err != nil {
			return err
		}

		return nil
	}

	if err := h.splitWrite(reqBodyBuf); err != nil {
		return err
	}

	return nil
}

// splitWrite sends the divided metric by max_bulk_limit.
func (h *Http) splitWrite(reqBodyBuf [][]byte) error {
	s := 0
	e := h.MaxBulkLimit
	mLength := len(reqBodyBuf)

	for true {
		if mLength <= e {
			if err := h.write(reqBodyBuf[s:mLength]); err != nil {
				return err
			}

			break
		} else {
			if err := h.write(reqBodyBuf[s:e]); err != nil {
				return err
			}
		}

		s += h.MaxBulkLimit
		e += h.MaxBulkLimit
	}

	return nil
}

func (h *Http) write(reqBodyBuf [][]byte) error {
	requestBody, err := makeReqBody(h.serializer, reqBodyBuf)

	if err != nil {
		return fmt.Errorf("E! Error serialized metric is not assembled : %s", err.Error())
	}

	req, err := http.NewRequest(POST, h.URL, bytes.NewBuffer(requestBody))

	for _, httpHeader := range h.HttpHeaders {
		keyAndValue := strings.Split(httpHeader, ":")
		req.Header.Set(keyAndValue[0], keyAndValue[1])
	}

	req.Close = true
	req.WithContext(h.cancelContext)

	response, err := h.client.Do(req)

	if err := h.isOk(response, err); err != nil {
		return err
	}

	response.Body.Close()

	return err
}

func (h *Http) isOk(res *http.Response, err error) error {
	if res == nil || err != nil {
		return fmt.Errorf("E! %s request failed! %s.", h.URL, err.Error())
	}

	if !h.isExpStatusCode(res.StatusCode) {
		return fmt.Errorf("E! %s response is unexpected status code : %d.", h.URL, res.StatusCode)
	}

	return nil
}

func (h *Http) isExpStatusCode(resStatusCode int) bool {
	if h.expStatusCode == nil {
		h.expStatusCode = make(map[int]bool)

		for _, expectedStatusCode := range h.ExpStatusCodes {
			h.expStatusCode[expectedStatusCode] = true
		}
	}

	if h.expStatusCode[resStatusCode] {
		return true
	}

	return false
}

// required option validate
func validate(h *Http) error {
	if h.URL == "" || len(h.HttpHeaders) == 0 || len(h.ExpStatusCodes) == 0 {
		return errors.New("E! Http ouput plugin is not working. Because your configuration omits the required option. Please check url, http_headers, expected_status_codes is empty!")
	}

	return nil
}

// makeReqBody translates each serializer's converted metric into a request body.
func makeReqBody(serializer serializers.Serializer, reqBodyBuf [][]byte) ([]byte, error) {
	switch serializer.(type) {
	case *json.JsonSerializer:
		return makeJsonFormatReqBody(reqBodyBuf)
	default:
		return makePlainTextFormatReqBody(reqBodyBuf)
	}
}

func makePlainTextFormatReqBody(reqBodyBuf [][]byte) ([]byte, error) {
	var reqBody bytes.Buffer

	for _, serializedMetric := range reqBodyBuf {
		reqBody.Write(serializedMetric)
	}

	return reqBody.Bytes(), nil
}

func makeJsonFormatReqBody(reqBodyBuf [][]byte) ([]byte, error) {
	var reqBody []map[string]interface{}

	for _, serializedMetric := range reqBodyBuf {
		var jsonObject map[string]interface{}

		err := ejson.Unmarshal(serializedMetric, &jsonObject)

		if err != nil {
			return nil, fmt.Errorf("E! HTTP json unmarshal is fail! It probably does not seem to fit in the json format. Please check %s", serializedMetric)
		}

		reqBody = append(reqBody, jsonObject)
	}

	return ejson.Marshal(reqBody)
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &Http{
			ResHeaderTimeout: DEFAULT_RESPONSE_HEADER_TIMEOUT,
			DialTimeOut:      DEFAULT_DIAL_TIME_OUT,
			MaxBulkLimit:     DEFAULT_MAX_BULK_LIMIT,
		}
	})
}
