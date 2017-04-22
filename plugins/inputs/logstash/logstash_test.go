package logstash

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

const processStatsCount = 18
const processJSON = `
{
  "host" : "f7a354acc27f",
  "version" : "5.3.0",
  "http_address" : "0.0.0.0:9600",
  "id" : "a360d8cf-6289-429d-8419-6145e324b574",
  "name" : "f7a354acc27f",
  "process" : {
    "open_file_descriptors" : 89,
    "peak_open_file_descriptors" : 100,
    "max_file_descriptors" : 1048576,
    "mem" : {
      "total_virtual_in_bytes" : 4809506816
    },
    "cpu" : {
      "total_in_millis" : 155260000000,
      "percent" : 3,
      "load_average" : {
        "1m" : 0.49,
        "5m" : 0.61,
        "15m" : 0.54
      }
    }
  }
}
`

var processStatsExpected = map[string]interface{}{
	"peak_open_file_descriptors": 100.0,
	"max_file_descriptors":       1.048576e+06,
	"mem_total_virtual_in_bytes": 4.809506816e+09,
	"cpu_total_in_millis":        1.5526e+11,
	"cpu_percent":                3.0,
	"cpu_load_average_5m":        0.61,
	"open_file_descriptors":      89.0,
	"cpu_load_average_1m":        0.49,
	"cpu_load_average_15m":       0.54,
}

const jvmStatsCount = 56
const jvmJSON = `
{
  "host" : "f7a354acc27f",
  "version" : "5.3.0",
  "http_address" : "0.0.0.0:9600",
  "id" : "a360d8cf-6289-429d-8419-6145e324b574",
  "name" : "f7a354acc27f",
  "jvm" : {
    "threads" : {
      "count" : 29,
      "peak_count" : 31
    },
    "mem" : {
      "heap_used_in_bytes" : 341270784,
      "heap_used_percent" : 16,
      "heap_committed_in_bytes" : 519045120,
      "heap_max_in_bytes" : 2077753344,
      "non_heap_used_in_bytes" : 268905936,
      "non_heap_committed_in_bytes" : 291487744,
      "pools" : {
        "survivor" : {
          "peak_used_in_bytes" : 8912896,
          "used_in_bytes" : 9419672,
          "peak_max_in_bytes" : 34865152,
          "max_in_bytes" : 69730304,
          "committed_in_bytes" : 17825792
        },
        "old" : {
          "peak_used_in_bytes" : 127900864,
          "used_in_bytes" : 255801728,
          "peak_max_in_bytes" : 724828160,
          "max_in_bytes" : 1449656320,
          "committed_in_bytes" : 357957632
        },
        "young" : {
          "peak_used_in_bytes" : 71630848,
          "used_in_bytes" : 76049384,
          "peak_max_in_bytes" : 279183360,
          "max_in_bytes" : 558366720,
          "committed_in_bytes" : 143261696
        }
      }
    },
    "gc" : {
      "collectors" : {
        "old" : {
          "collection_time_in_millis" : 114,
          "collection_count" : 2
        },
        "young" : {
          "collection_time_in_millis" : 3235,
          "collection_count" : 616
        }
      }
    },
    "uptime_in_millis" : 4803461
  }
}
`

var jvmStatsExpected = map[string]interface{}{
	"mem_pools_young_max_in_bytes":                  5.5836672e+08,
	"mem_pools_young_committed_in_bytes":            1.43261696e+08,
	"mem_heap_committed_in_bytes":                   5.1904512e+08,
	"threads_count":                                 29.0,
	"mem_pools_old_peak_used_in_bytes":              1.27900864e+08,
	"mem_pools_old_peak_max_in_bytes":               7.2482816e+08,
	"mem_heap_used_percent":                         16.0,
	"gc_collectors_young_collection_time_in_millis": 3235.0,
	"mem_pools_survivor_committed_in_bytes":         1.7825792e+07,
	"mem_pools_young_used_in_bytes":                 7.6049384e+07,
	"mem_non_heap_committed_in_bytes":               2.91487744e+08,
	"mem_pools_survivor_peak_max_in_bytes":          3.4865152e+07,
	"mem_pools_young_peak_max_in_bytes":             2.7918336e+08,
	"uptime_in_millis":                              4.803461e+06,
	"mem_pools_survivor_peak_used_in_bytes":         8.912896e+06,
	"mem_pools_survivor_max_in_bytes":               6.9730304e+07,
	"gc_collectors_old_collection_count":            2.0,
	"mem_pools_survivor_used_in_bytes":              9.419672e+06,
	"mem_pools_old_used_in_bytes":                   2.55801728e+08,
	"mem_pools_old_max_in_bytes":                    1.44965632e+09,
	"mem_pools_young_peak_used_in_bytes":            7.1630848e+07,
	"mem_heap_used_in_bytes":                        3.41270784e+08,
	"mem_heap_max_in_bytes":                         2.077753344e+09,
	"gc_collectors_young_collection_count":          616.0,
	"threads_peak_count":                            31.0,
	"mem_pools_old_committed_in_bytes":              3.57957632e+08,
	"gc_collectors_old_collection_time_in_millis":   114.0,
	"mem_non_heap_used_in_bytes":                    2.68905936e+08,
}

