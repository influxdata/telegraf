package logstash

const logstash6ProcessJSON = `
{
  "host" : "node-6",
  "version" : "6.4.2",
  "http_address" : "127.0.0.1:9600",
  "id" : "3044f675-21ce-4335-898a-8408aa678245",
  "name" : "node-6-test",
  "process" : {
    "open_file_descriptors" : 133,
    "peak_open_file_descriptors" : 145,
    "max_file_descriptors" : 262144,
    "mem" : {
      "total_virtual_in_bytes" : 17923452928
    },
    "cpu" : {
      "total_in_millis" : 5841460,
      "percent" : 0,
      "load_average" : {
        "1m" : 48.2,
        "5m" : 42.4,
        "15m" : 38.95
      }
    }
  }
}
`
const logstash6JvmJSON = `
{
  "host" : "node-6",
  "version" : "6.4.2",
  "http_address" : "127.0.0.1:9600",
  "id" : "3044f675-21ce-4335-898a-8408aa678245",
  "name" : "node-6-test",
  "jvm" : {
    "threads" : {
      "count" : 60,
      "peak_count" : 62
    },
    "mem" : {
      "heap_used_percent" : 2,
      "heap_committed_in_bytes" : 824963072,
      "heap_max_in_bytes" : 8389328896,
      "heap_used_in_bytes" : 202360704,
      "non_heap_used_in_bytes" : 197878896,
      "non_heap_committed_in_bytes" : 222986240,
      "pools" : {
        "survivor" : {
          "peak_used_in_bytes" : 8912896,
          "used_in_bytes" : 835008,
          "peak_max_in_bytes" : 200605696,
          "max_in_bytes" : 200605696,
          "committed_in_bytes" : 8912896
        },
        "old" : {
          "peak_used_in_bytes" : 696572600,
          "used_in_bytes" : 189750576,
          "peak_max_in_bytes" : 6583418880,
          "max_in_bytes" : 6583418880,
          "committed_in_bytes" : 744419328
        },
        "young" : {
          "peak_used_in_bytes" : 71630848,
          "used_in_bytes" : 11775120,
          "peak_max_in_bytes" : 1605304320,
          "max_in_bytes" : 1605304320,
          "committed_in_bytes" : 71630848
        }
      }
    },
    "gc" : {
      "collectors" : {
        "old" : {
          "collection_time_in_millis" : 7492,
          "collection_count" : 37
        },
        "young" : {
          "collection_time_in_millis" : 107321,
          "collection_count" : 2094
        }
      }
    },
    "uptime_in_millis" : 281850926
  }
}
`

const logstash6PipelinesJSON = `
{
  "host" : "node-6",
  "version" : "6.4.2",
  "http_address" : "127.0.0.1:9600",
  "id" : "3044f675-21ce-4335-898a-8408aa678245",
  "name" : "node-6-test",
  "pipelines" : {
    "main" : {
      "events" : {
        "duration_in_millis" : 8540751,
        "in" : 180659,
        "out" : 180659,
        "filtered" : 180659,
        "queue_push_duration_in_millis" : 366
      },
      "plugins" : {
        "inputs" : [
          {
            "id" : "input-kafka",
            "events" : {
              "out" : 180659,
              "queue_push_duration_in_millis" : 366
            },
            "name" : "kafka"
          }
        ],
        "filters" : [
          {
            "id" : "155b0ad18abbf3df1e0cb7bddef0d77c5ba699efe5a0f8a28502d140549baf54",
            "events" : {
              "duration_in_millis" : 2117,
              "in" : 27641,
              "out" : 27641
            },
            "name" : "mutate"
          },
          {
            "id" : "d079424bb6b7b8c7c61d9c5e0ddae445e92fa9ffa2e8690b0a669f7c690542f0",
            "events" : {
              "duration_in_millis" : 13149,
              "in" : 180659,
              "out" : 177549
            },
            "matches" : 177546,
            "failures" : 2,
            "name" : "date"
          },
          {
            "id" : "25afa60ab6dc30512fe80efa3493e4928b5b1b109765b7dc46a3e4bbf293d2d4",
            "events" : {
              "duration_in_millis" : 2814,
              "in" : 76602,
              "out" : 76602
            },
            "name" : "mutate"
          },
          {
            "id" : "2d9fa8f74eeb137bfa703b8050bad7d76636fface729e4585b789b5fc9bed668",
            "events" : {
              "duration_in_millis" : 9,
              "in" : 934,
              "out" : 934
            },
            "name" : "mutate"
          },
          {
            "id" : "4ed14c9ef0198afe16c31200041e98d321cb5c2e6027e30b077636b8c4842110",
            "events" : {
              "duration_in_millis" : 173,
              "in" : 3110,
              "out" : 0
            },
            "name" : "drop"
          },
          {
            "id" : "358ce1eb387de7cd5711c2fb4de64cd3b12e5ca9a4c45f529516bcb053a31df4",
            "events" : {
              "duration_in_millis" : 5605,
              "in" : 75482,
              "out" : 75482
            },
            "name" : "mutate"
          },
          {
            "id" : "82a9bbb02fff37a63c257c1f146b0a36273c7cbbebe83c0a51f086e5280bf7bb",
            "events" : {
              "duration_in_millis" : 313992,
              "in" : 180659,
              "out" : 180659
            },
            "name" : "csv"
          },
          {
            "id" : "8fb13a8cdd4257b52724d326aa1549603ffdd4e4fde6d20720c96b16238c18c3",
            "events" : {
              "duration_in_millis" : 0,
              "in" : 0,
              "out" : 0
            },
            "name" : "mutate"
          }
        ],
        "outputs" : [
          {
            "id" : "output-elk",
            "documents" : {
              "successes" : 221
            },
            "events" : {
              "duration_in_millis" : 651386,
              "in" : 177549,
              "out" : 177549
            },
            "bulk_requests" : {
              "successes" : 1,
              "responses" : {
                "200" : 748
              }
            },
            "name" : "elasticsearch"
          },
          {
            "id" : "output-kafka1",
            "events" : {
              "duration_in_millis" : 186751,
              "in" : 177549,
              "out" : 177549
            },
            "name" : "kafka"
          },
          {
            "id" : "output-kafka2",
            "events" : {
              "duration_in_millis" : 7335196,
              "in" : 177549,
              "out" : 177549
            },
            "name" : "kafka"
          }
        ]
      },
      "reloads" : {
        "last_error" : null,
        "successes" : 0,
        "last_success_timestamp" : null,
        "last_failure_timestamp" : null,
        "failures" : 0
      },
      "queue": {
        "events": 103,
        "type": "persisted",
        "capacity": {
          "queue_size_in_bytes": 1872391,
          "page_capacity_in_bytes": 67108864,
          "max_queue_size_in_bytes": 1073741824,
          "max_unread_events": 0
        },
        "data": {
          "path": "/var/lib/logstash/queue/main",
          "free_space_in_bytes": 36307369984,
          "storage_type": "ext4"
        }
      }
    }
  }
}
`
