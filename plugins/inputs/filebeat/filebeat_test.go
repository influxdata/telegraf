package filebeat

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

var filebeatTest = NewFilebeat()

var (
	filebeat6StatsAccumulator testutil.Accumulator
)

func Test_Filebeat6Stats(test *testing.T) {
	//filebeat6StatsAccumulator.SetDebug(true)
	fakeServer := httptest.NewUnstartedServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				if request.URL.String() == suffixInfo {
					fmt.Fprintf(writer, "%s", string(filebeat6Info))
				} else if request.URL.String() == suffixStats {
					fmt.Fprintf(writer, "%s", string(filebeat6Stats))
				} else {
					test.Logf("Unkown URL: " + request.URL.String())
				}
			},
		),
	)
	requestURL, err := url.Parse(filebeatTest.URL)
	if err != nil {
		test.Logf("Can't connect to: %s", filebeatTest.URL)
	}
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if filebeatTest.client == nil {
		client, err := filebeatTest.createHttpClient()

		if err != nil {
			test.Logf("Can't createHttpClient")
		}
		filebeatTest.client = client
	}

	err = filebeatTest.gatherStats(&filebeat6StatsAccumulator)
	if err != nil {
		test.Logf("Can't gather stats")
	}

	filebeat6StatsAccumulator.AssertContainsTaggedFields(
		test,
		"filebeat_beat",
		map[string]interface{}{
			"cpu_system_ticks":      float64(626970),
			"cpu_system_time_ms":    float64(626972),
			"cpu_total_ticks":       float64(5215010),
			"cpu_total_time_ms":     float64(5215018),
			"cpu_total_value":       float64(5215010),
			"cpu_user_ticks":        float64(4588040),
			"cpu_user_time_ms":      float64(4588046),
			"info_uptime_ms":        float64(327248661),
			"memstats_gc_next":      float64(20611808),
			"memstats_memory_alloc": float64(12692544),
			"memstats_memory_total": float64(462910102088),
			"memstats_rss":          float64(80273408),
		},
		map[string]string{
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.4.2"),
		},
	)

	filebeat6StatsAccumulator.AssertContainsTaggedFields(
		test,
		"filebeat",
		map[string]interface{}{
			"events_active":             float64(0),
			"events_added":              float64(182990),
			"events_done":               float64(182990),
			"harvester_closed":          float64(2222),
			"harvester_open_files":      float64(4),
			"harvester_running":         float64(4),
			"harvester_skipped":         float64(0),
			"harvester_started":         float64(2226),
			"input_log_files_renamed":   float64(0),
			"input_log_files_truncated": float64(0),
		},
		map[string]string{
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.4.2"),
		},
	)

	filebeat6StatsAccumulator.AssertContainsTaggedFields(
		test,
		"filebeat_libbeat",
		map[string]interface{}{
			"config_module_running":     float64(0),
			"config_module_starts":      float64(0),
			"config_module_stops":       float64(0),
			"config_reloads":            float64(0),
			"output_events_acked":       float64(172067),
			"output_events_active":      float64(0),
			"output_events_batches":     float64(1490),
			"output_events_dropped":     float64(0),
			"output_events_duplicates":  float64(0),
			"output_events_failed":      float64(0),
			"output_events_total":       float64(172067),
			"output_read_bytes":         float64(0),
			"output_read_errors":        float64(0),
			"output_write_bytes":        float64(0),
			"output_write_errors":       float64(0),
			"outputs_kafka_bytes_read":  float64(1048670),
			"outputs_kafka_bytes_write": float64(43136887),
			"pipeline_clients":          float64(1),
			"pipeline_events_active":    float64(0),
			"pipeline_events_dropped":   float64(0),
			"pipeline_events_failed":    float64(0),
			"pipeline_events_filtered":  float64(10923),
			"pipeline_events_published": float64(172067),
			"pipeline_events_retry":     float64(14),
			"pipeline_events_total":     float64(182990),
			"pipeline_queue_acked":      float64(172067),
		},
		map[string]string{
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.4.2"),
		},
	)

	filebeat6StatsAccumulator.AssertContainsTaggedFields(
		test,
		"filebeat_system",
		map[string]interface{}{
			"cpu_cores":    float64(32),
			"load_1":       float64(32.49),
			"load_15":      float64(41.9),
			"load_5":       float64(40.16),
			"load_norm_1":  float64(1.0153),
			"load_norm_15": float64(1.3094),
			"load_norm_5":  float64(1.255),
		},
		map[string]string{
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.4.2"),
		},
	)

}

func Test_Filebeat6Request(test *testing.T) {
	//filebeat6StatsAccumulator.SetDebug(true)
	fakeServer := httptest.NewUnstartedServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				if request.URL.String() == suffixInfo {
					fmt.Fprintf(writer, "%s", string(filebeat6Info))
				} else if request.URL.String() == suffixStats {
					fmt.Fprintf(writer, "%s", string(filebeat6Stats))
				} else {
					test.Logf("Unkown URL: " + request.URL.String())
				}

				assert.Equal(test, request.Host, "filebeat.test.local")
				assert.Equal(test, request.Method, "POST")
				assert.Equal(test, request.Header.Get("Authorization"), "Basic YWRtaW46UFdE")
				assert.Equal(test, request.Header.Get("X-Test"), "test-value")
			},
		),
	)
	requestURL, err := url.Parse(filebeatTest.URL)
	if err != nil {
		test.Logf("Can't connect to: %s", filebeatTest.URL)
	}
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if filebeatTest.client == nil {
		client, err := filebeatTest.createHttpClient()

		if err != nil {
			test.Logf("Can't createHttpClient")
		}
		filebeatTest.client = client
	}

	filebeatTest.Headers["X-Test"] = "test-value"
	filebeatTest.HostHeader = "filebeat.test.local"
	filebeatTest.Method = "POST"
	filebeatTest.Username = "admin"
	filebeatTest.Password = "PWD"

	err = filebeatTest.gatherStats(&filebeat6StatsAccumulator)
	if err != nil {
		test.Logf("Can't gather stats")
	}

}