const pipelineStatsCount = 24
const pipelineJSON = `
{
  "host" : "f7a354acc27f",
  "version" : "5.3.0",
  "http_address" : "0.0.0.0:9600",
  "id" : "a360d8cf-6289-429d-8419-6145e324b574",
  "name" : "f7a354acc27f",
  "pipeline" : {
    "events" : {
      "duration_in_millis" : 1151,
      "in" : 1269,
      "filtered" : 1269,
      "out" : 1269
    },
    "plugins" : {
      "inputs" : [ ],
      "filters" : [ ],
      "outputs" : [ {
        "id" : "582d5c2becb582a053e1e9a6bcc11d49b69a6dfd-3",
        "events" : {
          "duration_in_millis" : 228,
          "in" : 1269,
          "out" : 1269
        },
        "name" : "s3"
      }, {
        "id" : "582d5c2becb582a053e1e9a6bcc11d49b69a6dfd-2",
        "events" : {
          "duration_in_millis" : 360,
          "in" : 1269,
          "out" : 1269
        },
        "name" : "stdout"
      } ]
    },
    "reloads" : {
      "last_error" : null,
      "successes" : 0,
      "last_success_timestamp" : null,
      "last_failure_timestamp" : null,
      "failures" : 0
    },
    "queue" : {
      "type" : "memory"
    },
    "id" : "main"
  }
}
`

var eventsStatsExpected = map[string]interface{}{
	"duration_in_millis": 1151.0,
	"in":                 1269.0,
	"filtered":           1269.0,
	"out":                1269.0,
}

var outputS3StatsExpected = map[string]interface{}{
	"name":               "s3",
	"duration_in_millis": 228,
	"in":                 1269,
	"out":                1269,
}

var outputStdoutStatsExpected = map[string]interface{}{
	"in":                 1269,
	"out":                1269,
	"name":               "stdout",
	"duration_in_millis": 360,
}

var logstashTest = &Logstash{
	LogstashURL: "http://localhost:9600",
}

var (
	accPipelineStats testutil.Accumulator
	accProcessStats  testutil.Accumulator
	accJVMStats      testutil.Accumulator
)

func Test_gatherPipelineStats(t *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", string(pipelineJSON))
	}))
	requestURL, err := url.Parse(logstashTest.LogstashURL)
	if err != nil {
		t.Logf("Can't connect to: %s", logstashTest.LogstashURL)
	}
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()

		if err != nil {
			t.Logf("Can't createHttpClient")
		}
		logstashTest.client = client
	}

	if err := logstashTest.gatherPipelineStats(logstashTest.LogstashURL+pipelineStats, &accPipelineStats); err != nil {
		t.Logf("Can't gather Pipeline stats")
	}

	if !accPipelineStats.HasMeasurement("logstash_events") {
		t.Errorf("acc.HasMeasurement: expected logstash_events")
	}

	if !accPipelineStats.HasMeasurement("logstash_plugin_output_stdout") {
		t.Errorf("acc.HasMeasurement: expected logstash_plugin_output_stdout")
	}

	return
}

func Test_gatherProcessStats(t *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", string(processJSON))
	}))
	requestURL, err := url.Parse(logstashTest.LogstashURL)
	if err != nil {
		t.Logf("Can't connect to: %s", logstashTest.LogstashURL)
	}
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()

		if err != nil {
			t.Logf("Can't createHttpClient")
		}
		logstashTest.client = client
	}

	if err := logstashTest.gatherProcessStats(logstashTest.LogstashURL+processStats, &accProcessStats); err != nil {
		t.Logf("Can't gather JVM stats")
	}

	if !accProcessStats.HasMeasurement("logstash_process") {
		t.Errorf("acc.HasMeasurement: expected logstash_process")
	}

	return
}

func Test_gatherJVMStats(t *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", string(jvmJSON))
	}))
	requestURL, err := url.Parse(logstashTest.LogstashURL)
	if err != nil {
		t.Logf("Can't connect to: %s", logstashTest.LogstashURL)
	}
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if logstashTest.client == nil {
		client, err := logstashTest.createHTTPClient()

		if err != nil {
			t.Logf("Can't createHttpClient")
		}
		logstashTest.client = client
	}

	if err := logstashTest.gatherJVMStats(logstashTest.LogstashURL+jvmStats, &accJVMStats); err != nil {
		t.Logf("Can't gather JVM stats")
	}

	if !accJVMStats.HasMeasurement("logstash_jvm") {
		t.Errorf("acc.HasMeasurement: expected logstash_jvm")
	}

	return
}

func Test_Gather(t *testing.T) {

	Test_gatherJVMStats(t)
	Test_gatherPipelineStats(t)
	Test_gatherProcessStats(t)

	//Tests for processStats
	assert.Equal(t, accProcessStats.NFields(), processStatsCount)
	accProcessStats.AssertContainsFields(t, "logstash_process", processStatsExpected)

	//Tests for pipelineStats
	assert.Equal(t, accPipelineStats.NFields(), pipelineStatsCount)

	accPipelineStats.AssertContainsFields(t,
		"logstash_events",
		eventsStatsExpected)

	accPipelineStats.AssertContainsFields(t,
		"logstash_plugin_output_s3",
		outputS3StatsExpected)

	accPipelineStats.AssertContainsFields(t,
		"logstash_plugin_output_stdout",
		outputStdoutStatsExpected)

	//Test for jvmStats
	assert.Equal(t, accJVMStats.NFields(), jvmStatsCount)
	accJVMStats.AssertContainsFields(t, "logstash_jvm", jvmStatsExpected)

	if testing.Short() {
		t.Skip("Skipping Gather function test")
	}
}
