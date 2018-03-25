package logstash

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

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

const pipelineJSON = `

{
  "host" : "de22454128d2",
  "version" : "6.1.1",
  "http_address" : "0.0.0.0:9600",
  "id" : "e7bfb05c-ebd9-47b1-bc59-39e90ce98fcd",
  "name" : "de22454128d2",
  "pipelines" : {
    ".monitoring-logstash" : {
      "events" : null,
      "plugins" : {
        "inputs" : [ ],
        "filters" : [ ],
        "outputs" : [ ]
      },
      "reloads" : {
        "last_error" : null,
        "successes" : 0,
        "last_success_timestamp" : null,
        "last_failure_timestamp" : null,
        "failures" : 0
      },
      "queue" : null
    },
    "main" : {
      "events" : {
        "duration_in_millis" : 1151,
        "in" : 1269,
        "out" : 1269,
        "filtered" : 1269,
        "queue_push_duration_in_millis" : 1324
      },
      "plugins" : {
				"inputs" : [ {
					"id" : "a35197a509596954e905e38521bae12b1498b17d-1",
					"events" : {
						"out" : 2,
						"queue_push_duration_in_millis" : 32
					},
					"name" : "beats"
				} ],
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
      }
    }
  }
}
`

var logstashTest = &Logstash{
	URL: "http://localhost:9600",
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
	requestURL, err := url.Parse(logstashTest.URL)
	if err != nil {
		t.Logf("Can't connect to: %s", logstashTest.URL)
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

	if err := logstashTest.gatherPipelineStats(logstashTest.URL+pipelineStats, &accPipelineStats); err != nil {
		t.Logf("Can't gather Pipeline stats")
	}

	accPipelineStats.AssertContainsFields(
		t,
		"logstash_events",
		map[string]interface{}{
			"duration_in_millis":            float64(1151.0),
			"queue_push_duration_in_millis": float64(1324.0),
			"in":       float64(1269.0),
			"filtered": float64(1269.0),
			"out":      float64(1269.0),
		})

	accPipelineStats.AssertContainsTaggedFields(
		t,
		"logstash_plugins",
		map[string]interface{}{
			"queue_push_duration_in_millis": float64(32.0),
			"duration_in_millis":            float64(0.0),
			"in":                            float64(0.0),
			"out":                           float64(2.0),
		},
		map[string]string{
			"plugin": string("beats"),
			"type":   string("input"),
		})

	accPipelineStats.AssertContainsTaggedFields(
		t,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(360.0),
			"in":                 float64(1269.0),
			"out":                float64(1269.0),
		},
		map[string]string{
			"plugin": string("stdout"),
			"type":   string("output"),
		})

	accPipelineStats.AssertContainsTaggedFields(
		t,
		"logstash_plugins",
		map[string]interface{}{
			"duration_in_millis": float64(228.0),
			"in":                 float64(1269.0),
			"out":                float64(1269.0),
		},
		map[string]string{
			"plugin": string("s3"),
			"type":   string("output"),
		})
}

func Test_gatherProcessStats(t *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", string(processJSON))
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	if err != nil {
		t.Logf("Can't connect to: %s", logstashTest.URL)
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

	if err := logstashTest.gatherProcessStats(logstashTest.URL+processStats, &accProcessStats); err != nil {
		t.Logf("Can't gather Process stats")
	}

	accProcessStats.AssertContainsFields(
		t,
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
		})
}

func Test_gatherJVMStats(t *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", string(jvmJSON))
	}))
	requestURL, err := url.Parse(logstashTest.URL)
	if err != nil {
		t.Logf("Can't connect to: %s", logstashTest.URL)
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

	if err := logstashTest.gatherJVMStats(logstashTest.URL+jvmStats, &accJVMStats); err != nil {
		t.Logf("Can't gather JVM stats")
	}

	accJVMStats.AssertContainsFields(
		t,
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
		})
}
