package http_out

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	// "github.com/influxdata/telegraf/plugins/serializers/influx"
	"net/http"
	"net/url"
	// "strings"
	"encoding/json"
	// "io"
	"io/ioutil"
	// "io"
	"bytes"
	"os"
	"time"
)

type HttpOut struct {
	Name       string
	Server     string
	Method     string
	Parameters string
	Headers    map[string]string
	serializer serializers.Serializer
	request    *http.Request
}

type Metric struct {
	Name   string                 `json:"name"`
	Fields map[string]interface{} `json:"fields"`
	Tags   map[string]string      `json:"tags"`
	Time   int64                  `json:"time"`
}

func (h *HttpOut) Description() string {
	return `Send telegraf metric through HTTP(s) request`
}

func (h *HttpOut) SampleConfig() string {
	return `
  [[outputs.http_out]]
    name = "http_out_test"
    server = "http://localhost:3000"
    method = "POST"

    [outputs.http_out.headers]
      Content-Type = "application/json;charset=UTF-8"
`
}

func (h *HttpOut) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

// Connect to the Output
func (h *HttpOut) Connect() error {
	return nil
}

// Close any connections to the Output
func (h *HttpOut) Close() error {
	return nil
}

// Write takes in group of points to be written to the Output
func (h *HttpOut) Write(metrics []telegraf.Metric) error {
	// Don't make any request if metrics empty
	if len(metrics) == 0 {
		return nil
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

		m := Metric{
			Name:   metric.Name(),
			Tags:   metric.Tags(),
			Fields: metric.Fields(),
			Time:   metric.Time().UnixNano() / unitsNanoseconds,
		}

		Metrics = append(Metrics, m)
		// // Send request
		// client := http.Client{}
		// resp, err := client.Do(h.request)
		// if err != nil {
		// fmt.Errorf("Cannot make HTTP request: %+v", err)
		// }
	}

	// defer resp.Body.Close()

	// for _, metric := range metrics {
	// h.makeRequest(metric)
	// fmt.Printf("metric = %+v\n", metric)
	// b, err := h.serializer.Serialize(metric)
	// if err != nil {
	// fmt.Errorf("failed to serialized message %s", err)
	// }

	// fmt.Printf("DataFormat = %+v\n", string(b))

	// }
	// h.DebugToFile(metrics)

	if h.Server == "" {
		return fmt.Errorf("You need to setup a server")
	}

	// Prepare URL
	requestURL, err := url.Parse(h.Server)

	if err != nil {
		return fmt.Errorf("Invalid server URL \"%s\"", h.Server)
	}

	// Setup request body to send metrics data
	var jsonReq struct{ Metrics []Metric }
	jsonReq.Metrics = Metrics
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

	// Send HTTP(s) request
	client := http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	var parsedBody map[string]interface{}
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("Cannot real response body: %s", err)
	}

	err = json.Unmarshal([]byte(resBody), &parsedBody)
	if err != nil {
		fmt.Errorf("Cannot parse response body: %s", err)
	}

	return nil
}

// func (h *HttpOut) makeRequest(metric telegraf.Metric) error {
// if h.Server == "" {
// return fmt.Errorf("You need to setup a server")
// }

// // Prepare URL
// requestURL, err := url.Parse(h.Server)

// if err != nil {
// return fmt.Errorf("Invalid server URL \"%s\"", h.Server)
// }

// // reqBody := bytes.NewBufferString(metric)
// reqBody := influx.NewReader(metrics, h.serializer)
// req, err := http.NewRequest("POST", requestURL.String(), reqBody)
// if err != nil {
// fmt.Errorf("Cannot setup HTTP request: %s", err)
// }

// // Add headers parameters
// for k, v := range h.Headers {
// req.Header.Add(k, v)
// }

// client := http.Client{}
// resp, err := client.Do(req)

// defer resp.Body.Close()

// // var parsedBody HostData
// // resBody, err := ioutil.ReadAll(res.Body)

// // err = json.Unmarshal([]byte(resBody), &parsedBody)
// // if err != nil {
// // fmt.Errorf("Cannot parse response body: %s", err)
// // }

// return nil
// }

func (h *HttpOut) DebugToFile(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	f, err := os.OpenFile("/Users/opanmustopah/test-go.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	for _, metric := range metrics {
		d, err := h.serializer.Serialize(metric)
		if _, err = f.Write(d); err != nil {
			return err
		}
	}
	return nil
}

func IsError(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	outputs.Add("http_out", func() telegraf.Output {
		return &HttpOut{
			Method: "POST",
		}
	})
}
