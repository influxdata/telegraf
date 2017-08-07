package kairosdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
)

const httpEndpoint = "/api/v1/datapoints"

type httpOutput struct {
	client   *http.Client
	url      string
	timeout  time.Duration
	user     string
	password string
}

var _ innerOutput = (*httpOutput)(nil)

func (h *httpOutput) Connect() error {
	var err error
	h.client = &http.Client{
		Timeout: h.timeout,
	}
	return err
}

func (h *httpOutput) Close() error {
	return nil
}

func (h *httpOutput) Write(metrics []telegraf.Metric) error {
	tsBytes, err := jsonifyMetrics(metrics)
	if err != nil {
		return err
	}
	return h.postHttpMetrics(bytes.NewBuffer(tsBytes))
}

func (h *httpOutput) postHttpMetrics(body *bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, h.url+httpEndpoint, body)
	if err != nil {
		return fmt.Errorf("kairosdb: unable to create http.Request, %s\n", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")

	if h.user != "" {
		req.SetBasicAuth(h.user, h.password)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("kairosdb: error POSTing metrics, %s\n", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("kairosdb: received bad status code, %d\n", resp.StatusCode)
	}

	return nil
}

func jsonifyMetrics(metrics []telegraf.Metric) ([]byte, error) {
	count := 0
	for _, metric := range metrics {
		count += len(metric.Fields())
	}

	allMetrics := make([]datapoint, 0, count)
	for _, metric := range metrics {
		for fieldName, fieldVal := range metric.Fields() {
			datapoint, err := populateDatapoint(metric, fieldName, fieldVal)
			if err != nil {
				log.Println("kairosdb: skipping datapoint for metric ", fieldName, ": ", err)
				continue
			}
			allMetrics = append(allMetrics, datapoint)
		}
	}

	tsBytes, err := json.Marshal(allMetrics)
	if err != nil {
		return nil, fmt.Errorf("kairosdb: unable to marshal TimeSeries, %s\n", err.Error())
	}

	return tsBytes, err
}

type datapoint struct {
	Name      string            `json:"name"`
	Timestamp int64             `json:"timestamp"`
	Value     interface{}       `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
}
