package logstash

const logstash5ProcessJSON = `
{
  "host" : "node-5",
  "version" : "5.3.0",
  "http_address" : "0.0.0.0:9600",
  "id" : "a360d8cf-6289-429d-8419-6145e324b574",
  "name" : "node-5-test",
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

const logstash5JvmJSON = `
{
  "host" : "node-5",
  "version" : "5.3.0",
  "http_address" : "0.0.0.0:9600",
  "id" : "a360d8cf-6289-429d-8419-6145e324b574",
  "name" : "node-5-test",
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

const logstash5PipelineJSON = `
{
  "host" : "node-5",
  "version" : "5.3.0",
  "http_address" : "0.0.0.0:9600",
  "id" : "a360d8cf-6289-429d-8419-6145e324b574",
  "name" : "node-5-test",
  "pipeline" : {
    "events" : {
      "duration_in_millis" : 1151,
      "in" : 1269,
      "filtered" : 1269,
      "out" : 1269
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
      "events" : 208,
      "type" : "persisted",
      "capacity" : {
        "page_capacity_in_bytes" : 262144000,
        "max_queue_size_in_bytes" : 8589934592,
        "max_unread_events" : 0
      },
      "data" : {
        "path" : "/path/to/data/queue",
        "free_space_in_bytes" : 89280552960,
        "storage_type" : "hfs"
      }
    },
    "id" : "main"
  }
}
`
