# RethinkDB Input Plugin

This plugin collects metrics from [RethinkDB][rethinkdb] servers.

⭐ Telegraf v0.1.3
🏷️ server
💻 all

[rethinkdb]: https://www.rethinkdb.com/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from one or many RethinkDB servers
[[inputs.rethinkdb]]
  ## An array of URI to gather stats about. Specify an ip or hostname
  ## with optional port add password. ie,
  ##   rethinkdb://user:auth_key@10.10.3.30:28105,
  ##   rethinkdb://10.10.3.33:18832,
  ## For rethinkdb v2.3.0+ with username/password authorization you should use
  ##   rethinkdb2://username:password@127.0.0.1:28015"
  servers = ["127.0.0.1:28015"]
```

## Metrics

- rethinkdb
  - tags:
    - type
    - ns
    - rethinkdb_host
    - rethinkdb_hostname
  - fields:
    - cache_bytes_in_use (integer, bytes)
    - disk_read_bytes_per_sec (integer, reads)
    - disk_read_bytes_total (integer, bytes)
    - disk_written_bytes_per_sec (integer, bytes)
    - disk_written_bytes_total (integer, bytes)
    - disk_usage_data_bytes (integer, bytes)
    - disk_usage_garbage_bytes (integer, bytes)
    - disk_usage_metadata_bytes (integer, bytes)
    - disk_usage_preallocated_bytes (integer, bytes)

- rethinkdb_engine
  - tags:
    - type
    - ns
    - rethinkdb_host
    - rethinkdb_hostname
  - fields:
    - active_clients (integer, clients)
    - clients (integer, clients)
    - queries_per_sec (integer, queries)
    - total_queries (integer, queries)
    - read_docs_per_sec (integer, reads)
    - total_reads (integer, reads)
    - written_docs_per_sec (integer, writes)
    - total_writes (integer, writes)

## Example Output
