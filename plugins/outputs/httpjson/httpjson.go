package httpjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Httpjson struct {
	Name    string
	Server  string
	Data    map[string]string
	Headers map[string]string
}

type Metric struct {
	Name   string `json:"name"`
	Fields string `json:"fields"`
	Tags   string `json:"tags"`
	Time   int64  `json:"time"`
}

func (h *Httpjson) Description() string {
	return `Send telegraf metric through HTTP(s) request`
}

func (h *Httpjson) SampleConfig() string {
	return `
  ## Setup your HTTP Json service name
  # name = "your_httpjson_service_name"

  ## Set the target server. The URL must be a valid HTTP(s) URL
  # server = "http://localhost:3000"

  ## Setup additional data you want to sent along with the metrics data
  ## All value must be string
  # [outputs.httpjson.data]
  #   authToken = "12345"

  ## Setup additional headers for the HTTP(s) request
  ## All value must be string
  # [outputs.httpjson.headers]
  #   Content-Type = "application/json;charset=UTF-8"
`
}

// Connect to the Output
func (h *Httpjson) Connect() error {
	return nil
}

// Close any connections to the Output
func (h *Httpjson) Close() error {
	return nil
}

// Write takes in group of points to be written to the Output
func (h *Httpjson) Write(metrics []telegraf.Metric) error {
	// Don't make any request if metrics empty
	if len(metrics) == 0 {
		return nil
	}

	if h.Server == "" {
		return fmt.Errorf("You need to setup server")
	}

	// Prepare URL
	requestURL, err := url.ParseRequestURI(h.Server)
	if err != nil {
		return fmt.Errorf("Invalid server URL \"%s\"", h.Server)
	}

	// Collect metrics
	var Metrics []Metric
	for _, metric := range metrics {
		var timestamp time.Duration
		unitsNanoseconds := timestamp.Nanoseconds()

		// if the units passed in were less than or equal to zero,
		// then serialize the timestamp in seconds (the default)
		if unitsNanoseconds <= 0 {
			unitsNanoseconds = 1000000000
		}

		var fields, tags bytes.Buffer
		var index int

		// Construct fields metric into string
		for k, v := range metric.Fields() {
			if index == 0 {
				fields.WriteString(fmt.Sprintf("%s=%v", k, v))
			} else {
				fields.WriteString(fmt.Sprintf(",%s=%v", k, v))
			}
			index++
		}

		// Construct tags metric into string
		index = 0
		for k, v := range metric.Tags() {
			if index == 0 {
				tags.WriteString(fmt.Sprintf("%s=%v", k, v))
			} else {
				tags.WriteString(fmt.Sprintf(",%s=%v", k, v))
			}
			index++
		}

		m := Metric{
			Name:   metric.Name(),
			Tags:   tags.String(),
			Fields: fields.String(),
			Time:   metric.Time().UnixNano() / unitsNanoseconds,
		}

		Metrics = append(Metrics, m)
	}

	// Setup request body to send metrics data
	var jsonReq struct {
		Metrics []Metric          `json:"metrics"`
		Data    map[string]string `json:"data"`
	}
	jsonReq.Metrics = Metrics
	if len(h.Data) > 0 {
		jsonReq.Data = h.Data
	}

	// Encode request body
	reqBody, err := json.Marshal(jsonReq)

	// Initialize HTTP(s) request
	req, err := http.NewRequest("POST", requestURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Errorf("Cannot setup HTTP request: %s", err)
	}

	// Add headers parameters
	for k, v := range h.Headers {
		req.Header.Add(k, v)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	// Send HTTP(s) request
	client := http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	var parsedBody map[string]interface{}
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Cannot read response body: %s", err)
	}

	err = json.Unmarshal([]byte(resBody), &parsedBody)
	if err != nil {
		return fmt.Errorf("Cannot parse response body: %s", err)
	}

	return nil
}

func init() {
	outputs.Add("httpjson", func() telegraf.Output {
		return &Httpjson{}
	})
}
