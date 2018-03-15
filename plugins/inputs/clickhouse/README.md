# Telegraf Input Plugin: ClickHouse

This [ClickHouse](https://github.com/yandex/ClickHouse) plugin provides metrics for your ClickHouse server.

### Configuration example:
```
[[inputs.clickhouse]]
  dsn = "native://localhost:9000?username=user&password=qwerty"
```

### Metrics:
- clickhouse_events
  - tags:
    - server (ClickHouse server hostname)
    - hostname (Telegraf agent hostname)
  - fields:
    - all rows from system.events

- clickhouse_metrics
  - tags:
    - server (ClickHouse server hostname)
    - hostname (Telegraf agent hostname)
  - fields:
    - all rows from system.metrics

- clickhouse_asynchronous_metrics
  - tags:
    - server (ClickHouse server hostname)
    - hostname (Telegraf agent hostname)
  - fields:
    - all rows from system.asynchronous_metrics

- clickhouse_tables
  - tags:
    - server (ClickHouse server hostname)
    - hostname (Telegraf agent hostname)
    - table
    - database
  - fields:
    - bytes
    - parts
    - rows

