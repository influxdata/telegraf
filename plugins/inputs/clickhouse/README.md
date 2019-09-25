# Telegraf Input Plugin: ClickHouse

This plugin gathers the statistic data from [ClickHouse](https://github.com/ClickHouse/ClickHouse)  server.

### Configuration
```ini
# Read metrics from one or many ClickHouse servers
[[inputs.clickhouse]]
  timeout         = 5 # seconds
  servers         = ["http://username:password@127.0.0.1:8123"]
  auto_discovery  = true # If a setting is "true" plugin tries to connect to all servers in the cluster (system.clusters)
  cluster_include = []
  cluster_exclude = ["test_shard_localhost"]
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
