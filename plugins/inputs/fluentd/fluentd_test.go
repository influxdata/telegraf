package fluentd

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
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
      "retry_count": 0,
      "buffer_total_queued_size": 0
    },
    {
      "plugin_id": "object:output_td_1",
      "plugin_category": "output",
      "type": "tdlog",
      "config": {
        "@type": "tdlog",
        "@id": "output_td",
        "apikey": "xxxxxx",
        "auto_create_table": ""
      },
      "output_plugin": true,
      "buffer_queue_length": 0,
      "buffer_total_queued_size": 0,
      "retry_count": 0,
      "emit_records": 0,
      "emit_size": 0,
      "emit_count": 0,
      "write_count": 0,
      "rollback_count": 0,
      "slow_flush_count": 0,
      "flush_time_count": 0,
      "buffer_stage_length": 0,
      "buffer_stage_byte_size": 0,
      "buffer_queue_byte_size": 0,
      "buffer_available_buffer_space_ratios": 0
    }, 
    {
      "plugin_id": "object:output_td_2",
      "plugin_category": "output",
      "type": "tdlog",
      "config": {
        "@type": "tdlog",
        "@id": "output_td",
        "apikey": "xxxxxx",
        "auto_create_table": ""
      },
      "output_plugin": true,
      "buffer_queue_length": 0,
      "buffer_total_queued_size": 0,
      "retry_count": 0,
      "rollback_count": 0,
      "emit_records": 0,
      "slow_flush_count": 0,
      "buffer_available_buffer_space_ratios": 0
    }
  ]
}
`

var (
	zero           float64
	expectedOutput = []pluginData{
		// 		{"object:f48698", "dummy", "input", nil, nil, nil},
		// 		{"object:e27138", "dummy", "input", nil, nil, nil},
		// 		{"object:d74060", "monitor_agent", "input", nil, nil, nil},
		{"object:11a5e2c", "stdout", "output", &zero, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		{"object:11237ec", "s3", "output", &zero, &zero, &zero, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		{"object:output_td_1", "tdlog", "output", &zero, &zero, &zero, &zero, &zero, &zero, &zero, &zero, &zero, &zero, &zero, &zero, &zero, &zero},
		{"object:output_td_2", "tdlog", "output", &zero, &zero, &zero, &zero, &zero, nil, nil, nil, &zero, nil, nil, nil, nil, &zero},
	}
	fluentdTest = &Fluentd{
		Endpoint: "http://localhost:8081",
	}
)

func Test_parse(t *testing.T) {
	t.Log("Testing parser function")
	t.Logf("JSON (%s) ", sampleJSON)
	_, err := parse([]byte(sampleJSON))

	if err != nil {
		t.Error(err)
	}
}

func Test_Gather(t *testing.T) {
	t.Logf("Start HTTP mock (%s) with sampleJSON", fluentdTest.Endpoint)

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(w, "%s", string(sampleJSON))
		require.NoError(t, err)
	}))

	requestURL, err := url.Parse(fluentdTest.Endpoint)
	require.NoError(t, err)
	require.NotNil(t, requestURL)

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

	require.Equal(t, expectedOutput[0].PluginID, acc.Metrics[0].Tags["plugin_id"])
	require.Equal(t, expectedOutput[0].PluginType, acc.Metrics[0].Tags["plugin_type"])
	require.Equal(t, expectedOutput[0].PluginCategory, acc.Metrics[0].Tags["plugin_category"])
	require.Equal(t, *expectedOutput[0].RetryCount, acc.Metrics[0].Fields["retry_count"])

	require.Equal(t, expectedOutput[1].PluginID, acc.Metrics[1].Tags["plugin_id"])
	require.Equal(t, expectedOutput[1].PluginType, acc.Metrics[1].Tags["plugin_type"])
	require.Equal(t, expectedOutput[1].PluginCategory, acc.Metrics[1].Tags["plugin_category"])
	require.Equal(t, *expectedOutput[1].RetryCount, acc.Metrics[1].Fields["retry_count"])
	require.Equal(t, *expectedOutput[1].BufferQueueLength, acc.Metrics[1].Fields["buffer_queue_length"])
	require.Equal(t, *expectedOutput[1].BufferTotalQueuedSize, acc.Metrics[1].Fields["buffer_total_queued_size"])

	require.Equal(t, expectedOutput[2].PluginID, acc.Metrics[2].Tags["plugin_id"])
	require.Equal(t, expectedOutput[2].PluginType, acc.Metrics[2].Tags["plugin_type"])
	require.Equal(t, expectedOutput[2].PluginCategory, acc.Metrics[2].Tags["plugin_category"])
	require.Equal(t, *expectedOutput[2].RetryCount, acc.Metrics[2].Fields["retry_count"])
	require.Equal(t, *expectedOutput[2].BufferQueueLength, acc.Metrics[2].Fields["buffer_queue_length"])
	require.Equal(t, *expectedOutput[2].BufferTotalQueuedSize, acc.Metrics[2].Fields["buffer_total_queued_size"])
	require.Equal(t, *expectedOutput[2].EmitRecords, acc.Metrics[2].Fields["emit_records"])
	require.Equal(t, *expectedOutput[2].EmitSize, acc.Metrics[2].Fields["emit_size"])
	require.Equal(t, *expectedOutput[2].EmitCount, acc.Metrics[2].Fields["emit_count"])
	require.Equal(t, *expectedOutput[2].RollbackCount, acc.Metrics[2].Fields["rollback_count"])
	require.Equal(t, *expectedOutput[2].SlowFlushCount, acc.Metrics[2].Fields["slow_flush_count"])
	require.Equal(t, *expectedOutput[2].WriteCount, acc.Metrics[2].Fields["write_count"])
	require.Equal(t, *expectedOutput[2].FlushTimeCount, acc.Metrics[2].Fields["flush_time_count"])
	require.Equal(t, *expectedOutput[2].BufferStageLength, acc.Metrics[2].Fields["buffer_stage_length"])
	require.Equal(t, *expectedOutput[2].BufferStageByteSize, acc.Metrics[2].Fields["buffer_stage_byte_size"])
	require.Equal(t, *expectedOutput[2].BufferQueueByteSize, acc.Metrics[2].Fields["buffer_queue_byte_size"])
	require.Equal(t, *expectedOutput[2].AvailBufferSpaceRatios, acc.Metrics[2].Fields["buffer_available_buffer_space_ratios"])

	require.Equal(t, expectedOutput[3].PluginID, acc.Metrics[3].Tags["plugin_id"])
	require.Equal(t, expectedOutput[3].PluginType, acc.Metrics[3].Tags["plugin_type"])
	require.Equal(t, expectedOutput[3].PluginCategory, acc.Metrics[3].Tags["plugin_category"])
	require.Equal(t, *expectedOutput[3].RetryCount, acc.Metrics[3].Fields["retry_count"])
	require.Equal(t, *expectedOutput[3].BufferQueueLength, acc.Metrics[3].Fields["buffer_queue_length"])
	require.Equal(t, *expectedOutput[3].BufferTotalQueuedSize, acc.Metrics[3].Fields["buffer_total_queued_size"])
	require.Equal(t, *expectedOutput[3].EmitRecords, acc.Metrics[3].Fields["emit_records"])
	require.Equal(t, *expectedOutput[3].RollbackCount, acc.Metrics[3].Fields["rollback_count"])
	require.Equal(t, *expectedOutput[3].SlowFlushCount, acc.Metrics[3].Fields["slow_flush_count"])
	require.Equal(t, *expectedOutput[3].AvailBufferSpaceRatios, acc.Metrics[3].Fields["buffer_available_buffer_space_ratios"])
}
