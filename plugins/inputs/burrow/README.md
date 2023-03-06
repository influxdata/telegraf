# Burrow Kafka Consumer Lag Checking Input Plugin

Collect Kafka topic, consumer and partition status via
[Burrow](https://github.com/linkedin/Burrow) HTTP
[API](https://github.com/linkedin/Burrow/wiki/HTTP-Endpoint).

Supported Burrow version: `1.x`

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Collect Kafka topics and consumers status from Burrow HTTP API.
[[inputs.burrow]]
  ## Burrow API endpoints in format "schema://host:port".
  ## Default is "http://localhost:8000".
  servers = ["http://localhost:8000"]

  ## Override Burrow API prefix.
  ## Useful when Burrow is behind reverse-proxy.
  # api_prefix = "/v3/kafka"

  ## Maximum time to receive response.
  # response_timeout = "5s"

  ## Limit per-server concurrent connections.
  ## Useful in case of large number of topics or consumer groups.
  # concurrent_connections = 20

  ## Filter clusters, default is no filtering.
  ## Values can be specified as glob patterns.
  # clusters_include = []
  # clusters_exclude = []

  ## Filter consumer groups, default is no filtering.
  ## Values can be specified as glob patterns.
  # groups_include = []
  # groups_exclude = []

  ## Filter topics, default is no filtering.
  ## Values can be specified as glob patterns.
  # topics_include = []
  # topics_exclude = []

  ## Credentials for basic HTTP authentication.
  # username = ""
  # password = ""

  ## Optional SSL config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  # insecure_skip_verify = false
```

## Group/Partition Status mappings

* `OK` = 1
* `NOT_FOUND` = 2
* `WARN` = 3
* `ERR` = 4
* `STOP` = 5
* `STALL` = 6

> unknown value will be mapped to 0

## Metrics

### Fields

* `burrow_group` (one event per each consumer group)
  * status (string, see Partition Status mappings)
  * status_code (int, `1..6`, see Partition status mappings)
  * partition_count (int, `number of partitions`)
  * offset (int64, `total offset of all partitions`)
  * total_lag (int64, `totallag`)
  * lag (int64, `maxlag.current_lag || 0`)
  * timestamp (int64, `end.timestamp`)

* `burrow_partition` (one event per each topic partition)
  * status (string, see Partition Status mappings)
  * status_code (int, `1..6`, see Partition status mappings)
  * lag (int64, `current_lag || 0`)
  * offset (int64, `end.timestamp`)
  * timestamp (int64, `end.timestamp`)

* `burrow_topic` (one event per topic offset)
  * offset (int64)

### Tags

* `burrow_group`
  * cluster (string)
  * group (string)

* `burrow_partition`
  * cluster (string)
  * group (string)
  * topic (string)
  * partition (int)
  * owner (string)

* `burrow_topic`
  * cluster (string)
  * topic (string)
  * partition (int)

## Example Output
