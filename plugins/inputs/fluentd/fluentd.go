package fluentd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	measurement  = "fluentd"
	description  = "Read metrics exposed by fluentd in_monitor plugin"
	sampleConfig = `
  ## This plugin reads information exposed by fluentd (using /api/plugins.json endpoint).
  ##
  ## Endpoint:
  ## - only one URI is allowed
  ## - https is not supported
  endpoint = "http://localhost:24220/api/plugins.json"

  ## Define which plugins have to be excluded (based on "type" field - e.g. monitor_agent)
  exclude = [
	  "monitor_agent",
	  "dummy",
  ]
`
)

// Fluentd - plugin main structure
type Fluentd struct {
	Endpoint string
	Exclude  []string
	client   *http.Client
}

type endpointInfo struct {
	Payload []pluginData `json:"plugins"`
}

type pluginData struct {
	PluginID              string   `json:"plugin_id"`
	PluginType            string   `json:"type"`
	PluginCategory        string   `json:"plugin_category"`
	RetryCount            *float64 `json:"retry_count"`
	BufferQueueLength     *float64 `json:"buffer_queue_length"`
	BufferTotalQueuedSize *float64 `json:"buffer_total_queued_size"`
}

// parse JSON from fluentd Endpoint
// Parameters:
// 		data: unprocessed json recivied from endpoint
//
// Returns:
//		pluginData:		slice that contains parsed plugins
//		error:			error that may have occurred
func parse(data []byte) (datapointArray []pluginData, err error) {
	var endpointData endpointInfo

	if err = json.Unmarshal(data, &endpointData); err != nil {
		err = fmt.Errorf("Processing JSON structure")
		return
	}

	for _, point := range endpointData.Payload {
		datapointArray = append(datapointArray, point)
	}

	return
}

// Description - display description
func (h *Fluentd) Description() string { return description }

// SampleConfig - generate configuretion
func (h *Fluentd) SampleConfig() string { return sampleConfig }

// Gather - Main code responsible for gathering, processing and creating metrics
func (h *Fluentd) Gather(acc telegraf.Accumulator) error {

	_, err := url.Parse(h.Endpoint)
	if err != nil {
		return fmt.Errorf("Invalid URL \"%s\"", h.Endpoint)
	}

	if h.client == nil {

		tr := &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
		}

		client := &http.Client{
			Transport: tr,
			Timeout:   time.Duration(4 * time.Second),
		}

		h.client = client
	}

	resp, err := h.client.Get(h.Endpoint)

	if err != nil {
		return fmt.Errorf("Unable to perform HTTP client GET on \"%s\": %s", h.Endpoint, err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("Unable to read the HTTP body \"%s\": %s", string(body), err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status ok not met")
	}

	dataPoints, err := parse(body)

	if err != nil {
		return fmt.Errorf("Problem with parsing")
	}

	// Go through all plugins one by one
	for _, p := range dataPoints {

		skip := false

		// Check if this specific type was excluded in configuration
		for _, exclude := range h.Exclude {
			if exclude == p.PluginType {
				skip = true
			}
		}

		// If not, create new metric and add it to Accumulator
		if !skip {
			tmpFields := make(map[string]interface{})

			tmpTags := map[string]string{
				"plugin_id":       p.PluginID,
				"plugin_category": p.PluginCategory,
				"plugin_type":     p.PluginType,
			}

			if p.BufferQueueLength != nil {
				tmpFields["buffer_queue_length"] = *p.BufferQueueLength

			}
			if p.RetryCount != nil {
				tmpFields["retry_count"] = *p.RetryCount
			}

			if p.BufferTotalQueuedSize != nil {
				tmpFields["buffer_total_queued_size"] = *p.BufferTotalQueuedSize
			}

			if !((p.BufferQueueLength == nil) && (p.RetryCount == nil) && (p.BufferTotalQueuedSize == nil)) {
				acc.AddFields(measurement, tmpFields, tmpTags)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("fluentd", func() telegraf.Input { return &Fluentd{} })
}
