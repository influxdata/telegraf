# MongoDB Input Plugin

### Configuration:

```toml
[[inputs.mongodb]]
  ## An array of URLs of the form:
  ##   "mongodb://" [user ":" pass "@"] host [ ":" port]
  ## For example:
  ##   mongodb://user:auth_key@10.10.3.30:27017,
  ##   mongodb://10.10.3.33:18832,
  servers = ["mongodb://127.0.0.1:27017"]

  ## When true, collect per database stats
  # gather_perdb_stats = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

#### Permissions:

If your MongoDB instance has access control enabled you will need to connect
as a user with sufficient rights.

With MongoDB 3.4 and higher, the `clusterMonitor` role can be used.  In
version 3.2 you may also need these additional permissions:
```
> db.grantRolesToUser("user", [{role: "read", actions: "find", db: "local"}])
```

If the user is missing required privileges you may see an error in the
Telegraf logs similar to:
```
Error in input [mongodb]: not authorized on admin to execute command { serverStatus: 1, recordStats: 0 }
```

### Metrics:

- mongodb
  - tags:
    - hostname
  - fields:
    - active_reads (integer)
    - active_writes (integer)
    - commands_per_sec (integer)
    - deletes_per_sec (integer)
    - flushes_per_sec (integer)
    - getmores_per_sec (integer)
    - inserts_per_sec (integer)
    - jumbo_chunks (integer)
    - member_status (string)
    - net_in_bytes (integer)
    - net_out_bytes (integer)
    - open_connections (integer)
    - percent_cache_dirty (float)
    - percent_cache_used (float)
    - queries_per_sec (integer)
    - queued_reads (integer)
    - queued_writes (integer)
    - repl_commands_per_sec (integer)
    - repl_deletes_per_sec (integer)
    - repl_getmores_per_sec (integer)
    - repl_inserts_per_sec (integer)
    - repl_lag (integer)
    - repl_queries_per_sec (integer)
    - repl_updates_per_sec (integer)
    - repl_oplog_window_sec (integer)
    - resident_megabytes (integer)
    - state (string)
    - total_available (integer)
    - total_created (integer)
    - total_in_use (integer)
    - total_refreshing (integer)
    - ttl_deletes_per_sec (integer)
    - ttl_passes_per_sec (integer)
    - updates_per_sec (integer)
    - vsize_megabytes (integer)
    - wtcache_app_threads_page_read_count (integer)
    - wtcache_app_threads_page_read_time (integer)
    - wtcache_app_threads_page_write_count (integer)
    - wtcache_bytes_read_into (integer)
    - wtcache_bytes_written_from (integer)
    - wtcache_current_bytes (integer)
    - wtcache_max_bytes_configured (integer)
    - wtcache_pages_evicted_by_app_thread (integer)
    - wtcache_pages_queued_for_eviction (integer)
    - wtcache_server_evicting_pages (integer)
    - wtcache_tracked_dirty_bytes (integer)
    - wtcache_worker_thread_evictingpages (integer)

- mongodb_db_stats
  - tags:
    - db_name
    - hostname
  - fields:
    - avg_obj_size (float)
    - collections (integer)
    - data_size (integer)
    - index_size (integer)
    - indexes (integer)
    - num_extents (integer)
    - objects (integer)
    - ok (integer)
    - storage_size (integer)
    - type (string)

- mongodb_shard_stats
  - tags:
    - hostname
  - fields:
    - in_use (integer)
    - available (integer)
    - created (integer)
    - refreshing (integer)

### Example Output:
```
mongodb,hostname=127.0.0.1:27017 active_reads=0i,active_writes=0i,commands_per_sec=6i,deletes_per_sec=0i,flushes_per_sec=0i,getmores_per_sec=1i,inserts_per_sec=0i,jumbo_chunks=0i,member_status="PRI",net_in_bytes=851i,net_out_bytes=23904i,open_connections=6i,percent_cache_dirty=0,percent_cache_used=0,queries_per_sec=2i,queued_reads=0i,queued_writes=0i,repl_commands_per_sec=0i,repl_deletes_per_sec=0i,repl_getmores_per_sec=0i,repl_inserts_per_sec=0i,repl_lag=0i,repl_queries_per_sec=0i,repl_updates_per_sec=0i,resident_megabytes=67i,state="PRIMARY",total_available=0i,total_created=0i,total_in_use=0i,total_refreshing=0i,ttl_deletes_per_sec=0i,ttl_passes_per_sec=0i,updates_per_sec=0i,vsize_megabytes=729i,wtcache_app_threads_page_read_count=4i,wtcache_app_threads_page_read_time=18i,wtcache_app_threads_page_write_count=6i,wtcache_bytes_read_into=10075i,wtcache_bytes_written_from=115711i,wtcache_current_bytes=86038i,wtcache_max_bytes_configured=1073741824i,wtcache_pages_evicted_by_app_thread=0i,wtcache_pages_queued_for_eviction=0i,wtcache_server_evicting_pages=0i,wtcache_tracked_dirty_bytes=0i,wtcache_worker_thread_evictingpages=0i 1522798796000000000
mongodb_db_stats,db_name=local,hostname=127.0.0.1:27017 avg_obj_size=818.625,collections=5i,data_size=6549i,index_size=86016i,indexes=4i,num_extents=0i,objects=8i,ok=1i,storage_size=118784i,type="db_stat" 1522799074000000000
mongodb_shard_stats,hostname=127.0.0.1:27017,in_use=3i,available=3i,created=4i,refreshing=0i 1522799074000000000
```
