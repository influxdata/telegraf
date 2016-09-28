# Telegraf Plugin: Redis

### Configuration:

```
# Read Redis's basic status information
[[inputs.redis]]
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://:password@192.168.99.100
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]
```

### Measurements & Fields:

- Measurement
    - uptime_in_seconds
    - connected_clients
    - used_memory
    - used_memory_rss
    - used_memory_peak
    - used_memory_lua
    - rdb_changes_since_last_save
    - total_connections_received
    - total_commands_processed
    - instantaneous_ops_per_sec
    - instantaneous_input_kbps
    - instantaneous_output_kbps
    - sync_full
    - sync_partial_ok
    - sync_partial_err
    - expired_keys
    - evicted_keys
    - keyspace_hits
    - keyspace_misses
    - pubsub_channels
    - pubsub_patterns
    - latest_fork_usec
    - connected_slaves
    - master_repl_offset
    - master_last_io_seconds_ago
    - repl_backlog_active
    - repl_backlog_size
    - repl_backlog_histlen
    - mem_fragmentation_ratio
    - used_cpu_sys
    - used_cpu_user
    - used_cpu_sys_children
    - used_cpu_user_children

### Tags:

- All measurements have the following tags:
    - port
    - server
    - replication role

### Example Output:

Using this configuration:
```
[[inputs.redis]]
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://:password@192.168.99.100
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]
```

When run with:
```
./telegraf -config telegraf.conf -input-filter redis -test
```

It produces:
```
* Plugin: redis, Collection 1
> redis,port=6379,server=localhost clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i 1457052084987848383
```
