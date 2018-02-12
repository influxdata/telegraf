package fluentd

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

// sampleJSON from fluentd version '0.14.9'
const sampleJSON = `
{
  "plugins": [
    {
      "plugin_id": "object:f48698",
      "plugin_category": "input",
      "type": "dummy",
      "config": {
        "@type": "dummy",
        "@log_level": "info",
        "tag": "stdout.page.node",
        "rate": "",
        "dummy": "{\"hello\":\"world_from_first_dummy\"}",
        "auto_increment_key": "id1"
      },
      "output_plugin": false,
      "retry_count": null
    },
    {
      "plugin_id": "object:e27138",
      "plugin_category": "input",
      "type": "dummy",
      "config": {
        "@type": "dummy",
        "@log_level": "info",
        "tag": "stdout.superproject.supercontainer",
        "rate": "",
        "dummy": "{\"hello\":\"world_from_second_dummy\"}",
        "auto_increment_key": "id1"
      },
      "output_plugin": false,
      "retry_count": null
    },
    {
      "plugin_id": "object:d74060",
      "plugin_category": "input",
      "type": "monitor_agent",
      "config": {
        "@type": "monitor_agent",
        "@log_level": "error",
        "bind": "0.0.0.0",
        "port": "24220"
      },
      "output_plugin": false,
      "retry_count": null
    },
    {
      "plugin_id": "object:11a5e2c",
      "plugin_category": "output",
      "type": "stdout",
      "config": {
        "@type": "stdout"
      },
      "output_plugin": true,
      "retry_count": 0
    },
    {
      "plugin_id": "object:11237ec",
      "plugin_category": "output",
      "type": "s3",
      "config": {
        "@type": "s3",
        "@log_level": "info",
        "aws_key_id": "xxxxxx",
        "aws_sec_key": "xxxxxx",
        "s3_bucket": "bucket",
        "s3_endpoint": "http://mock:4567",
        "path": "logs/%Y%m%d_%H/${tag[1]}/",
        "time_slice_format": "%M",
        "s3_object_key_format": "%{path}%{time_slice}_%{hostname}_%{index}_%{hex_random}.%{file_extension}",
        "store_as": "gzip"
      },
      "output_plugin": true,
      "buffer_queue_length": 0,
      "buffer_total_queued_size": 0,
      "retry_count": 0
    }
  ]
}
`

var (
	zero           float64
	err            error
	pluginOutput   []pluginData
	expectedOutput = []pluginData{
		// 		{"object:f48698", "dummy", "input", nil, nil, nil},
		// 		{"object:e27138", "dummy", "input", nil, nil, nil},
		// 		{"object:d74060", "monitor_agent", "input", nil, nil, nil},
		{"object:11a5e2c", "stdout", "output", (*float64)(&zero), nil, nil},
		{"object:11237ec", "s3", "output", (*float64)(&zero), (*float64)(&zero), (*float64)(&zero)},
	}
	fluentdTest = &Fluentd{
		Endpoint: "http://localhost:8081",
	}
)

func Test_parse(t *testing.T) {

	t.Log("Testing parser function")
	_, err := parse([]byte(sampleJSON))

	if err != nil {
		t.Error(err)
	}

}

func Test_Gather(t *testing.T) {
	t.Logf("Start HTTP mock (%s) with sampleJSON", fluentdTest.Endpoint)

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", string(sampleJSON))
	}))

	requestURL, err := url.Parse(fluentdTest.Endpoint)

	ts.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))

	ts.Start()

	defer ts.Close()

	var acc testutil.Accumulator
	err = fluentdTest.Gather(&acc)

	if err != nil {
		t.Error(err)
	}

	if !acc.HasMeasurement("fluentd") {
		t.Errorf("acc.HasMeasurement: expected fluentd")
	}

	assert.Equal(t, expectedOutput[0].PluginID, acc.Metrics[0].Tags["plugin_id"])
	assert.Equal(t, expectedOutput[0].PluginType, acc.Metrics[0].Tags["plugin_type"])
	assert.Equal(t, expectedOutput[0].PluginCategory, acc.Metrics[0].Tags["plugin_category"])
	assert.Equal(t, *expectedOutput[0].RetryCount, acc.Metrics[0].Fields["retry_count"])

	assert.Equal(t, expectedOutput[1].PluginID, acc.Metrics[1].Tags["plugin_id"])
	assert.Equal(t, expectedOutput[1].PluginType, acc.Metrics[1].Tags["plugin_type"])
	assert.Equal(t, expectedOutput[1].PluginCategory, acc.Metrics[1].Tags["plugin_category"])
	assert.Equal(t, *expectedOutput[1].RetryCount, acc.Metrics[1].Fields["retry_count"])
	assert.Equal(t, *expectedOutput[1].BufferQueueLength, acc.Metrics[1].Fields["buffer_queue_length"])
	assert.Equal(t, *expectedOutput[1].BufferTotalQueuedSize, acc.Metrics[1].Fields["buffer_total_queued_size"])

}
