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

The plugin gathers the results of the [INFO](https://redis.io/commands/info) redis command.
There are two separate measurements: _redis_ and _redis\_keyspace_, the latter is used for gathering database related statistics.

Additionally the plugin also calculates the hit/miss ratio (keyspace\_hitrate) and the elapsed time since the last rdb save (rdb\_last\_save\_time\_elapsed).

- redis
    - keyspace_hitrate
    - rdb_last_save_time_elapsed

    - uptime
    - lru_clock

    - clients
    - client_longest_output_list
    - client_biggest_input_buf
    - blocked_clients

    - used_memory
    - used_memory_rss
    - used_memory_peak
    - total_system_memory
    - used_memory_lua
    - maxmemory
    - maxmemory_policy
    - mem_fragmentation_ratio

    - loading
    - rdb_changes_since_last_save
    - rdb_bgsave_in_progress
    - rdb_last_save_time
    - rdb_last_bgsave_status
    - rdb_last_bgsave_time_sec
    - rdb_current_bgsave_time_sec
    - aof_enabled
    - aof_rewrite_in_progress
    - aof_rewrite_scheduled
    - aof_last_rewrite_time_sec
    - aof_current_rewrite_time_sec
    - aof_last_bgrewrite_status
    - aof_last_write_status

    - total_connections_received
    - total_commands_processed
    - instantaneous_ops_per_sec
    - total_net_input_bytes
    - total_net_output_bytes
    - instantaneous_input_kbps
    - instantaneous_output_kbps
    - rejected_connections
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
    - migrate_cached_sockets

    - connected_slaves
    - master_repl_offset
    - repl_backlog_active
    - repl_backlog_size
    - repl_backlog_first_byte_offset
    - repl_backlog_histlen

    - used_cpu_sys
    - used_cpu_user
    - used_cpu_sys_children
    - used_cpu_user_children

    - cluster_enabled

- redis_keyspace
    - keys
    - expires
    - avg_ttl

### Tags:

- All measurements have the following tags:
    - port
    - server
    - replication_role

- The redis_keyspace measurement has an additional database tag:
    - database

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
> redis,server=localhost,port=6379,replication_role=master,host=host keyspace_hitrate=1,clients=2i,blocked_clients=0i,instantaneous_input_kbps=0,sync_full=0i,pubsub_channels=0i,pubsub_patterns=0i,total_net_output_bytes=6659253i,used_memory=842448i,total_system_memory=8351916032i,aof_current_rewrite_time_sec=-1i,rdb_changes_since_last_save=0i,sync_partial_err=0i,latest_fork_usec=508i,instantaneous_output_kbps=0,expired_keys=0i,used_memory_peak=843416i,aof_rewrite_in_progress=0i,aof_last_bgrewrite_status="ok",migrate_cached_sockets=0i,connected_slaves=0i,maxmemory_policy="noeviction",aof_rewrite_scheduled=0i,total_net_input_bytes=3125i,used_memory_rss=9564160i,repl_backlog_histlen=0i,rdb_last_bgsave_status="ok",aof_last_rewrite_time_sec=-1i,keyspace_misses=0i,client_biggest_input_buf=5i,used_cpu_user=1.33,maxmemory=0i,rdb_current_bgsave_time_sec=-1i,total_commands_processed=271i,repl_backlog_size=1048576i,used_cpu_sys=3,uptime=2822i,lru_clock=16706281i,used_memory_lua=37888i,rejected_connections=0i,sync_partial_ok=0i,evicted_keys=0i,rdb_last_save_time_elapsed=1922i,rdb_last_save_time=1493099368i,instantaneous_ops_per_sec=0i,used_cpu_user_children=0,client_longest_output_list=0i,master_repl_offset=0i,repl_backlog_active=0i,keyspace_hits=2i,used_cpu_sys_children=0,cluster_enabled=0i,rdb_last_bgsave_time_sec=0i,aof_last_write_status="ok",total_connections_received=263i,aof_enabled=0i,repl_backlog_first_byte_offset=0i,mem_fragmentation_ratio=11.35,loading=0i,rdb_bgsave_in_progress=0i 1493101290000000000
```

redis_keyspace:
```
> redis_keyspace,database=db1,host=host,server=localhost,port=6379,replication_role=master keys=1i,expires=0i,avg_ttl=0i 1493101350000000000
```
