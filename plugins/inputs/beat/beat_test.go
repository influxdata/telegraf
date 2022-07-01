package beat

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func Test_BeatStats(t *testing.T) {
	var beat6StatsAccumulator testutil.Accumulator
	var beatTest = NewBeat()
	// System stats are disabled by default
	beatTest.Includes = []string{"beat", "libbeat", "system", "filebeat"}
	require.NoError(t, beatTest.Init())
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var jsonFilePath string

		switch request.URL.Path {
		case suffixInfo:
			jsonFilePath = "beat6_info.json"
		case suffixStats:
			jsonFilePath = "beat6_stats.json"
		default:
			require.FailNow(t, "cannot handle request")
		}

		data, err := os.ReadFile(jsonFilePath)
		require.NoErrorf(t, err, "could not read from data file %s", jsonFilePath)
		_, err = w.Write(data)
		require.NoError(t, err, "could not write data")
	}))
	requestURL, err := url.Parse(beatTest.URL)
	require.NoErrorf(t, err, "can't parse URL %s", beatTest.URL)
	fakeServer.Listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	require.NoErrorf(t, err, "can't listen for %s: %v", requestURL, err)

	fakeServer.Start()
	defer fakeServer.Close()

	require.NoError(t, err, beatTest.Gather(&beat6StatsAccumulator))

	beat6StatsAccumulator.AssertContainsTaggedFields(
		t,
		"beat",
		map[string]interface{}{
			"cpu_system_ticks":      float64(626970),
			"cpu_system_time_ms":    float64(626972),
			"cpu_total_ticks":       float64(5215010),
			"cpu_total_time_ms":     float64(5215018),
			"cpu_total_value":       float64(5215010),
			"cpu_user_ticks":        float64(4588040),
			"cpu_user_time_ms":      float64(4588046),
			"info_uptime_ms":        float64(327248661),
			"info_ephemeral_id":     "809e3b63-4fa0-4f74-822a-8e3c08298336",
			"memstats_gc_next":      float64(20611808),
			"memstats_memory_alloc": float64(12692544),
			"memstats_memory_total": float64(462910102088),
			"memstats_rss":          float64(80273408),
		},
		map[string]string{
			"beat_beat":    string("filebeat"),
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.4.2"),
		},
	)
	beat6StatsAccumulator.AssertContainsTaggedFields(
		t,
		"beat_filebeat",
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
			"beat_beat":    string("filebeat"),
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.4.2"),
		},
	)
	beat6StatsAccumulator.AssertContainsTaggedFields(
		t,
		"beat_libbeat",
		map[string]interface{}{
			"config_module_running":     float64(0),
			"config_module_starts":      float64(0),
			"config_module_stops":       float64(0),
			"config_reloads":            float64(0),
			"output_type":               "kafka",
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
			"beat_beat":    string("filebeat"),
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.4.2"),
		},
	)
	beat6StatsAccumulator.AssertContainsTaggedFields(
		t,
		"beat_system",
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
			"beat_beat":    string("filebeat"),
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.4.2"),
		},
	)
}

func Test_BeatRequest(t *testing.T) {
	var beat6StatsAccumulator testutil.Accumulator
	beatTest := NewBeat()
	// System stats are disabled by default
	beatTest.Includes = []string{"beat", "libbeat", "system", "filebeat"}
	require.NoError(t, beatTest.Init())
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var jsonFilePath string

		switch request.URL.Path {
		case suffixInfo:
			jsonFilePath = "beat6_info.json"
		case suffixStats:
			jsonFilePath = "beat6_stats.json"
		default:
			require.FailNow(t, "cannot handle request")
		}

		data, err := os.ReadFile(jsonFilePath)
		require.NoErrorf(t, err, "could not read from data file %s", jsonFilePath)
		require.Equal(t, request.Host, "beat.test.local")
		require.Equal(t, request.Method, "POST")
		require.Equal(t, request.Header.Get("Authorization"), "Basic YWRtaW46UFdE")
		require.Equal(t, request.Header.Get("X-Test"), "test-value")

		_, err = w.Write(data)
		require.NoError(t, err, "could not write data")
	}))

	requestURL, err := url.Parse(beatTest.URL)
	require.NoErrorf(t, err, "can't parse URL %s", beatTest.URL)
	fakeServer.Listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	require.NoErrorf(t, err, "can't listen for %s: %v", requestURL, err)
	fakeServer.Start()
	defer fakeServer.Close()

	beatTest.Headers["X-Test"] = "test-value"
	beatTest.HostHeader = "beat.test.local"
	beatTest.Method = "POST"
	beatTest.Username = "admin"
	beatTest.Password = "PWD"

	require.NoError(t, beatTest.Gather(&beat6StatsAccumulator))
}
