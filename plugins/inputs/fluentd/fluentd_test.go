package fluentd

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

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
	err            error
	pluginOutput   []pluginData
	expectedOutput = []pluginData{
		{"object:f48698", "dummy", "input", 0, 0, 0},
		{"object:e27138", "dummy", "input", 0, 0, 0},
		{"object:d74060", "monitor_agent", "input", 0, 0, 0},
		{"object:11a5e2c", "stdout", "output", 0, 0, 0},
		{"object:11237ec", "s3", "output", 0, 0, 0},
	}
	fluentdTest = &Fluentd{
		Endpoint: "http://localhost:8081",
	}
)

func Test_parse(t *testing.T) {

	t.Log("Testing parser function")
	pluginOutput, err := parse([]byte(sampleJSON))

	if err != nil {
		t.Error(err)
	}

	if len(pluginOutput) != len(expectedOutput) {
		t.Errorf("lengthOfPluginOutput: expected %d, actual %d", len(pluginOutput), len(expectedOutput))
	}

	if !reflect.DeepEqual(pluginOutput, expectedOutput) {
        t.Errorf("pluginOutput is different from expectedOutput")
	}

}

func Test_Gather(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gather function test")
	}

	t.Log("Testing Gather function")

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

	if len(expectedOutput) != len(acc.Metrics) {
		t.Errorf("acc.Metrics: expected %d, actual %d", len(expectedOutput), len(acc.Metrics))
	}

	for i := 0; i < len(acc.Metrics); i++ {
		assert.Equal(t, expectedOutput[i].PluginID, acc.Metrics[i].Tags["PluginID"])
		assert.Equal(t, expectedOutput[i].PluginType, acc.Metrics[i].Tags["PluginType"])
		assert.Equal(t, expectedOutput[i].PluginCategory, acc.Metrics[i].Tags["PluginCategory"])
		assert.Equal(t, expectedOutput[i].RetryCount, acc.Metrics[i].Fields["RetryCount"])
		assert.Equal(t, expectedOutput[i].BufferQueueLength, acc.Metrics[i].Fields["BufferQueueLength"])
		assert.Equal(t, expectedOutput[i].BufferTotalQueuedSize, acc.Metrics[i].Fields["BufferTotalQueuedSize"])
	}

}
