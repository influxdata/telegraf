# Redis Sentinel Input Plugin

A plugin for Redis Sentinel to monitor multiple Sentinel instances that are
monitoring multiple Redis servers and replicas.

## Configuration

```toml
# Read metrics from one or many redis-sentinel servers
[[inputs.redis_sentinel]]
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:26379
  ##    tcp://:password@192.168.99.100
  ##    unix:///var/run/redis-sentinel.sock
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 26379 is used
  # servers = ["tcp://localhost:26379"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
```

## Measurements & Fields

The plugin gathers the results of these commands and measurements:

* `sentinel masters` - `redis_sentinel_masters`
* `sentinel sentinels` - `redis_sentinels`
* `sentinel replicas` - `redis_replicas`
* `info all` - `redis_sentinel`

The `has_quorum` field in `redis_sentinel_masters` is from calling the command `sentinels ckquorum`.

There are 5 remote network requests made for each server listed in the config.

## Metrics

* redis_sentinel_masters
  * tags:
    * host
    * master
    * port
    * source

  * fields:
    * config_epoch (int)
    * down_after_milliseconds (int)
    * failover_timeout (int)
    * flags (string)
    * has_quorum (bool)
    * info_refresh (int)
    * ip (string)
    * last_ok_ping_reply (int)
    * last_ping_reply (int)
    * last_ping_sent (int)
    * link_pending_commands (int)
    * link_refcount (int)
    * num_other_sentinels (int)
    * num_slaves (int)
    * parallel_syncs (int)
    * port (int)
    * quorum (int)
    * role_reported (string)
    * role_reported_time (int)

* redis_sentinel_sentinels
  * tags:
    * host
    * master
    * port
    * sentinel_ip
    * sentinel_port
    * source

  * fields:
    * down_after_milliseconds (int)
    * flags (string)
    * last_hello_message (int)
    * last_ok_ping_reply (int)
    * last_ping_reply (int)
    * last_ping_sent (int)
    * link_pending_commands (int)
    * link_refcount (int)
    * name (string)
    * voted_leader (string)
    * voted_leader_epoch (int)

* redis_sentinel_replicas
  * tags:
    * host
    * master
    * port
    * replica_ip
    * replica_port
    * source

  * fields:
    * down_after_milliseconds (int)
    * flags (string)
    * info_refresh (int)
    * last_ok_ping_reply (int)
    * last_ping_reply (int)
    * last_ping_sent (int)
    * link_pending_commands (int)
    * link_refcount (int)
    * master_host (string)
    * master_link_down_time (int)
    * master_link_status (string)
    * master_port (int)
    * name (string)
    * role_reported (string)
    * role_reported_time (int)
    * slave_priority (int)
    * slave_repl_offset (int)

* redis_sentinel
  * tags:
    * host
    * port
    * source

  * fields:
    * active_defrag_hits (int)
    * active_defrag_key_hits (int)
    * active_defrag_key_misses (int)
    * active_defrag_misses (int)
    * blocked_clients (int)
    * client_recent_max_input_buffer (int)
    * client_recent_max_output_buffer (int)
    * clients (int)
    * evicted_keys (int)
    * expired_keys (int)
    * expired_stale_perc (float)
    * expired_time_cap_reached_count (int)
    * instantaneous_input_kbps (float)
    * instantaneous_ops_per_sec (int)
    * instantaneous_output_kbps (float)
    * keyspace_hits (int)
    * keyspace_misses (int)
    * latest_fork_usec (int)
    * lru_clock (int)
    * migrate_cached_sockets (int)
    * pubsub_channels (int)
    * pubsub_patterns (int)
    * redis_version (string)
    * rejected_connections (int)
    * sentinel_masters (int)
    * sentinel_running_scripts (int)
    * sentinel_scripts_queue_length (int)
    * sentinel_simulate_failure_flags (int)
    * sentinel_tilt (int)
    * slave_expires_tracked_keys (int)
    * sync_full (int)
    * sync_partial_err (int)
    * sync_partial_ok (int)
    * total_commands_processed (int)
    * total_connections_received (int)
    * total_net_input_bytes (int)
    * total_net_output_bytes (int)
    * uptime_ns (int, nanoseconds)
    * used_cpu_sys (float)
    * used_cpu_sys_children (float)
    * used_cpu_user (float)
    * used_cpu_user_children (float)

## Example Output

An example of 2 Redis Sentinel instances monitoring a single master and replica. It produces:

### redis_sentinel_masters

```sh
redis_sentinel_masters,host=somehostname,master=mymaster,port=26380,source=localhost config_epoch=0i,down_after_milliseconds=30000i,failover_timeout=180000i,flags="master",has_quorum=1i,info_refresh=110i,ip="127.0.0.1",last_ok_ping_reply=819i,last_ping_reply=819i,last_ping_sent=0i,link_pending_commands=0i,link_refcount=1i,num_other_sentinels=1i,num_slaves=1i,parallel_syncs=1i,port=6379i,quorum=2i,role_reported="master",role_reported_time=311248i 1570207377000000000

redis_sentinel_masters,host=somehostname,master=mymaster,port=26379,source=localhost config_epoch=0i,down_after_milliseconds=30000i,failover_timeout=180000i,flags="master",has_quorum=1i,info_refresh=1650i,ip="127.0.0.1",last_ok_ping_reply=1003i,last_ping_reply=1003i,last_ping_sent=0i,link_pending_commands=0i,link_refcount=1i,num_other_sentinels=1i,num_slaves=1i,parallel_syncs=1i,port=6379i,quorum=2i,role_reported="master",role_reported_time=302990i 1570207377000000000
```

