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

The `has-quorum` (bool) field in `redis_sentinel_masters` is from calling the command `sentinels ckquorum`.

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
    - has-quorum (bool)
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

It produces:

redis_sentinel_masters:
```
redis_sentinel_masters,host=host,master_name=my-other-master,port=26380,source=localhost config-epoch=0i,down-after-milliseconds=30000i,failover-timeout=180000i,flags="master,disconnected",has-quorum=false,info-refresh=1559589511843i,ip="127.0.0.1",last-ok-ping-reply=3375i,last-ping-reply=3375i,last-ping-sent=3375i,link-pending-commands=3i,link-refcount=1i,name="my-other-master",num-other-sentinels=0i,num-slaves=0i,parallel-syncs=1i,port=6380i,quorum=2i,role-reported="master",role-reported-time=3375i,runid="" 1559589512000000000
redis_sentinel_masters,host=host,master_name=my-master,port=26379,source=localhost config-epoch=0i,down-after-milliseconds=30000i,failover-timeout=180000i,flags="master",has-quorum=true,info-refresh=4050i,ip="127.0.0.1",last-ok-ping-reply=679i,last-ping-reply=679i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,name="my-master",num-other-sentinels=0i,num-slaves=0i,parallel-syncs=1i,port=6379i,quorum=1i,role-reported="master",role-reported-time=184657i,runid="b56d5bc6e8d7fd29333a685c390407078778458b" 1559589133000000000
```

redis_sentinels:
```
redis_sentinels,host=host,master_name=my-other-master,port=26381,sentinel_ip=127.0.0.1,sentinel_port=26380,source=localhost down-after-milliseconds=30000i,flags="sentinel",ip="127.0.0.1",last-hello-message=467i,last-ok-ping-reply=414i,last-ping-reply=414i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,name="36a31f2676b638555e5e411c7aa0f2ea6c2caf65",port=26380i,runid="36a31f2676b638555e5e411c7aa0f2ea6c2caf65",voted-leader="?",voted-leader-epoch=0i 1559591099000000000
```

redis_replicas:
```
redis_replicas,host=host,master_name=my-master,port=26379,replica_ip=127.0.0.1,replica_port=6382,source=localhost down-after-milliseconds=30000i,flags="slave",info-refresh=2227i,ip="127.0.0.1",last-ok-ping-reply=181i,last-ping-reply=181i,last-ping-sent=0i,link-pending-commands=0i,link-refcount=1i,master-host="127.0.0.1",master-link-down-time=0i,master-link-status="ok",master-port=6379i,name="127.0.0.1:6382",port=6382i,role-reported="slave",role-reported-time=2293i,runid="69e077a76096bf9c24ac0dc310b377c3ce77042a",slave-priority=100i,slave-repl-offset=0i 1559589814000000000
```

redis_sentinel
```
redis_sentinel,host=host,master_name=my-master,port=26379,replica_ip=127.0.0.1,replica_port=6382,source=localhost active_defrag_hits=0i,active_defrag_key_hits=0i,active_defrag_key_misses=0i,active_defrag_misses=0i,blocked_clients=0i,client_recent_max_input_buffer=2i,client_recent_max_output_buffer=0i,clients=1i,evicted_keys=0i,expired_keys=0i,expired_stale_perc=0,expired_time_cap_reached_count=0i,instantaneous_input_kbps=0,instantaneous_ops_per_sec=0i,instantaneous_output_kbps=0,keyspace_hits=0i,keyspace_misses=0i,latest_fork_usec=0i,lru_clock=16092078i,master0="name=my-master,status=ok,address=127.0.0.1:6379,slaves=2,sentinels=1",migrate_cached_sockets=0i,pubsub_channels=0i,pubsub_patterns=0i,redis_version="5.0.3",rejected_connections=0i,sentinel_masters=1i,sentinel_running_scripts=0i,sentinel_scripts_queue_length=0i,sentinel_simulate_failure_flags=0i,sentinel_tilt=0i,slave_expires_tracked_keys=0i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=9i,total_connections_received=3i,total_net_input_bytes=392i,total_net_output_bytes=5559i,uptime=641i,used_cpu_sys=0.411944,used_cpu_sys_children=0,used_cpu_user=0.403248,used_cpu_user_children=0 1559595950000000000
```
