# Telegraf Plugin: Burrow

Collect Kafka topics and consumers status
from [Burrow](https://github.com/linkedin/Burrow) HTTP API.

### Configuration:

```
[[inputs.burrow]]
  ## Burrow endpoints in format "sheme://[user:password@]host:port"
  ## e.g.
  ##   servers = ["http://localhost:8080"]
  ##   servers = ["https://example.com:8000"]
  ##   servers = ["http://user:pass@example.com:8000"]
  ##
  servers = [ "http://127.0.0.1:8000" ]

  ## Prefix all HTTP API requests.
  #api_prefix = "/v2/kafka"

  ## Maximum time to receive response.
  #timeout = "5s"

  ## Optional, gather info only about specific clusters.
  ## Default is gather all.
  #clusters = ["clustername1"]

  ## Optional, gather stats only about specific groups.
  ## Default is gather all.
  #groups = ["group1"]

  ## Optional, gather info only about specific topics.
  ## Default is gather all
  #topics = ["topicA"]

  ## Concurrent connections limit (per server), default is 4.
  #max_concurrent_connections = 10

  ## Internal working queue adjustments (per measurement, per server), default is 4.
  #worker_queue_length = 5

  ## Credentials for basic HTTP authentication.
  #username = ""
  #password = ""

  ## Optional SSL config
  #ssl_ca = "/etc/telegraf/ca.pem"
  #ssl_cert = "/etc/telegraf/cert.pem"
  #ssl_key = "/etc/telegraf/key.pem"

  ## Use SSL but skip chain & host verification
  #insecure_skip_verify = false
```

Due to the nature of Burrow API (REST), each topic or consumer metric
collection requires an HTTP request, so, in order to keep things running
smooth, consider two parameters:

1. `max_concurrent_connection` - limit maximum number of concurrent HTTP
requests (per server).
2. `worker_queue_length` - number of concurrent workers processing
each measurement (per measurement, per server).

Just keep in mind, each worker in queue requires an HTTP connection,
so keep `max_concurrent_connection` and `worker_queue_length` balanced
in ratio 2:1.

### Partition Status mappings

* OK = 1
* NOT_FOUND = 2
* WARN = 3
* ERR = 4
* STOP = 5
* STALL = 6

### Measurements & Fields:

- burrow_topic (one event for each topic offset)
  - offset (int64)

- burrow_consumer
  - start.offset (int64)
  - start.lag (int64)
  - start.timestamp (int64)
  - end.offset (int64)
  - end.lag (int64)
  - end.timestamp (int64)
  - status (1..6, see Partition status mappings)

### Tags

- burrow_topic
  - cluster
  - topic
  - partition

- burrow_consumer
  - cluster
  - group
  - topic
  - partition