### redis_sentinel_sentinels

```sh
redis_sentinel_sentinels,host=somehostname,master=mymaster,port=26380,sentinel_ip=127.0.0.1,sentinel_port=26379,source=localhost down_after_milliseconds=30000i,flags="sentinel",last_hello_message=1337i,last_ok_ping_reply=566i,last_ping_reply=566i,last_ping_sent=0i,link_pending_commands=0i,link_refcount=1i,name="fd7444de58ecc00f2685cd89fc11ff96c72f0569",voted_leader="?",voted_leader_epoch=0i 1570207377000000000

redis_sentinel_sentinels,host=somehostname,master=mymaster,port=26379,sentinel_ip=127.0.0.1,sentinel_port=26380,source=localhost down_after_milliseconds=30000i,flags="sentinel",last_hello_message=1510i,last_ok_ping_reply=1004i,last_ping_reply=1004i,last_ping_sent=0i,link_pending_commands=0i,link_refcount=1i,name="d06519438fe1b35692cb2ea06d57833c959f9114",voted_leader="?",voted_leader_epoch=0i 1570207377000000000
```

### redis_sentinel_replicas

```sh
redis_sentinel_replicas,host=somehostname,master=mymaster,port=26379,replica_ip=127.0.0.1,replica_port=6380,source=localhost down_after_milliseconds=30000i,flags="slave",info_refresh=1651i,last_ok_ping_reply=1005i,last_ping_reply=1005i,last_ping_sent=0i,link_pending_commands=0i,link_refcount=1i,master_host="127.0.0.1",master_link_down_time=0i,master_link_status="ok",master_port=6379i,name="127.0.0.1:6380",role_reported="slave",role_reported_time=302983i,slave_priority=100i,slave_repl_offset=40175i 1570207377000000000

redis_sentinel_replicas,host=somehostname,master=mymaster,port=26380,replica_ip=127.0.0.1,replica_port=6380,source=localhost down_after_milliseconds=30000i,flags="slave",info_refresh=111i,last_ok_ping_reply=821i,last_ping_reply=821i,last_ping_sent=0i,link_pending_commands=0i,link_refcount=1i,master_host="127.0.0.1",master_link_down_time=0i,master_link_status="ok",master_port=6379i,name="127.0.0.1:6380",role_reported="slave",role_reported_time=311243i,slave_priority=100i,slave_repl_offset=40441i 1570207377000000000
```

### redis_sentinel

```sh
redis_sentinel,host=somehostname,port=26379,source=localhost active_defrag_hits=0i,active_defrag_key_hits=0i,active_defrag_key_misses=0i,active_defrag_misses=0i,blocked_clients=0i,client_recent_max_input_buffer=2i,client_recent_max_output_buffer=0i,clients=3i,evicted_keys=0i,expired_keys=0i,expired_stale_perc=0,expired_time_cap_reached_count=0i,instantaneous_input_kbps=0.01,instantaneous_ops_per_sec=0i,instantaneous_output_kbps=0,keyspace_hits=0i,keyspace_misses=0i,latest_fork_usec=0i,lru_clock=9926289i,migrate_cached_sockets=0i,pubsub_channels=0i,pubsub_patterns=0i,redis_version="5.0.5",rejected_connections=0i,sentinel_masters=1i,sentinel_running_scripts=0i,sentinel_scripts_queue_length=0i,sentinel_simulate_failure_flags=0i,sentinel_tilt=0i,slave_expires_tracked_keys=0i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=459i,total_connections_received=6i,total_net_input_bytes=24517i,total_net_output_bytes=14864i,uptime_ns=303000000000i,used_cpu_sys=0.404,used_cpu_sys_children=0,used_cpu_user=0.436,used_cpu_user_children=0 1570207377000000000

redis_sentinel,host=somehostname,port=26380,source=localhost active_defrag_hits=0i,active_defrag_key_hits=0i,active_defrag_key_misses=0i,active_defrag_misses=0i,blocked_clients=0i,client_recent_max_input_buffer=2i,client_recent_max_output_buffer=0i,clients=2i,evicted_keys=0i,expired_keys=0i,expired_stale_perc=0,expired_time_cap_reached_count=0i,instantaneous_input_kbps=0.01,instantaneous_ops_per_sec=0i,instantaneous_output_kbps=0,keyspace_hits=0i,keyspace_misses=0i,latest_fork_usec=0i,lru_clock=9926289i,migrate_cached_sockets=0i,pubsub_channels=0i,pubsub_patterns=0i,redis_version="5.0.5",rejected_connections=0i,sentinel_masters=1i,sentinel_running_scripts=0i,sentinel_scripts_queue_length=0i,sentinel_simulate_failure_flags=0i,sentinel_tilt=0i,slave_expires_tracked_keys=0i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=442i,total_connections_received=2i,total_net_input_bytes=23861i,total_net_output_bytes=4443i,uptime_ns=312000000000i,used_cpu_sys=0.46,used_cpu_sys_children=0,used_cpu_user=0.416,used_cpu_user_children=0 1570207377000000000
```
