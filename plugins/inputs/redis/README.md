# Redis Input Plugin

This plugin gathers metrics from [Redis][redis] servers.

⭐ Telegraf v0.1.1
🏷️ server
💻 all

[redis]: https://redis.io/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from one or many redis servers
[[inputs.redis]]
  ## specify servers via a url matching:
  ##  [protocol://][username:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://username:password@192.168.99.100
  ##    unix:///var/run/redis.sock
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]

  ## Optional. Specify redis commands to retrieve values
  # [[inputs.redis.commands]]
  #   # The command to run where each argument is a separate element
  #   command = ["get", "sample-key"]
  #   # The field to store the result in
  #   field = "sample-key-value"
  #   # The type of the result
  #   # Can be "string", "integer", or "float"
  #   type = "string"

  ## Specify username and password for ACL auth (Redis 6.0+). You can add this
  ## to the server URI above or specify it here. The values here take
  ## precedence.
  # username = ""
  # password = ""

  ## Optional TLS Config
  ## Check tls/config.go ClientConfig for more options
  # tls_enable = true
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
```

## Metrics

The plugin gathers the results of the [INFO](https://redis.io/commands/info)
redis command.  There are two separate measurements: _redis_ and
_redis\_keyspace_, the latter is used for gathering database related statistics.

Additionally the plugin also calculates the hit/miss ratio (keyspace\_hitrate)
and the elapsed time since the last rdb save (rdb\_last\_save\_time\_elapsed).

- redis
  - keyspace_hitrate(float, number)
  - rdb_last_save_time_elapsed(int, seconds)

    **Server**
  - uptime(int, seconds)
  - lru_clock(int, number)
  - redis_version(string)

    **Clients**
  - clients(int, number)
  - client_longest_output_list(int, number)
  - client_biggest_input_buf(int, number)
  - blocked_clients(int, number)

    **Memory**
  - used_memory(int, bytes)
  - used_memory_rss(int, bytes)
  - used_memory_peak(int, bytes)
  - total_system_memory(int, bytes)
  - used_memory_lua(int, bytes)
  - maxmemory(int, bytes)
  - maxmemory_policy(string)
  - mem_fragmentation_ratio(float, number)

    **Persistence**
  - loading(int,flag)
  - rdb_changes_since_last_save(int, number)
  - rdb_bgsave_in_progress(int, flag)
  - rdb_last_save_time(int, seconds)
  - rdb_last_bgsave_status(string)
  - rdb_last_bgsave_time_sec(int, seconds)
  - rdb_current_bgsave_time_sec(int, seconds)
  - aof_enabled(int, flag)
  - aof_rewrite_in_progress(int, flag)
  - aof_rewrite_scheduled(int, flag)
  - aof_last_rewrite_time_sec(int, seconds)
  - aof_current_rewrite_time_sec(int, seconds)
  - aof_last_bgrewrite_status(string)
  - aof_last_write_status(string)

    **Stats**
  - total_connections_received(int, number)
  - total_commands_processed(int, number)
  - instantaneous_ops_per_sec(int, number)
  - total_net_input_bytes(int, bytes)
  - total_net_output_bytes(int, bytes)
  - instantaneous_input_kbps(float, KB/sec)
  - instantaneous_output_kbps(float, KB/sec)
  - rejected_connections(int, number)
  - sync_full(int, number)
  - sync_partial_ok(int, number)
  - sync_partial_err(int, number)
  - expired_keys(int, number)
  - evicted_keys(int, number)
  - keyspace_hits(int, number)
  - keyspace_misses(int, number)
  - pubsub_channels(int, number)
  - pubsub_patterns(int, number)
  - latest_fork_usec(int, microseconds)
  - migrate_cached_sockets(int, number)

    **Replication**
  - connected_slaves(int, number)
  - master_link_down_since_seconds(int, number)
  - master_link_status(string)
  - master_repl_offset(int, number)
  - second_repl_offset(int, number)
  - repl_backlog_active(int, number)
  - repl_backlog_size(int, bytes)
  - repl_backlog_first_byte_offset(int, number)
  - repl_backlog_histlen(int, bytes)

    **CPU**
  - used_cpu_sys(float, number)
  - used_cpu_user(float, number)
  - used_cpu_sys_children(float, number)
  - used_cpu_user_children(float, number)

    **Cluster**
  - cluster_enabled(int, flag)

- redis_keyspace
  - keys(int, number)
  - expires(int, number)
  - avg_ttl(int, number)

- redis_cmdstat
    Every Redis used command could have the following fields:
  - calls(int, number)
  - failed_calls(int, number)
  - rejected_calls(int, number)
  - usec(int, mircoseconds)
  - usec_per_call(float, microseconds)

- redis_latency_percentiles_usec
  - fields:
    - p50(float, microseconds)
    - p99(float, microseconds)
    - p99.9(float, microseconds)

- redis_replication
  - tags:
    - replication_role
    - replica_ip
    - replica_port
    - state (either "online", "wait_bgsave", or "send_bulk")

  - fields:
    - lag(int, number)
    - offset(int, number)

- redis_errorstat
  - tags:
    - err
  - fields:
    - total (int, number)

All measurements have the following tags:

- port
- server
- replication_role

The `redis_keyspace` measurement has an additional `database` tag.

The `redis_cmdstat` and `redis_latency_percentiles_usec` measurements have an
additional `command` tag.

## Example Output

The following configuration

```toml
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

