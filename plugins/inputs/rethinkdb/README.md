# RethinkDB Input

Collect metrics from [RethinkDB](https://www.rethinkdb.com/).

### Configuration

```toml
[[inputs.rethinkdb]]
  ## An array of URI to gather stats about. Specify an ip or hostname
  ## with optional port add password. ie,
  ##   rethinkdb://user:auth_key@10.10.3.30:28105,
  ##   rethinkdb://10.10.3.33:18832,
  ##   10.0.0.1:10000, etc.
  servers = ["127.0.0.1:28015"]
  ##
  ## If you use actual rethinkdb of > 2.3.0 with username/password authorization,
  ## protocol have to be named "rethinkdb2" - it will use 1_0 H.
  # servers = ["rethinkdb2://username:password@127.0.0.1:28015"]
  ##
  ## If you use older versions of rethinkdb (<2.2) with auth_key, protocol
  ## have to be named "rethinkdb".
  # servers = ["rethinkdb://username:auth_key@127.0.0.1:28015"]
```

### Metrics

- rethinkdb
  - type
  - ns
  - rethinkdb_host
  - rethinkdb_hostname
    - cache_bytes_in_use
    - disk_read_bytes_per_sec
    - disk_read_bytes_total
    - disk_written_bytes_per_sec
    - disk_written_bytes_total
    - disk_usage_data_bytes
    - disk_usage_garbage_bytes
    - disk_usage_metadata_bytes
    - disk_usage_preallocated_bytes  
  
- rethinkdb_engine
  - type
  - ns
  - rethinkdb_host
  - rethinkdb_hostname
    - active_clients
    - clients
    - queries_per_sec
    - total_queries
    - read_docs_per_sec
    - total_reads
    - written_docs_per_sec
    - total_writes
