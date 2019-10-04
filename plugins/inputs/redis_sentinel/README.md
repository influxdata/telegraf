# Telegraf Plugin: Redis Sentinel

A plugin for Redis Sentinel to monitor multiple Sentinel instances that are
mointoring multiple Redis servers and replicas.

### Configuration:

```
# Read Redis Sentinel's basic status information
[[inputs.redis_sentinel]]
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:26379
  ##    tcp://:password@192.168.99.100
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 26379 is used
  servers = ["tcp://localhost:26379"]

  ## specify server password
  # password = "s#cr@t%"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
```

### Measurements & Fields:

The plugin gathers the results of these commands and measurements:

* `sentinel masters` - `redis_sentinel_masters`
* `sentinel sentinels` - `redis_sentinels`
* `sentinel replicas` - `redis_replicas`
* `info all` - `redis_sentinel`

The `has-quorum` (int) field in `redis_sentinel_masters` is from calling the command `sentinels ckquorum`. `1` is for true, `0` is for false.

There are 5 remote network requests made for each server listed in the config.


### Metrics

- redis_sentinel_masters
  - tags:
    - host
    - master_name
    - port
    - source

  - fields:
    - config-epoch (int)
    - down-after-milliseconds (int)
    - failover-timeout (int)
    - flags (string)
    - has-quorum (int)
    - info-refresh (int)
    - ip (string)
    - last-ok-ping-reply (int)
    - last-ping-reply (int)
    - last-ping-sent (int)
    - link-pending-commands (int)
    - link-refcount (int)
    - name (string)
    - num-other-sentinels (int)
    - num-slaves (int)
    - parallel-syncs (int)
    - port (int)
    - quorum (int)
    - role-reported (string)
    - role-reported-time (int)
    - runid (string)

- redis_sentinels
  - tags:
    - host
    - master_name
    - port
    - sentinel_ip
    - sentinel_port
    - source

  - fields:
    - down-after-milliseconds (int)
    - flags (string)
    - ip (string)
    - last-hello-message (int)
    - last-ok-ping-reply (int)
    - last-ping-reply (int)
    - last-ping-sent (int)
    - link-pending-commands (int)
    - link-refcount (int)
    - name (string)
    - port (int)
    - runid (string)
    - voted-leader (string)
    - voted-leader-epoch (int)

- redis_replicas
  - tags:
    - host
    - master_name
    - port
    - replica_ip
    - replica_port
    - source

  - fields:
    - down-after-milliseconds (int)
    - flags (string)
    - info-refresh (int)
    - ip (string)
    - last-ok-ping-reply (int)
    - last-ping-reply (int)
    - last-ping-sent (int)
    - link-pending-commands (int)
    - link-refcount (int)
    - master-host (string)
    - master-link-down-time (int)
    - master-link-status (string)
    - master-port (int)
    - name (string)
    - port (int)
    - role-reported (string)
    - role-reported-time (int)
    - runid (string)
    - slave-priority (int)
    - slave-repl-offset (int)

- redis_sentinel
  - tags:
    - host
    - port
    - source

  - fields:
    - active_defrag_hits (int)
    - active_defrag_key_hits (int)
    - active_defrag_key_misses (int)
    - active_defrag_misses (int)
    - blocked_clients (int)
    - client_recent_max_input_buffer (int)
    - client_recent_max_output_buffer (int)
    - clients (int)
    - evicted_keys (int)
    - expired_keys (int)
    - expired_stale_perc (float)
    - expired_time_cap_reached_count (int)
    - instantaneous_input_kbps (float)
    - instantaneous_ops_per_sec (int)
    - instantaneous_output_kbps (float)
    - keyspace_hits (int)
    - keyspace_misses (int)
    - latest_fork_usec (int)
    - lru_clock (int)
    - migrate_cached_sockets (int)
    - pubsub_channels (int)
    - pubsub_patterns (int)
    - redis_version (string)
    - rejected_connections (int)
    - sentinel_masters (int)
    - sentinel_running_scripts (int)
    - sentinel_scripts_queue_length (int)
    - sentinel_simulate_failure_flags (int)
    - sentinel_tilt (int)
    - slave_expires_tracked_keys (int)
    - sync_full (int)
    - sync_partial_err (int)
    - sync_partial_ok (int)
    - total_commands_processed (int)
    - total_connections_received (int)
    - total_net_input_bytes (int)
    - total_net_output_bytes (int)
    - uptime (int, seconds)
    - used_cpu_sys (float)
    - used_cpu_sys_children (float)
    - used_cpu_user (float)
    - used_cpu_user_children (float)


### Example Output:

An example of 2 Redis Sentinel instances monitoring a single master and replica. It produces:

redis_sentinel_masters:
```
redis_sentinel_masters,host=somehostname,master_name=mymaster,port=26380,source=localhost config-epoch=0i,down-after-milliseconds=30000i,failover-timeout=180000i,flags="master",has-quorum=1i,info-refresh=110i,ip="127.0.0.1",last-ok-ping-reply=819i,last-ping-reply=819i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,name="mymaster",num-other-sentinels=1i,num-slaves=1i,parallel-syncs=1i,port=6379i,quorum=2i,role-reported="master",role-reported-time=311248i,runid="c77be03053dbb5df31dea24b833b9724162ba525" 1570207377000000000

redis_sentinel_masters,host=somehostname,master_name=mymaster,port=26379,source=localhost config-epoch=0i,down-after-milliseconds=30000i,failover-timeout=180000i,flags="master",has-quorum=1i,info-refresh=1650i,ip="127.0.0.1",last-ok-ping-reply=1003i,last-ping-reply=1003i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,name="mymaster",num-other-sentinels=1i,num-slaves=1i,parallel-syncs=1i,port=6379i,quorum=2i,role-reported="master",role-reported-time=302990i,runid="c77be03053dbb5df31dea24b833b9724162ba525" 1570207377000000000
```

