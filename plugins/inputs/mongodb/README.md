# MongoDB Input Plugin

All MongoDB server versions from 2.6 and higher are supported.

## Configuration

```toml
# Read metrics from one or many MongoDB servers
[[inputs.mongodb]]
  ## An array of URLs of the form:
  ##   "mongodb://" [user ":" pass "@"] host [ ":" port]
  ## For example:
  ##   mongodb://user:auth_key@10.10.3.30:27017,
  ##   mongodb://10.10.3.33:18832,
  servers = ["mongodb://127.0.0.1:27017/?connect=direct"]

  ## When true, collect cluster status.
  ## Note that the query that counts jumbo chunks triggers a COLLSCAN, which
  ## may have an impact on performance.
  # gather_cluster_status = true

  ## When true, collect per database stats
  # gather_perdb_stats = false

  ## When true, collect per collection stats
  # gather_col_stats = false

  ## When true, collect usage statistics for each collection
  ## (insert, update, queries, remove, getmore, commands etc...).
  # gather_top_stat = false

  ## List of db where collections stats are collected
  ## If empty, all db are concerned
  # col_stats_dbs = ["local"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Permissions

If your MongoDB instance has access control enabled you will need to connect
as a user with sufficient rights.

With MongoDB 3.4 and higher, the `clusterMonitor` role can be used.  In
version 3.2 you may also need these additional permissions:

```shell
> db.grantRolesToUser("user", [{role: "read", actions: "find", db: "local"}])
```

If the user is missing required privileges you may see an error in the
Telegraf logs similar to:

```shell
Error in input [mongodb]: not authorized on admin to execute command { serverStatus: 1, recordStats: 0 }
```

Some permission related errors are logged at debug level, you can check these
messages by setting `debug = true` in the agent section of the configuration or
by running Telegraf with the `--debug` argument.

### Metrics

- mongodb
  - tags:
    - hostname
    - node_type
    - rs_name
  - fields:
    - active_reads (integer)
    - active_writes (integer)
    - aggregate_command_failed (integer)
    - aggregate_command_total (integer)
    - assert_msg (integer)
    - assert_regular (integer)
    - assert_rollovers (integer)
    - assert_user (integer)
    - assert_warning (integer)
    - available_reads (integer)
    - available_writes (integer)
    - commands (integer)
    - connections_available (integer)
    - connections_current (integer)
    - connections_total_created (integer)
    - count_command_failed (integer)
    - count_command_total (integer)
    - cursor_no_timeout_count (integer)
    - cursor_pinned_count (integer)
    - cursor_timed_out_count (integer)
    - cursor_total_count (integer)
    - delete_command_failed (integer)
    - delete_command_total (integer)
    - deletes (integer)
    - distinct_command_failed (integer)
    - distinct_command_total (integer)
    - document_deleted (integer)
    - document_inserted (integer)
    - document_returned (integer)
    - document_updated (integer)
    - find_and_modify_command_failed (integer)
    - find_and_modify_command_total (integer)
    - find_command_failed (integer)
    - find_command_total (integer)
    - flushes (integer)
    - flushes_total_time_ns (integer)
    - get_more_command_failed (integer)
    - get_more_command_total (integer)
    - getmores (integer)
    - insert_command_failed (integer)
    - insert_command_total (integer)
    - inserts (integer)
    - jumbo_chunks (integer)
    - latency_commands_count (integer)
    - latency_commands (integer)
    - latency_reads_count (integer)
    - latency_reads (integer)
    - latency_writes_count (integer)
    - latency_writes (integer)
    - member_status (string)
    - net_in_bytes_count (integer)
    - net_out_bytes_count (integer)
    - open_connections (integer)
    - operation_scan_and_order (integer)
    - operation_write_conflicts (integer)
    - page_faults (integer)
    - percent_cache_dirty (float)
    - percent_cache_used (float)
    - queries (integer)
    - queued_reads (integer)
    - queued_writes (integer)
    - repl_apply_batches_num (integer)
    - repl_apply_batches_total_millis (integer)
    - repl_apply_ops (integer)
    - repl_buffer_count (integer)
    - repl_buffer_size_bytes (integer)
    - repl_commands (integer)
    - repl_deletes (integer)
    - repl_executor_pool_in_progress_count (integer)
    - repl_executor_queues_network_in_progress (integer)
    - repl_executor_queues_sleepers (integer)
    - repl_executor_unsignaled_events (integer)
    - repl_getmores (integer)
    - repl_inserts (integer)
    - repl_lag (integer)
    - repl_network_bytes (integer)
    - repl_network_getmores_num (integer)
    - repl_network_getmores_total_millis (integer)
    - repl_network_ops (integer)
    - repl_queries (integer)
    - repl_updates (integer)
    - repl_oplog_window_sec (integer)
    - repl_state (integer)
    - resident_megabytes (integer)
    - state (string)
    - storage_freelist_search_bucket_exhausted (integer)
    - storage_freelist_search_requests (integer)
    - storage_freelist_search_scanned (integer)
    - tcmalloc_central_cache_free_bytes (integer)
    - tcmalloc_current_allocated_bytes (integer)
    - tcmalloc_current_total_thread_cache_bytes (integer)
    - tcmalloc_heap_size (integer)
    - tcmalloc_max_total_thread_cache_bytes (integer)
    - tcmalloc_pageheap_commit_count (integer)
    - tcmalloc_pageheap_committed_bytes (integer)
    - tcmalloc_pageheap_decommit_count (integer)
    - tcmalloc_pageheap_free_bytes (integer)
    - tcmalloc_pageheap_reserve_count (integer)
    - tcmalloc_pageheap_scavenge_count (integer)
    - tcmalloc_pageheap_total_commit_bytes (integer)
    - tcmalloc_pageheap_total_decommit_bytes (integer)
    - tcmalloc_pageheap_total_reserve_bytes (integer)
    - tcmalloc_pageheap_unmapped_bytes (integer)
    - tcmalloc_spinlock_total_delay_ns (integer)
    - tcmalloc_thread_cache_free_bytes (integer)
    - tcmalloc_total_free_bytes (integer)
    - tcmalloc_transfer_cache_free_bytes (integer)
    - total_available (integer)
    - total_created (integer)
    - total_docs_scanned (integer)
    - total_in_use (integer)
    - total_keys_scanned (integer)
    - total_refreshing (integer)
    - total_tickets_reads (integer)
    - total_tickets_writes (integer)
    - ttl_deletes (integer)
    - ttl_passes (integer)
    - update_command_failed (integer)
    - update_command_total (integer)
    - updates (integer)
    - uptime_ns (integer)
    - version (string)
    - vsize_megabytes (integer)
    - wt_connection_files_currently_open (integer)
    - wt_data_handles_currently_active (integer)
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
    - ttl_deletes_per_sec (integer, deprecated in 1.10; use `ttl_deletes`))
    - ttl_passes_per_sec (integer, deprecated in 1.10; use `ttl_passes`))
    - updates_per_sec (integer, deprecated in 1.10; use `updates`))

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
    - fs_used_size (integer)
    - fs_total_size (integer)

- mongodb_col_stats
  - tags:
    - hostname
    - collection
    - db_name
  - fields:
    - size (integer)
    - avg_obj_size (integer)
    - storage_size (integer)
    - total_index_size (integer)
    - ok (integer)
    - count (integer)
    - type (string)

- mongodb_shard_stats
  - tags:
    - hostname
  - fields:
    - in_use (integer)
    - available (integer)
    - created (integer)
    - refreshing (integer)

- mongodb_top_stats
  - tags:
    - collection
  - fields:
    - total_time (integer)
    - total_count (integer)
    - read_lock_time (integer)
    - read_lock_count (integer)
    - write_lock_time (integer)
    - write_lock_count (integer)
    - queries_time (integer)
    - queries_count (integer)
    - get_more_time (integer)
    - get_more_count (integer)
    - insert_time (integer)
    - insert_count (integer)
    - update_time (integer)
    - update_count (integer)
    - remove_time (integer)
    - remove_count (integer)
    - commands_time (integer)
    - commands_count (integer)

### Example Output

```shell
mongodb,hostname=127.0.0.1:27017 active_reads=1i,active_writes=0i,aggregate_command_failed=0i,aggregate_command_total=0i,assert_msg=0i,assert_regular=0i,assert_rollovers=0i,assert_user=0i,assert_warning=0i,available_reads=127i,available_writes=128i,commands=65i,commands_per_sec=4i,connections_available=51199i,connections_current=1i,connections_total_created=5i,count_command_failed=0i,count_command_total=7i,cursor_no_timeout=0i,cursor_no_timeout_count=0i,cursor_pinned=0i,cursor_pinned_count=0i,cursor_timed_out=0i,cursor_timed_out_count=0i,cursor_total=0i,cursor_total_count=0i,delete_command_failed=0i,delete_command_total=1i,deletes=1i,deletes_per_sec=0i,distinct_command_failed=0i,distinct_command_total=0i,document_deleted=0i,document_inserted=0i,document_returned=0i,document_updated=0i,find_and_modify_command_failed=0i,find_and_modify_command_total=0i,find_command_failed=0i,find_command_total=1i,flushes=52i,flushes_per_sec=0i,flushes_total_time_ns=364000000i,get_more_command_failed=0i,get_more_command_total=0i,getmores=0i,getmores_per_sec=0i,insert_command_failed=0i,insert_command_total=0i,inserts=0i,inserts_per_sec=0i,jumbo_chunks=0i,latency_commands=5740i,latency_commands_count=46i,latency_reads=348i,latency_reads_count=7i,latency_writes=0i,latency_writes_count=0i,net_in_bytes=296i,net_in_bytes_count=4262i,net_out_bytes=29322i,net_out_bytes_count=242103i,open_connections=1i,operation_scan_and_order=0i,operation_write_conflicts=0i,page_faults=1i,percent_cache_dirty=0,percent_cache_used=0,queries=1i,queries_per_sec=0i,queued_reads=0i,queued_writes=0i,resident_megabytes=33i,storage_freelist_search_bucket_exhausted=0i,storage_freelist_search_requests=0i,storage_freelist_search_scanned=0i,tcmalloc_central_cache_free_bytes=0i,tcmalloc_current_allocated_bytes=0i,tcmalloc_current_total_thread_cache_bytes=0i,tcmalloc_heap_size=0i,tcmalloc_max_total_thread_cache_bytes=0i,tcmalloc_pageheap_commit_count=0i,tcmalloc_pageheap_committed_bytes=0i,tcmalloc_pageheap_decommit_count=0i,tcmalloc_pageheap_free_bytes=0i,tcmalloc_pageheap_reserve_count=0i,tcmalloc_pageheap_scavenge_count=0i,tcmalloc_pageheap_total_commit_bytes=0i,tcmalloc_pageheap_total_decommit_bytes=0i,tcmalloc_pageheap_total_reserve_bytes=0i,tcmalloc_pageheap_unmapped_bytes=0i,tcmalloc_spinlock_total_delay_ns=0i,tcmalloc_thread_cache_free_bytes=0i,tcmalloc_total_free_bytes=0i,tcmalloc_transfer_cache_free_bytes=0i,total_available=0i,total_created=0i,total_docs_scanned=0i,total_in_use=0i,total_keys_scanned=0i,total_refreshing=0i,total_tickets_reads=128i,total_tickets_writes=128i,ttl_deletes=0i,ttl_deletes_per_sec=0i,ttl_passes=51i,ttl_passes_per_sec=0i,update_command_failed=0i,update_command_total=0i,updates=0i,updates_per_sec=0i,uptime_ns=6135152000000i,version="4.0.19",vsize_megabytes=5088i,wt_connection_files_currently_open=13i,wt_data_handles_currently_active=18i,wtcache_app_threads_page_read_count=99i,wtcache_app_threads_page_read_time=44528i,wtcache_app_threads_page_write_count=19i,wtcache_bytes_read_into=3248195i,wtcache_bytes_written_from=170612i,wtcache_current_bytes=3648788i,wtcache_internal_pages_evicted=0i,wtcache_max_bytes_configured=8053063680i,wtcache_modified_pages_evicted=0i,wtcache_pages_evicted_by_app_thread=0i,wtcache_pages_queued_for_eviction=0i,wtcache_pages_read_into=234i,wtcache_pages_requested_from=18235i,wtcache_server_evicting_pages=0i,wtcache_tracked_dirty_bytes=0i,wtcache_unmodified_pages_evicted=0i,wtcache_worker_thread_evictingpages=0i 1595691605000000000
mongodb,hostname=127.0.0.1:27017,node_type=PRI,rs_name=rs0 active_reads=1i,active_writes=0i,aggregate_command_failed=0i,aggregate_command_total=0i,assert_msg=0i,assert_regular=0i,assert_rollovers=0i,assert_user=25i,assert_warning=0i,available_reads=127i,available_writes=128i,commands=345i,commands_per_sec=4i,connections_available=838853i,connections_current=7i,connections_total_created=13i,count_command_failed=0i,count_command_total=5i,cursor_no_timeout=0i,cursor_no_timeout_count=0i,cursor_pinned=0i,cursor_pinned_count=2i,cursor_timed_out=0i,cursor_timed_out_count=0i,cursor_total=0i,cursor_total_count=4i,delete_command_failed=0i,delete_command_total=0i,deletes=0i,deletes_per_sec=0i,distinct_command_failed=0i,distinct_command_total=0i,document_deleted=0i,document_inserted=2i,document_returned=56i,document_updated=0i,find_and_modify_command_failed=0i,find_and_modify_command_total=0i,find_command_failed=0i,find_command_total=23i,flushes=4i,flushes_per_sec=0i,flushes_total_time_ns=43000000i,get_more_command_failed=0i,get_more_command_total=88i,getmores=88i,getmores_per_sec=0i,insert_command_failed=0i,insert_command_total=2i,inserts=2i,inserts_per_sec=0i,jumbo_chunks=0i,latency_commands=82532i,latency_commands_count=337i,latency_reads=30633i,latency_reads_count=111i,latency_writes=0i,latency_writes_count=0i,member_status="PRI",net_in_bytes=636i,net_in_bytes_count=172300i,net_out_bytes=38849i,net_out_bytes_count=335459i,open_connections=7i,operation_scan_and_order=1i,operation_write_conflicts=0i,page_faults=1i,percent_cache_dirty=0,percent_cache_used=0,queries=23i,queries_per_sec=2i,queued_reads=0i,queued_writes=0i,repl_apply_batches_num=0i,repl_apply_batches_total_millis=0i,repl_apply_ops=0i,repl_buffer_count=0i,repl_buffer_size_bytes=0i,repl_commands=0i,repl_commands_per_sec=0i,repl_deletes=0i,repl_deletes_per_sec=0i,repl_executor_pool_in_progress_count=0i,repl_executor_queues_network_in_progress=0i,repl_executor_queues_sleepers=3i,repl_executor_unsignaled_events=0i,repl_getmores=0i,repl_getmores_per_sec=0i,repl_inserts=0i,repl_inserts_per_sec=0i,repl_lag=0i,repl_network_bytes=0i,repl_network_getmores_num=0i,repl_network_getmores_total_millis=0i,repl_network_ops=0i,repl_oplog_window_sec=140i,repl_queries=0i,repl_queries_per_sec=0i,repl_state=1i,repl_updates=0i,repl_updates_per_sec=0i,resident_megabytes=81i,state="PRIMARY",storage_freelist_search_bucket_exhausted=0i,storage_freelist_search_requests=0i,storage_freelist_search_scanned=0i,tcmalloc_central_cache_free_bytes=322128i,tcmalloc_current_allocated_bytes=143566680i,tcmalloc_current_total_thread_cache_bytes=1098968i,tcmalloc_heap_size=181317632i,tcmalloc_max_total_thread_cache_bytes=260046848i,tcmalloc_pageheap_commit_count=53i,tcmalloc_pageheap_committed_bytes=149106688i,tcmalloc_pageheap_decommit_count=1i,tcmalloc_pageheap_free_bytes=3244032i,tcmalloc_pageheap_reserve_count=51i,tcmalloc_pageheap_scavenge_count=1i,tcmalloc_pageheap_total_commit_bytes=183074816i,tcmalloc_pageheap_total_decommit_bytes=33968128i,tcmalloc_pageheap_total_reserve_bytes=181317632i,tcmalloc_pageheap_unmapped_bytes=32210944i,tcmalloc_spinlock_total_delay_ns=0i,tcmalloc_thread_cache_free_bytes=1098968i,tcmalloc_total_free_bytes=2295976i,tcmalloc_transfer_cache_free_bytes=874880i,total_available=0i,total_created=0i,total_docs_scanned=56i,total_in_use=0i,total_keys_scanned=2i,total_refreshing=0i,total_tickets_reads=128i,total_tickets_writes=128i,ttl_deletes=0i,ttl_deletes_per_sec=0i,ttl_passes=2i,ttl_passes_per_sec=0i,update_command_failed=0i,update_command_total=0i,updates=0i,updates_per_sec=0i,uptime_ns=166481000000i,version="4.0.19",vsize_megabytes=1482i,wt_connection_files_currently_open=26i,wt_data_handles_currently_active=44i,wtcache_app_threads_page_read_count=0i,wtcache_app_threads_page_read_time=0i,wtcache_app_threads_page_write_count=56i,wtcache_bytes_read_into=0i,wtcache_bytes_written_from=130403i,wtcache_current_bytes=100312i,wtcache_internal_pages_evicted=0i,wtcache_max_bytes_configured=506462208i,wtcache_modified_pages_evicted=0i,wtcache_pages_evicted_by_app_thread=0i,wtcache_pages_queued_for_eviction=0i,wtcache_pages_read_into=0i,wtcache_pages_requested_from=2085i,wtcache_server_evicting_pages=0i,wtcache_tracked_dirty_bytes=63929i,wtcache_unmodified_pages_evicted=0i,wtcache_worker_thread_evictingpages=0i 1595691605000000000
mongodb_db_stats,db_name=admin,hostname=127.0.0.1:27017 avg_obj_size=241,collections=2i,data_size=723i,index_size=49152i,indexes=3i,num_extents=0i,objects=3i,ok=1i,storage_size=53248i,type="db_stat" 1547159491000000000
mongodb_db_stats,db_name=local,hostname=127.0.0.1:27017 avg_obj_size=813.9705882352941,collections=6i,data_size=55350i,index_size=102400i,indexes=5i,num_extents=0i,objects=68i,ok=1i,storage_size=204800i,type="db_stat" 1547159491000000000
mongodb_col_stats,collection=foo,db_name=local,hostname=127.0.0.1:27017 size=375005928i,avg_obj_size=5494,type="col_stat",storage_size=249307136i,total_index_size=2138112i,ok=1i,count=68251i 1547159491000000000
mongodb_shard_stats,hostname=127.0.0.1:27017,in_use=3i,available=3i,created=4i,refreshing=0i 1522799074000000000
mongodb_top_stats,collection=foo,total_time=1471,total_count=158,read_lock_time=49614,read_lock_count=657,write_lock_time=49125456,write_lock_count=9841,queries_time=174,queries_count=495,get_more_time=498,get_more_count=46,insert_time=2651,insert_count=1265,update_time=0,update_count=0,remove_time=0,remove_count=0,commands_time=498611,commands_count=4615
```