produces these metrics:

```text
redis,server=localhost,port=6379,replication_role=master,host=host keyspace_hitrate=1,clients=2i,blocked_clients=0i,instantaneous_input_kbps=0,sync_full=0i,pubsub_channels=0i,pubsub_patterns=0i,total_net_output_bytes=6659253i,used_memory=842448i,total_system_memory=8351916032i,aof_current_rewrite_time_sec=-1i,rdb_changes_since_last_save=0i,sync_partial_err=0i,latest_fork_usec=508i,instantaneous_output_kbps=0,expired_keys=0i,used_memory_peak=843416i,aof_rewrite_in_progress=0i,aof_last_bgrewrite_status="ok",migrate_cached_sockets=0i,connected_slaves=0i,maxmemory_policy="noeviction",aof_rewrite_scheduled=0i,total_net_input_bytes=3125i,used_memory_rss=9564160i,repl_backlog_histlen=0i,rdb_last_bgsave_status="ok",aof_last_rewrite_time_sec=-1i,keyspace_misses=0i,client_biggest_input_buf=5i,used_cpu_user=1.33,maxmemory=0i,rdb_current_bgsave_time_sec=-1i,total_commands_processed=271i,repl_backlog_size=1048576i,used_cpu_sys=3,uptime=2822i,lru_clock=16706281i,used_memory_lua=37888i,rejected_connections=0i,sync_partial_ok=0i,evicted_keys=0i,rdb_last_save_time_elapsed=1922i,rdb_last_save_time=1493099368i,instantaneous_ops_per_sec=0i,used_cpu_user_children=0,client_longest_output_list=0i,master_repl_offset=0i,repl_backlog_active=0i,keyspace_hits=2i,used_cpu_sys_children=0,cluster_enabled=0i,rdb_last_bgsave_time_sec=0i,aof_last_write_status="ok",total_connections_received=263i,aof_enabled=0i,repl_backlog_first_byte_offset=0i,mem_fragmentation_ratio=11.35,loading=0i,rdb_bgsave_in_progress=0i 1493101290000000000
redis_keyspace,database=db1,host=host,server=localhost,port=6379,replication_role=master keys=1i,expires=0i,avg_ttl=0i 1493101350000000000
redis_cmdstat,command=publish,host=host,port=6379,replication_role=master,server=localhost calls=569514i,failed_calls=0i,rejected_calls=0i,usec=9916334i,usec_per_call=17.41 1559227136000000000
redis_latency_percentiles_usec,command=zadd,host=host,port=6379,replication_role=master,server=localhost p50=9.023,p99=28.031,p99.9=43.007 1559227136000000000
redis_errorstat,err=MOVED,host=host,port=6379,replication_role=master,server=localhost total=4284 1691119309000000000
```
