package logstash

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var logstashTest = NewLogstash()

var (
	logstash5accPipelineStats  testutil.Accumulator
	logstash6accPipelinesStats testutil.Accumulator
	logstash7accPipelinesStats testutil.Accumulator
	logstash5accProcessStats   testutil.Accumulator
	logstash6accProcessStats   testutil.Accumulator
	logstash5accJVMStats       testutil.Accumulator
	logstash6accJVMStats       testutil.Accumulator
)

func Test_Logstash5GatherProcessStats(test *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(writer, "%s", string(logstash5ProcessJSON))
		require.NoError(test, err)
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	require.NoErrorf(test, err, "Can't connect to: %s", logstashTest.URL)
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()
		require.NoError(test, err, "Can't createHTTPClient")
		logstashTest.client = client
	}

	err = logstashTest.gatherProcessStats(logstashTest.URL+processStats, &logstash5accProcessStats)
	require.NoError(test, err, "Can't gather Process stats")

	logstash5accProcessStats.AssertContainsTaggedFields(
		test,
		"logstash_process",
		map[string]interface{}{
			"open_file_descriptors":      float64(89.0),
			"max_file_descriptors":       float64(1.048576e+06),
			"cpu_percent":                float64(3.0),
			"cpu_load_average_5m":        float64(0.61),
			"cpu_load_average_15m":       float64(0.54),
			"mem_total_virtual_in_bytes": float64(4.809506816e+09),
			"cpu_total_in_millis":        float64(1.5526e+11),
			"cpu_load_average_1m":        float64(0.49),
			"peak_open_file_descriptors": float64(100.0),
		},
		map[string]string{
			"node_id":      string("a360d8cf-6289-429d-8419-6145e324b574"),
			"node_name":    string("node-5-test"),
			"source":       string("node-5"),
			"node_version": string("5.3.0"),
		},
	)
}

func Test_Logstash6GatherProcessStats(test *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(writer, "%s", string(logstash6ProcessJSON))
		require.NoError(test, err)
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	require.NoErrorf(test, err, "Can't connect to: %s", logstashTest.URL)
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()
		require.NoError(test, err, "Can't createHTTPClient")
		logstashTest.client = client
	}

	err = logstashTest.gatherProcessStats(logstashTest.URL+processStats, &logstash6accProcessStats)
	require.NoError(test, err, "Can't gather Process stats")

	logstash6accProcessStats.AssertContainsTaggedFields(
		test,
		"logstash_process",
		map[string]interface{}{
			"open_file_descriptors":      float64(133.0),
			"max_file_descriptors":       float64(262144.0),
			"cpu_percent":                float64(0.0),
			"cpu_load_average_5m":        float64(42.4),
			"cpu_load_average_15m":       float64(38.95),
			"mem_total_virtual_in_bytes": float64(17923452928.0),
			"cpu_total_in_millis":        float64(5841460),
			"cpu_load_average_1m":        float64(48.2),
			"peak_open_file_descriptors": float64(145.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
		},
	)
}

