package logstash

const logstash7PipelinesJSON = `
{
  "host" : "HOST01.local",
  "version" : "7.4.2",
  "http_address" : "127.0.0.1:9600",
  "id" : "28580380-ad2c-4032-934b-76359125edca",
  "name" : "HOST01.local",
  "ephemeral_id" : "bd95ff6b-3fa8-42ae-be32-098a4e4ea1ec", 
  "status" : "green", 
  "snapshot" : true, 
  "pipeline" : { 
    "workers" : 8, 
    "batch_size" : 125, 
    "batch_delay" : 50 
  },
  "pipelines" : {
    "infra" : {
      "events" : {
        "in" : 2665549,
        "out" : 2665549,
        "duration_in_millis" : 3032875,
        "filtered" : 2665549,
        "queue_push_duration_in_millis" : 13300
      },
      "plugins" : {
        "inputs" : [ {
          "id" : "8526dc80bc2257ab08f96018f96b0c68dd03abc5695bb22fb9e96339a8dfb4f86",
          "events" : {
            "out" : 2665549,
            "queue_push_duration_in_millis" : 13300
          },
          "peak_connections" : 1, 
          "name" : "beats", 
          "current_connections" : 1 
        } ],
        "codecs" : [ { 
          "id" : "plain_7312c097-1e7f-41db-983b-4f5a87a9eba2", 
          "encode" : { 
            "duration_in_millis" : 0, 
            "writes_in" : 0 
          }, 
          "name" : "plain", 
          "decode" : { 
            "out" : 0, 
            "duration_in_millis" : 0, 
            "writes_in" : 0 
          } 
        }, {
          "id" : "rubydebug_e958e3dc-10f6-4dd6-b7c5-ae3de2892afb",
          "encode" : {
            "duration_in_millis" : 0,
            "writes_in" : 0
          },
          "name" : "rubydebug",
          "decode" : {
            "out" : 0,
            "duration_in_millis" : 0,
            "writes_in" : 0
          }
        }, {
          "id" : "plain_addb97be-fb77-4cbc-b45c-0424cd5d0ac7",
          "encode" : {
            "duration_in_millis" : 0,
            "writes_in" : 0
          },
          "name" : "plain",
          "decode" : {
            "out" : 0,
            "duration_in_millis" : 0,
            "writes_in" : 0
          }
        } ],
        "filters" : [ {
          "id" : "9e8297a6ee7b61864f77853317dccde83d29952ef869010c385dcfc9064ab8b8",
          "events" : {
            "in" : 2665549,
            "out" : 2665549,
            "duration_in_millis" : 8648
          },
          "name" : "date",
          "matches" : 2665549 
        }, {
          "id" : "bec0c77b3f53a78c7878449c72ec59f97be31c1f12f9621f61ed2d4563bad869",
          "events" : {
            "in" : 2665549,
            "out" : 2665549,
            "duration_in_millis" : 195138
          },
          "name" : "fingerprint"
        } ],
        "outputs" : [ {
          "id" : "df59066a933f038354c1845ba44de692f70dbd0d2009ab07a12b98b776be7e3f",
          "events" : {
            "in" : 0,
            "out" : 0,
            "duration_in_millis" : 25
          },
          "name" : "stdout"
        }, {
          "id" : "38967f09bbd2647a95aa00702b6b557bdbbab31da6a04f991d38abe5629779e3",
          "events" : {
            "in" : 2665549,
            "out" : 2665549,
            "duration_in_millis" : 2802177
          },
          "name" : "elasticsearch",
          "bulk_requests" : {
            "successes" : 2870,
            "responses" : {
              "200" : 2870
            },
            "failures": 262,
            "with_errors": 9089
          },
          "documents" : {
            "successes" : 2665549,
            "retryable_failures": 13733
          }
        } ]
      },
      "reloads" : {
        "successes" : 4,
        "last_error" : null,
        "failures" : 0,
        "last_success_timestamp" : "2020-06-05T08:06:12.538Z",
        "last_failure_timestamp" : null
      },
      "queue" : {
        "type" : "persisted",
        "events_count" : 0,
        "queue_size_in_bytes" : 32028566,
        "max_queue_size_in_bytes" : 4294967296
      },
      "hash" : "5bc589ae4b02cb3e436626429b50928b9d99360639c84dc7fc69268ac01a9fd0", 
      "ephemeral_id" : "4bcacefa-6cbf-461e-b14e-184edd9ebdf3" 
    }
  }
}`
