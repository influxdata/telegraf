//go:generate ../../../tools/readme_config_includer/generator
package fluentd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const measurement = "fluentd"

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
	PluginID               string   `json:"plugin_id"`
	PluginType             string   `json:"type"`
	PluginCategory         string   `json:"plugin_category"`
	RetryCount             *float64 `json:"retry_count"`
	BufferQueueLength      *float64 `json:"buffer_queue_length"`
	BufferTotalQueuedSize  *float64 `json:"buffer_total_queued_size"`
	RollbackCount          *float64 `json:"rollback_count"`
	EmitRecords            *float64 `json:"emit_records"`
	EmitSize               *float64 `json:"emit_size"`
	EmitCount              *float64 `json:"emit_count"`
	WriteCount             *float64 `json:"write_count"`
	SlowFlushCount         *float64 `json:"slow_flush_count"`
	FlushTimeCount         *float64 `json:"flush_time_count"`
	BufferStageLength      *float64 `json:"buffer_stage_length"`
	BufferStageByteSize    *float64 `json:"buffer_stage_byte_size"`
	BufferQueueByteSize    *float64 `json:"buffer_queue_byte_size"`
	AvailBufferSpaceRatios *float64 `json:"buffer_available_buffer_space_ratios"`
}

// parse JSON from fluentd Endpoint
// Parameters:
//
//	data: unprocessed json received from endpoint
//
// Returns:
//
//	pluginData:		slice that contains parsed plugins
//	error:			error that may have occurred
func parse(data []byte) (datapointArray []pluginData, err error) {
	var endpointData endpointInfo

	if err = json.Unmarshal(data, &endpointData); err != nil {
		err = fmt.Errorf("processing JSON structure")
		return nil, err
	}

	datapointArray = append(datapointArray, endpointData.Payload...)
	return datapointArray, err
}

func (*Fluentd) SampleConfig() string {
	return sampleConfig
}

// Gather - Main code responsible for gathering, processing and creating metrics
func (h *Fluentd) Gather(acc telegraf.Accumulator) error {
	_, err := url.Parse(h.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid URL \"%s\"", h.Endpoint)
	}

	if h.client == nil {
		tr := &http.Transport{
			ResponseHeaderTimeout: 3 * time.Second,
		}

		client := &http.Client{
			Transport: tr,
			Timeout:   4 * time.Second,
		}

		h.client = client
	}

	resp, err := h.client.Get(h.Endpoint)

	if err != nil {
		return fmt.Errorf("unable to perform HTTP client GET on \"%s\": %v", h.Endpoint, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("unable to read the HTTP body \"%s\": %v", string(body), err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status ok not met")
	}

	dataPoints, err := parse(body)

	if err != nil {
		return fmt.Errorf("problem with parsing")
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

			if p.RollbackCount != nil {
				tmpFields["rollback_count"] = *p.RollbackCount
			}

			if p.EmitRecords != nil {
				tmpFields["emit_records"] = *p.EmitRecords
			}

			if p.EmitCount != nil {
				tmpFields["emit_count"] = *p.EmitCount
			}

			if p.EmitSize != nil {
				tmpFields["emit_size"] = *p.EmitSize
			}

			if p.WriteCount != nil {
				tmpFields["write_count"] = *p.WriteCount
			}

			if p.SlowFlushCount != nil {
				tmpFields["slow_flush_count"] = *p.SlowFlushCount
			}

			if p.FlushTimeCount != nil {
				tmpFields["flush_time_count"] = *p.FlushTimeCount
			}

			if p.BufferStageLength != nil {
				tmpFields["buffer_stage_length"] = *p.BufferStageLength
			}

			if p.BufferStageByteSize != nil {
				tmpFields["buffer_stage_byte_size"] = *p.BufferStageByteSize
			}

			if p.BufferQueueByteSize != nil {
				tmpFields["buffer_queue_byte_size"] = *p.BufferQueueByteSize
			}

			if p.AvailBufferSpaceRatios != nil {
				tmpFields["buffer_available_buffer_space_ratios"] = *p.AvailBufferSpaceRatios
			}

			if !((p.BufferQueueLength == nil) &&
				(p.RetryCount == nil) &&
				(p.BufferTotalQueuedSize == nil) &&
				(p.EmitCount == nil) &&
				(p.EmitRecords == nil) &&
				(p.EmitSize == nil) &&
				(p.WriteCount == nil) &&
				(p.FlushTimeCount == nil) &&
				(p.SlowFlushCount == nil) &&
				(p.RollbackCount == nil) &&
				(p.BufferStageLength == nil) &&
				(p.BufferStageByteSize == nil) &&
				(p.BufferQueueByteSize == nil) &&
				(p.AvailBufferSpaceRatios == nil)) {
				acc.AddFields(measurement, tmpFields, tmpTags)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("fluentd", func() telegraf.Input { return &Fluentd{} })
}
