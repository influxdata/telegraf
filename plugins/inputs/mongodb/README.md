# Telegraf plugin: MongoDB

#### Configuration

```toml
[[inputs.mongodb]]
  ## An array of URI to gather stats about. Specify an ip or hostname
  ## with optional port add password. ie,
  ##   mongodb://user:auth_key@10.10.3.30:27017,
  ##   mongodb://10.10.3.33:18832,
  ##   10.0.0.1:10000, etc.
  servers = ["127.0.0.1:27017"]
```

For authenticated mongodb istances use connection mongdb connection URI

```toml
[[inputs.mongodb]]
  servers = ["mongodb://username:password@10.XX.XX.XX:27101/mydatabase?authSource=admin"]
```
This connection uri may be different based on your environement and mongodb
setup. If the user doesn't have the required privilege to execute serverStatus 
command the you will get this error on telegraf

```
Error in input [mongodb]: not authorized on admin to execute command { serverStatus: 1, recordStats: 0 }
```

#### Description

The telegraf plugin collects mongodb stats exposed by serverStatus and few more
and create a single measurement containing values e.g.
 * active_reads
 * active_writes
 * commands_per_sec
 * deletes_per_sec
 * flushes_per_sec
 * getmores_per_sec
 * inserts_per_sec
 * net_in_bytes
 * net_out_bytes
 * open_connections
 * percent_cache_dirty
 * percent_cache_used
 * queries_per_sec
 * queued_reads
 * queued_writes
 * resident_megabytes
 * updates_per_sec
 * vsize_megabytes
 * ttl_deletes_per_sec
 * ttl_passes_per_sec
 * repl_lag
 * jumbo_chunks (only if mongos or mongo config)