func Test_Logstash5GatherPipelineStats(test *testing.T) {
	//logstash5accPipelineStats.SetDebug(true)
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(writer, "%s", string(logstash5PipelineJSON))
		require.NoError(test, err)
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	require.NoErrorf(test, err, "Can't connect to: %s", logstashTest.URL)
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()
		require.NoError(test, err, "Can't createHTTPClient")
		logstashTest.client = client
	}

	err = logstashTest.gatherPipelineStats(logstashTest.URL+pipelineStats, &logstash5accPipelineStats)
	require.NoError(test, err, "Can't gather Pipeline stats")

	logstash5accPipelineStats.AssertContainsTaggedFields(
		test,
		"logstash_events",
		map[string]interface{}{
			"duration_in_millis": float64(1151.0),
			"in":                 float64(1269.0),
			"filtered":           float64(1269.0),
			"out":                float64(1269.0),
		},
		map[string]string{
			"node_id":      string("a360d8cf-6289-429d-8419-6145e324b574"),
			"node_name":    string("node-5-test"),
			"source":       string("node-5"),
			"node_version": string("5.3.0"),
		},
	)

	fields := make(map[string]interface{})
	fields["queue_push_duration_in_millis"] = float64(32.0)
	fields["out"] = float64(2.0)

	logstash5accPipelineStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		fields,
		map[string]string{
			"node_id":      string("a360d8cf-6289-429d-8419-6145e324b574"),
			"node_name":    string("node-5-test"),
			"source":       string("node-5"),
			"node_version": string("5.3.0"),
			"plugin_name":  string("beats"),
			"plugin_id":    string("a35197a509596954e905e38521bae12b1498b17d-1"),
			"plugin_type":  string("input"),
		},
	)

	logstash5accPipelineStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(360.0),
			"in":                 float64(1269.0),
			"out":                float64(1269.0),
		},
		map[string]string{
			"node_id":      string("a360d8cf-6289-429d-8419-6145e324b574"),
			"node_name":    string("node-5-test"),
			"source":       string("node-5"),
			"node_version": string("5.3.0"),
			"plugin_name":  string("stdout"),
			"plugin_id":    string("582d5c2becb582a053e1e9a6bcc11d49b69a6dfd-2"),
			"plugin_type":  string("output"),
		},
	)

	logstash5accPipelineStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(228.0),
			"in":                 float64(1269.0),
			"out":                float64(1269.0),
		},
		map[string]string{
			"node_id":      string("a360d8cf-6289-429d-8419-6145e324b574"),
			"node_name":    string("node-5-test"),
			"source":       string("node-5"),
			"node_version": string("5.3.0"),
			"plugin_name":  string("s3"),
			"plugin_id":    string("582d5c2becb582a053e1e9a6bcc11d49b69a6dfd-3"),
			"plugin_type":  string("output"),
		},
	)
}

