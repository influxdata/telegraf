# Telegraf Input Plugin: ClickHouse

This plugin gathers the statistic data from [ClickHouse](https://github.com/yandex/ClickHouse)  server.

### Configuration
```
# Read metrics from one or many ClickHouse servers
[[inputs.clickhouse]]
  dsn     = "native://localhost:9000?username=user&password=qwerty"
  cluster = true # If a setting is "true" plugin tries to connect to all servers in the cluster (system.clusters)
  ignored_clusters = ["test_shard_localhost"] ## ignored cluster names
```

### Metrics:
- clickhouse_events
  - tags:
    - hostname (ClickHouse server hostname)
    - cluster (Name of the cluster [optional])
    - shard_num (Shard number in the cluster [optional])
  - fields:
    - all rows from system.events

- clickhouse_metrics
  - tags:
    - hostname (ClickHouse server hostname)
    - cluster (Name of the cluster [optional])
    - shard_num (Shard number in the cluster [optional])
  - fields:
    - all rows from system.metrics

- clickhouse_asynchronous_metrics
  - tags:
    - hostname (ClickHouse server hostname)
    - cluster (Name of the cluster [optional])
    - shard_num (Shard number in the cluster [optional])
  - fields:
    - all rows from system.asynchronous_metrics

- clickhouse_tables
  - tags:
    - hostname (ClickHouse server hostname)
    - table
    - database
    - cluster (Name of the cluster [optional])
    - shard_num (Shard number in the cluster [optional])
  - fields:
    - bytes
    - parts
    - rows