redis_sentinels:
```
redis_sentinels,host=somehostname,master_name=mymaster,port=26380,sentinel_ip=127.0.0.1,sentinel_port=26379,source=localhost down-after-milliseconds=30000i,flags="sentinel",ip="127.0.0.1",last-hello-message=1337i,last-ok-ping-reply=566i,last-ping-reply=566i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,name="fd7444de58ecc00f2685cd89fc11ff96c72f0569",port=26379i,runid="fd7444de58ecc00f2685cd89fc11ff96c72f0569",voted-leader="?",voted-leader-epoch=0i 1570207377000000000

redis_sentinels,host=somehostname,master_name=mymaster,port=26379,sentinel_ip=127.0.0.1,sentinel_port=26380,source=localhost down-after-milliseconds=30000i,flags="sentinel",ip="127.0.0.1",last-hello-message=1510i,last-ok-ping-reply=1004i,last-ping-reply=1004i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,name="d06519438fe1b35692cb2ea06d57833c959f9114",port=26380i,runid="d06519438fe1b35692cb2ea06d57833c959f9114",voted-leader="?",voted-leader-epoch=0i 1570207377000000000
```

redis_replicas:
```
redis_replicas,host=somehostname,master_name=mymaster,port=26379,replica_ip=127.0.0.1,replica_port=6380,source=localhost down-after-milliseconds=30000i,flags="slave",info-refresh=1651i,ip="127.0.0.1",last-ok-ping-reply=1005i,last-ping-reply=1005i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,master-host="127.0.0.1",master-link-down-time=0i,master-link-status="ok",master-port=6379i,name="127.0.0.1:6380",port=6380i,role-reported="slave",role-reported-time=302983i,runid="6e569078c6024a3d0c293a5a965baad5ece46ecd",slave-priority=100i,slave-repl-offset=40175i 1570207377000000000

redis_replicas,host=somehostname,master_name=mymaster,port=26380,replica_ip=127.0.0.1,replica_port=6380,source=localhost down-after-milliseconds=30000i,flags="slave",info-refresh=111i,ip="127.0.0.1",last-ok-ping-reply=821i,last-ping-reply=821i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,master-host="127.0.0.1",master-link-down-time=0i,master-link-status="ok",master-port=6379i,name="127.0.0.1:6380",port=6380i,role-reported="slave",role-reported-time=311243i,runid="6e569078c6024a3d0c293a5a965baad5ece46ecd",slave-priority=100i,slave-repl-offset=40441i 1570207377000000000
```

redis_sentinel
```
redis_sentinel,host=somehostname,port=26379,source=localhost active_defrag_hits=0i,active_defrag_key_hits=0i,active_defrag_key_misses=0i,active_defrag_misses=0i,blocked_clients=0i,client_recent_max_input_buffer=2i,client_recent_max_output_buffer=0i,clients=3i,evicted_keys=0i,expired_keys=0i,expired_stale_perc=0,expired_time_cap_reached_count=0i,instantaneous_input_kbps=0.01,instantaneous_ops_per_sec=0i,instantaneous_output_kbps=0,keyspace_hits=0i,keyspace_misses=0i,latest_fork_usec=0i,lru_clock=9926289i,migrate_cached_sockets=0i,pubsub_channels=0i,pubsub_patterns=0i,redis_version="5.0.5",rejected_connections=0i,sentinel_masters=1i,sentinel_running_scripts=0i,sentinel_scripts_queue_length=0i,sentinel_simulate_failure_flags=0i,sentinel_tilt=0i,slave_expires_tracked_keys=0i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=459i,total_connections_received=6i,total_net_input_bytes=24517i,total_net_output_bytes=14864i,uptime_ns=303000000000i,used_cpu_sys=0.404,used_cpu_sys_children=0,used_cpu_user=0.436,used_cpu_user_children=0 1570207377000000000

redis_sentinel,host=somehostname,port=26380,source=localhost active_defrag_hits=0i,active_defrag_key_hits=0i,active_defrag_key_misses=0i,active_defrag_misses=0i,blocked_clients=0i,client_recent_max_input_buffer=2i,client_recent_max_output_buffer=0i,clients=2i,evicted_keys=0i,expired_keys=0i,expired_stale_perc=0,expired_time_cap_reached_count=0i,instantaneous_input_kbps=0.01,instantaneous_ops_per_sec=0i,instantaneous_output_kbps=0,keyspace_hits=0i,keyspace_misses=0i,latest_fork_usec=0i,lru_clock=9926289i,migrate_cached_sockets=0i,pubsub_channels=0i,pubsub_patterns=0i,redis_version="5.0.5",rejected_connections=0i,sentinel_masters=1i,sentinel_running_scripts=0i,sentinel_scripts_queue_length=0i,sentinel_simulate_failure_flags=0i,sentinel_tilt=0i,slave_expires_tracked_keys=0i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=442i,total_connections_received=2i,total_net_input_bytes=23861i,total_net_output_bytes=4443i,uptime_ns=312000000000i,used_cpu_sys=0.46,used_cpu_sys_children=0,used_cpu_user=0.416,used_cpu_user_children=0 1570207377000000000
```
