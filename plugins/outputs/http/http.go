package http

import (
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf"
	"bytes"
	"net/http"
	"net"
	"time"
	"log"
	"github.com/influxdata/telegraf/plugins/outputs"
	"fmt"
	"errors"
)

var sampleConfig = `
  ## Will be transmitted telegraf metrics to the HTTP Server using the below URL.
  url = "http://127.0.0.1:8080/metric"
  ## HTTP Content-Type. Default : application/json
  content_type = "application/json"
  ## Set the number of times to retry when the status code is not 200 or an error occurs during HTTP call. Default: 3
  retry = 3
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
	CONTENT_TYPE = "Content-Type"

	DEFAULT_RETRY = 3
	DEFAULT_CONTENT_TYPE = "application/json"
	DEFAULT_TLS_HANDSHAKE_TIMEOUT = 10
	DEFAULT_RESPONSE_HEADER_TIMEOUT = 3
	DEFAULT_DIAL_TIME_OUT = 3
	DEFAULT_KEEP_ALIVE = 3
	DEFAULT_EXPECT_CONTINUE_TIMEOUT = 3
	DEFAULT_IDLE_CONN_TIMEOUT = 3
)

type Http struct {
	// http required option
	URL                   string `toml:"url"`

	// http default value가 있는 option
	ContentType           string `toml:"content_type"`
	Retry                 int `toml:"retry"`
	TLSHandshakeTimeout   int `toml:"tls_handshake_timeout"`
	ResponseHeaderTimeout int `toml:"response_header_timeout"`
	DialTimeOut           int `toml:"dial_timeout"`
	KeepAlive             int `toml:"keepalive"`
	ExpectContinueTimeout int `toml:"expect_continue_timeout"`
	IdleConnTimeout       int `toml:"idle_conn_timeout"`

	client                http.Client
	serializer            serializers.Serializer
}

func (h *Http) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

// Connect to the Output
func (h Http) Connect() error {
	h.client = http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   time.Duration(h.DialTimeOut) * time.Second,
				KeepAlive: time.Duration(h.KeepAlive) * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   time.Duration(h.TLSHandshakeTimeout) * time.Second,
			ResponseHeaderTimeout: time.Duration(h.ResponseHeaderTimeout) * time.Second,
			ExpectContinueTimeout: time.Duration(h.ExpectContinueTimeout) * time.Second,
			IdleConnTimeout: time.Duration(h.IdleConnTimeout) * time.Second,
		},
	}

	return nil
}

// Close is not implemented. Because http.Client not provided connection close policy. Instead, uses the response.Body.Close() pattern.
func (h Http) Close() error {
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
// If the request is failed, try again until retry limit.
func (h Http) Write(metrics []telegraf.Metric) error {
	if h.URL == "" {
		return errors.New("Http Output URL Option is empty! It is necessary.")
	}

	for _, metric := range metrics {
		buf, err := h.serializer.Serialize(metric)

		if err != nil {
			log.Println("E! Error serializing some metrics: %s", err.Error())
		}

		if response, err := write(h, buf); err != nil || response.StatusCode != 200 {
			for i := 1; i <= h.Retry; i++ {
				response, err := write(h, buf)

				responseBodyClose(response)

				if err == nil || response.StatusCode == 200 {
					break;
				}

				if err != nil || response.StatusCode != 200 {
					log.Println(fmt.Sprintf("E! [Try %d] %s request failed! Try again because retry limit is %d. Http Error: %s", i, h.URL, h.Retry, err.Error()))

					if (i == h.Retry) {
						return errors.New(fmt.Sprintf("E! Since the retry limit %d has been reached, this request is discarded.", h.Retry))
					}
				}
			}
		} else {
			responseBodyClose(response)
		}
	}

	return nil
}

func responseBodyClose(response *http.Response) {
	if response != nil {
		defer response.Body.Close()
	}
}

func write(h Http, buf []byte) (*http.Response, error) {
	req, err := http.NewRequest(POST, h.URL, bytes.NewBuffer(buf))

	req.Header.Set(CONTENT_TYPE, h.ContentType)
	req.Close = true

	response, err := h.client.Do(req);

	return response, err
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &Http{
			ContentType:           DEFAULT_CONTENT_TYPE,
			Retry:                 DEFAULT_RETRY,
			TLSHandshakeTimeout:   DEFAULT_TLS_HANDSHAKE_TIMEOUT,
			ResponseHeaderTimeout: DEFAULT_RESPONSE_HEADER_TIMEOUT,
			DialTimeOut:           DEFAULT_DIAL_TIME_OUT,
			KeepAlive:             DEFAULT_KEEP_ALIVE,
			ExpectContinueTimeout: DEFAULT_EXPECT_CONTINUE_TIMEOUT,
			IdleConnTimeout:       DEFAULT_IDLE_CONN_TIMEOUT,
		}
	})
}