func Test_Logstash6GatherPipelinesStats(test *testing.T) {
	//logstash6accPipelinesStats.SetDebug(true)
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(writer, "%s", string(logstash6PipelinesJSON))
		require.NoError(test, err)
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	require.NoErrorf(test, err, "Can't connect to: %s", logstashTest.URL)
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()
		require.NoError(test, err, "Can't createHTTPClient")
		logstashTest.client = client
	}

	err = logstashTest.gatherPipelinesStats(logstashTest.URL+pipelineStats, &logstash6accPipelinesStats)
	require.NoError(test, err, "Can't gather Pipeline stats")

	fields := make(map[string]interface{})
	fields["duration_in_millis"] = float64(8540751.0)
	fields["queue_push_duration_in_millis"] = float64(366.0)
	fields["in"] = float64(180659.0)
	fields["filtered"] = float64(180659.0)
	fields["out"] = float64(180659.0)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_events",
		fields,
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
		},
	)

	fields = make(map[string]interface{})
	fields["queue_push_duration_in_millis"] = float64(366.0)
	fields["out"] = float64(180659.0)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		fields,
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("kafka"),
			"plugin_id":    string("input-kafka"),
			"plugin_type":  string("input"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(2117.0),
			"in":                 float64(27641.0),
			"out":                float64(27641.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("mutate"),
			"plugin_id":    string("155b0ad18abbf3df1e0cb7bddef0d77c5ba699efe5a0f8a28502d140549baf54"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(2117.0),
			"in":                 float64(27641.0),
			"out":                float64(27641.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("mutate"),
			"plugin_id":    string("155b0ad18abbf3df1e0cb7bddef0d77c5ba699efe5a0f8a28502d140549baf54"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(13149.0),
			"in":                 float64(180659.0),
			"out":                float64(177549.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("date"),
			"plugin_id":    string("d079424bb6b7b8c7c61d9c5e0ddae445e92fa9ffa2e8690b0a669f7c690542f0"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(2814.0),
			"in":                 float64(76602.0),
			"out":                float64(76602.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("mutate"),
			"plugin_id":    string("25afa60ab6dc30512fe80efa3493e4928b5b1b109765b7dc46a3e4bbf293d2d4"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(9.0),
			"in":                 float64(934.0),
			"out":                float64(934.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("mutate"),
			"plugin_id":    string("2d9fa8f74eeb137bfa703b8050bad7d76636fface729e4585b789b5fc9bed668"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(173.0),
			"in":                 float64(3110.0),
			"out":                float64(0.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("drop"),
			"plugin_id":    string("4ed14c9ef0198afe16c31200041e98d321cb5c2e6027e30b077636b8c4842110"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(5605.0),
			"in":                 float64(75482.0),
			"out":                float64(75482.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("mutate"),
			"plugin_id":    string("358ce1eb387de7cd5711c2fb4de64cd3b12e5ca9a4c45f529516bcb053a31df4"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(313992.0),
			"in":                 float64(180659.0),
			"out":                float64(180659.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("csv"),
			"plugin_id":    string("82a9bbb02fff37a63c257c1f146b0a36273c7cbbebe83c0a51f086e5280bf7bb"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(0.0),
			"in":                 float64(0.0),
			"out":                float64(0.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("mutate"),
			"plugin_id":    string("8fb13a8cdd4257b52724d326aa1549603ffdd4e4fde6d20720c96b16238c18c3"),
			"plugin_type":  string("filter"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(651386.0),
			"in":                 float64(177549.0),
			"out":                float64(177549.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("elasticsearch"),
			"plugin_id":    string("output-elk"),
			"plugin_type":  string("output"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(186751.0),
			"in":                 float64(177549.0),
			"out":                float64(177549.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("kafka"),
			"plugin_id":    string("output-kafka1"),
			"plugin_type":  string("output"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(7335196.0),
			"in":                 float64(177549.0),
			"out":                float64(177549.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"plugin_name":  string("kafka"),
			"plugin_id":    string("output-kafka2"),
			"plugin_type":  string("output"),
		},
	)

	logstash6accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_queue",
		map[string]interface{}{
			"events":                  float64(103),
			"free_space_in_bytes":     float64(36307369984),
			"max_queue_size_in_bytes": float64(1073741824),
			"max_unread_events":       float64(0),
			"page_capacity_in_bytes":  float64(67108864),
			"queue_size_in_bytes":     float64(1872391),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
			"pipeline":     string("main"),
			"queue_type":   string("persisted"),
		},
	)
}

func Test_Logstash5GatherJVMStats(test *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(writer, "%s", string(logstash5JvmJSON))
		require.NoError(test, err)
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	require.NoErrorf(test, err, "Can't connect to: %s", logstashTest.URL)
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()
		require.NoError(test, err, "Can't createHTTPClient")
		logstashTest.client = client
	}

	err = logstashTest.gatherJVMStats(logstashTest.URL+jvmStats, &logstash5accJVMStats)
	require.NoError(test, err, "Can't gather JVM stats")

	logstash5accJVMStats.AssertContainsTaggedFields(
		test,
		"logstash_jvm",
		map[string]interface{}{
			"mem_pools_young_max_in_bytes":                  float64(5.5836672e+08),
			"mem_pools_young_committed_in_bytes":            float64(1.43261696e+08),
			"mem_heap_committed_in_bytes":                   float64(5.1904512e+08),
			"threads_count":                                 float64(29.0),
			"mem_pools_old_peak_used_in_bytes":              float64(1.27900864e+08),
			"mem_pools_old_peak_max_in_bytes":               float64(7.2482816e+08),
			"mem_heap_used_percent":                         float64(16.0),
			"gc_collectors_young_collection_time_in_millis": float64(3235.0),
			"mem_pools_survivor_committed_in_bytes":         float64(1.7825792e+07),
			"mem_pools_young_used_in_bytes":                 float64(7.6049384e+07),
			"mem_non_heap_committed_in_bytes":               float64(2.91487744e+08),
			"mem_pools_survivor_peak_max_in_bytes":          float64(3.4865152e+07),
			"mem_pools_young_peak_max_in_bytes":             float64(2.7918336e+08),
			"uptime_in_millis":                              float64(4.803461e+06),
			"mem_pools_survivor_peak_used_in_bytes":         float64(8.912896e+06),
			"mem_pools_survivor_max_in_bytes":               float64(6.9730304e+07),
			"gc_collectors_old_collection_count":            float64(2.0),
			"mem_pools_survivor_used_in_bytes":              float64(9.419672e+06),
			"mem_pools_old_used_in_bytes":                   float64(2.55801728e+08),
			"mem_pools_old_max_in_bytes":                    float64(1.44965632e+09),
			"mem_pools_young_peak_used_in_bytes":            float64(7.1630848e+07),
			"mem_heap_used_in_bytes":                        float64(3.41270784e+08),
			"mem_heap_max_in_bytes":                         float64(2.077753344e+09),
			"gc_collectors_young_collection_count":          float64(616.0),
			"threads_peak_count":                            float64(31.0),
			"mem_pools_old_committed_in_bytes":              float64(3.57957632e+08),
			"gc_collectors_old_collection_time_in_millis":   float64(114.0),
			"mem_non_heap_used_in_bytes":                    float64(2.68905936e+08),
		},
		map[string]string{
			"node_id":      string("a360d8cf-6289-429d-8419-6145e324b574"),
			"node_name":    string("node-5-test"),
			"source":       string("node-5"),
			"node_version": string("5.3.0"),
		},
	)
}

func Test_Logstash6GatherJVMStats(test *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(writer, "%s", string(logstash6JvmJSON))
		require.NoError(test, err)
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	require.NoErrorf(test, err, "Can't connect to: %s", logstashTest.URL)
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()
		require.NoError(test, err, "Can't createHTTPClient")
		logstashTest.client = client
	}

	err = logstashTest.gatherJVMStats(logstashTest.URL+jvmStats, &logstash6accJVMStats)
	require.NoError(test, err, "Can't gather JVM stats")

	logstash6accJVMStats.AssertContainsTaggedFields(
		test,
		"logstash_jvm",
		map[string]interface{}{
			"mem_pools_young_max_in_bytes":                  float64(1605304320.0),
			"mem_pools_young_committed_in_bytes":            float64(71630848.0),
			"mem_heap_committed_in_bytes":                   float64(824963072.0),
			"threads_count":                                 float64(60.0),
			"mem_pools_old_peak_used_in_bytes":              float64(696572600.0),
			"mem_pools_old_peak_max_in_bytes":               float64(6583418880.0),
			"mem_heap_used_percent":                         float64(2.0),
			"gc_collectors_young_collection_time_in_millis": float64(107321.0),
			"mem_pools_survivor_committed_in_bytes":         float64(8912896.0),
			"mem_pools_young_used_in_bytes":                 float64(11775120.0),
			"mem_non_heap_committed_in_bytes":               float64(222986240.0),
			"mem_pools_survivor_peak_max_in_bytes":          float64(200605696),
			"mem_pools_young_peak_max_in_bytes":             float64(1605304320.0),
			"uptime_in_millis":                              float64(281850926.0),
			"mem_pools_survivor_peak_used_in_bytes":         float64(8912896.0),
			"mem_pools_survivor_max_in_bytes":               float64(200605696.0),
			"gc_collectors_old_collection_count":            float64(37.0),
			"mem_pools_survivor_used_in_bytes":              float64(835008.0),
			"mem_pools_old_used_in_bytes":                   float64(189750576.0),
			"mem_pools_old_max_in_bytes":                    float64(6583418880.0),
			"mem_pools_young_peak_used_in_bytes":            float64(71630848.0),
			"mem_heap_used_in_bytes":                        float64(202360704.0),
			"mem_heap_max_in_bytes":                         float64(8389328896.0),
			"gc_collectors_young_collection_count":          float64(2094.0),
			"threads_peak_count":                            float64(62.0),
			"mem_pools_old_committed_in_bytes":              float64(744419328.0),
			"gc_collectors_old_collection_time_in_millis":   float64(7492.0),
			"mem_non_heap_used_in_bytes":                    float64(197878896.0),
		},
		map[string]string{
			"node_id":      string("3044f675-21ce-4335-898a-8408aa678245"),
			"node_name":    string("node-6-test"),
			"source":       string("node-6"),
			"node_version": string("6.4.2"),
		},
	)
}

func Test_Logstash7GatherPipelinesQueueStats(test *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(writer, "%s", string(logstash7PipelinesJSON))
		if err != nil {
			test.Logf("Can't print test json")
		}
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	if err != nil {
		test.Logf("Can't connect to: %s", logstashTest.URL)
	}
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()

		if err != nil {
			test.Logf("Can't createHTTPClient")
		}
		logstashTest.client = client
	}

	if err := logstashTest.gatherPipelinesStats(logstashTest.URL+pipelineStats, &logstash7accPipelinesStats); err != nil {
		test.Logf("Can't gather Pipeline stats")
	}

	fields := make(map[string]interface{})
	fields["duration_in_millis"] = float64(3032875.0)
	fields["queue_push_duration_in_millis"] = float64(13300.0)
	fields["in"] = float64(2665549.0)
	fields["filtered"] = float64(2665549.0)
	fields["out"] = float64(2665549.0)

	logstash7accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_events",
		fields,
		map[string]string{
			"node_id":      string("28580380-ad2c-4032-934b-76359125edca"),
			"node_name":    string("HOST01.local"),
			"source":       string("HOST01.local"),
			"node_version": string("7.4.2"),
			"pipeline":     string("infra"),
		},
	)

	logstash7accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(2802177.0),
			"in":                 float64(2665549.0),
			"out":                float64(2665549.0),
		},
		map[string]string{
			"node_id":      string("28580380-ad2c-4032-934b-76359125edca"),
			"node_name":    string("HOST01.local"),
			"source":       string("HOST01.local"),
			"node_version": string("7.4.2"),
			"pipeline":     string("infra"),
			"plugin_name":  string("elasticsearch"),
			"plugin_id":    string("38967f09bbd2647a95aa00702b6b557bdbbab31da6a04f991d38abe5629779e3"),
			"plugin_type":  string("output"),
		},
	)
	logstash7accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"bulk_requests_successes":     float64(2870),
			"bulk_requests_responses_200": float64(2870),
			"bulk_requests_failures":      float64(262),
			"bulk_requests_with_errors":   float64(9089),
		},
		map[string]string{
			"node_id":      string("28580380-ad2c-4032-934b-76359125edca"),
			"node_name":    string("HOST01.local"),
			"source":       string("HOST01.local"),
			"node_version": string("7.4.2"),
			"pipeline":     string("infra"),
			"plugin_name":  string("elasticsearch"),
			"plugin_id":    string("38967f09bbd2647a95aa00702b6b557bdbbab31da6a04f991d38abe5629779e3"),
			"plugin_type":  string("output"),
		},
	)
	logstash7accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_plugins",
		map[string]interface{}{
			"documents_successes":          float64(2665549),
			"documents_retryable_failures": float64(13733),
		},
		map[string]string{
			"node_id":      string("28580380-ad2c-4032-934b-76359125edca"),
			"node_name":    string("HOST01.local"),
			"source":       string("HOST01.local"),
			"node_version": string("7.4.2"),
			"pipeline":     string("infra"),
			"plugin_name":  string("elasticsearch"),
			"plugin_id":    string("38967f09bbd2647a95aa00702b6b557bdbbab31da6a04f991d38abe5629779e3"),
			"plugin_type":  string("output"),
		},
	)

	logstash7accPipelinesStats.AssertContainsTaggedFields(
		test,
		"logstash_queue",
		map[string]interface{}{
			"events":                  float64(0),
			"max_queue_size_in_bytes": float64(4294967296),
			"queue_size_in_bytes":     float64(32028566),
		},
		map[string]string{
			"node_id":      string("28580380-ad2c-4032-934b-76359125edca"),
			"node_name":    string("HOST01.local"),
			"source":       string("HOST01.local"),
			"node_version": string("7.4.2"),
			"pipeline":     string("infra"),
			"queue_type":   string("persisted"),
		},
	)
}
