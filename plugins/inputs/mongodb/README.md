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
    - commands (integer)
    - connections_current (integer)
    - connections_available (integer)
    - connections_total_created (integer)
    - cursor_timed_out_count (integer)
    - cursor_no_timeout_count (integer)
    - cursor_pinned_count (integer)
    - cursor_total_count (integer)
    - deletes (integer)
    - document_deleted (integer)
    - document_inserted (integer)
    - document_returned (integer)
    - document_updated (integer)
    - flushes (integer)
    - flushes_total_time_ns (integer)
    - getmores (integer)
    - inserts (integer
    - jumbo_chunks (integer)
    - member_status (string)
    - net_in_bytes_count (integer)
    - net_out_bytes_count (integer)
    - open_connections (integer)
    - percent_cache_dirty (float)
    - percent_cache_used (float)
    - queries (integer)
    - queued_reads (integer)
    - queued_writes (integer)
    - repl_commands (integer)
    - repl_deletes (integer)
    - repl_getmores (integer)
    - repl_inserts (integer)
    - repl_lag (integer)
    - repl_queries (integer)
    - repl_updates (integer)
    - repl_oplog_window_sec (integer)
    - resident_megabytes (integer)
    - state (string)
    - total_available (integer)
    - total_created (integer)
    - total_in_use (integer)
    - total_refreshing (integer)
    - ttl_deletes (integer)
    - ttl_passes (integer)
    - updates (integer)
    - vsize_megabytes (integer)
    - wtcache_app_threads_page_read_count (integer)
    - wtcache_app_threads_page_read_time (integer)
    - wtcache_app_threads_page_write_count (integer)
    - wtcache_bytes_read_into (integer)
    - wtcache_bytes_written_from (integer)
    - wtcache_pages_read_into (integer)
    - wtcache_pages_requested_from (integer)
    - wtcache_current_bytes (integer)
    - wtcache_max_bytes_configured (integer)
    - wtcache_internal_pages_evicted (integer)
    - wtcache_modified_pages_evicted (integer)
    - wtcache_unmodified_pages_evicted (integer)
    - wtcache_pages_evicted_by_app_thread (integer)
    - wtcache_pages_queued_for_eviction (integer)
    - wtcache_server_evicting_pages (integer)
    - wtcache_tracked_dirty_bytes (integer)
    - wtcache_worker_thread_evictingpages (integer)
    - commands_per_sec (integer, deprecated in 1.10; use `commands`))
    - cursor_no_timeout (integer, opened/sec, deprecated in 1.10; use `cursor_no_timeout_count`))
    - cursor_pinned (integer, opened/sec, deprecated in 1.10; use `cursor_pinned_count`))
    - cursor_timed_out (integer, opened/sec, deprecated in 1.10; use `cursor_timed_out_count`))
    - cursor_total (integer, opened/sec, deprecated in 1.10; use `cursor_total_count`))
    - deletes_per_sec (integer, deprecated in 1.10; use `deletes`))
    - flushes_per_sec (integer, deprecated in 1.10; use `flushes`))
    - getmores_per_sec (integer, deprecated in 1.10; use `getmores`))
    - inserts_per_sec (integer, deprecated in 1.10; use `inserts`))
    - net_in_bytes (integer, bytes/sec, deprecated in 1.10; use `net_out_bytes_count`))
    - net_out_bytes (integer, bytes/sec, deprecated in 1.10; use `net_out_bytes_count`))
    - queries_per_sec (integer, deprecated in 1.10; use `queries`))
    - repl_commands_per_sec (integer, deprecated in 1.10; use `repl_commands`))
    - repl_deletes_per_sec (integer, deprecated in 1.10; use `repl_deletes`)
    - repl_getmores_per_sec (integer, deprecated in 1.10; use `repl_getmores`)
    - repl_inserts_per_sec (integer, deprecated in 1.10; use `repl_inserts`))
    - repl_queries_per_sec (integer, deprecated in 1.10; use `repl_queries`))
    - repl_updates_per_sec (integer, deprecated in 1.10; use `repl_updates`))
    - ttl_deletes_per_sec (integer, deprecated in 1.10; use `ttl_deltes`))
    - ttl_passes_per_sec (integer, deprecated in 1.10; use `ttl_passes`))
    - updates_per_sec (integer, deprecated in 1.10; use `updates`))

+ mongodb_db_stats
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
mongodb,hostname=127.0.0.1:27017 active_reads=0i,active_writes=0i,commands=1335i,commands_per_sec=7i,connections_available=814i,connections_current=5i,connections_total_created=0i,cursor_no_timeout=0i,cursor_no_timeout_count=0i,cursor_pinned=0i,cursor_pinned_count=1i,cursor_timed_out=0i,cursor_timed_out_count=0i,cursor_total=0i,cursor_total_count=1i,deletes=0i,deletes_per_sec=0i,document_deleted=0i,document_inserted=0i,document_returned=13i,document_updated=0i,flushes=5i,flushes_per_sec=0i,getmores=269i,getmores_per_sec=0i,inserts=0i,inserts_per_sec=0i,jumbo_chunks=0i,member_status="PRI",net_in_bytes=986i,net_in_bytes_count=358006i,net_out_bytes=23906i,net_out_bytes_count=661507i,open_connections=5i,percent_cache_dirty=0,percent_cache_used=0,queries=18i,queries_per_sec=3i,queued_reads=0i,queued_writes=0i,repl_commands=0i,repl_commands_per_sec=0i,repl_deletes=0i,repl_deletes_per_sec=0i,repl_getmores=0i,repl_getmores_per_sec=0i,repl_inserts=0i,repl_inserts_per_sec=0i,repl_lag=0i,repl_oplog_window_sec=24355215i,repl_queries=0i,repl_queries_per_sec=0i,repl_updates=0i,repl_updates_per_sec=0i,resident_megabytes=62i,state="PRIMARY",total_available=0i,total_created=0i,total_in_use=0i,total_refreshing=0i,ttl_deletes=0i,ttl_deletes_per_sec=0i,ttl_passes=23i,ttl_passes_per_sec=0i,updates=0i,updates_per_sec=0i,vsize_megabytes=713i,wtcache_app_threads_page_read_count=13i,wtcache_app_threads_page_read_time=74i,wtcache_app_threads_page_write_count=0i,wtcache_bytes_read_into=55271i,wtcache_bytes_written_from=125402i,wtcache_current_bytes=117050i,wtcache_max_bytes_configured=1073741824i,wtcache_pages_evicted_by_app_thread=0i,wtcache_pages_queued_for_eviction=0i,wtcache_server_evicting_pages=0i,wtcache_tracked_dirty_bytes=0i,wtcache_worker_thread_evictingpages=0i 1547159491000000000
mongodb_db_stats,db_name=admin,hostname=127.0.0.1:27017 avg_obj_size=241,collections=2i,data_size=723i,index_size=49152i,indexes=3i,num_extents=0i,objects=3i,ok=1i,storage_size=53248i,type="db_stat" 1547159491000000000
mongodb_db_stats,db_name=local,hostname=127.0.0.1:27017 avg_obj_size=813.9705882352941,collections=6i,data_size=55350i,index_size=102400i,indexes=5i,num_extents=0i,objects=68i,ok=1i,storage_size=204800i,type="db_stat" 1547159491000000000
mongodb_shard_stats,hostname=127.0.0.1:27017,in_use=3i,available=3i,created=4i,refreshing=0i 1522799074000000000
```
