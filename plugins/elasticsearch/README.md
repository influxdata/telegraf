# Elasticsearch plugin

#### Plugin arguments:
- **servers** []string: list of one or more Elasticsearch servers
- **local** boolean: If false, it will read the indices stats from all nodes
- **cluster_health** boolean: If true, it will also obtain cluster level stats

#### Description

The [elasticsearch](https://www.elastic.co/) plugin queries endpoints to obtain
[node](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-nodes-stats.html)
and optionally [cluster](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-health.html) stats.

Example:

```
    [elasticsearch]

    servers = ["http://localhost:9200"]

    local = true

    cluster_health = true
```

# Measurements
#### cluster measurements (utilizes fields instead of single values):

contains `status`, `timed_out`, `number_of_nodes`, `number_of_data_nodes`,
`active_primary_shards`, `active_shards`, `relocating_shards`,
`initializing_shards`, `unassigned_shards` fields
- elasticsearch_cluster_health

contains `status`, `number_of_shards`, `number_of_replicas`,
`active_primary_shards`, `active_shards`, `relocating_shards`,
`initializing_shards`, `unassigned_shards` fields
- elasticsearch_indices

#### node measurements:

field data circuit breaker measurement names:
- elasticsearch_breakers_fielddata_estimated_size_in_bytes value=0
- elasticsearch_breakers_fielddata_overhead value=1.03
- elasticsearch_breakers_fielddata_tripped value=0
- elasticsearch_breakers_fielddata_limit_size_in_bytes value=623326003
- elasticsearch_breakers_request_estimated_size_in_bytes value=0
- elasticsearch_breakers_request_overhead value=1.0
- elasticsearch_breakers_request_tripped value=0
- elasticsearch_breakers_request_limit_size_in_bytes value=415550668
- elasticsearch_breakers_parent_overhead value=1.0
- elasticsearch_breakers_parent_tripped value=0
- elasticsearch_breakers_parent_limit_size_in_bytes value=727213670
- elasticsearch_breakers_parent_estimated_size_in_bytes value=0

File system information, data path, free disk space, read/write measurement names:
- elasticsearch_fs_timestamp value=1436460392946
- elasticsearch_fs_total_free_in_bytes value=16909316096
- elasticsearch_fs_total_available_in_bytes value=15894814720
- elasticsearch_fs_total_total_in_bytes value=19507089408

indices size, document count, indexing and deletion times, search times,
field cache size, merges and flushes measurement names:
- elasticsearch_indices_id_cache_memory_size_in_bytes value=0
- elasticsearch_indices_completion_size_in_bytes value=0
- elasticsearch_indices_suggest_total value=0
- elasticsearch_indices_suggest_time_in_millis value=0
- elasticsearch_indices_suggest_current value=0
- elasticsearch_indices_query_cache_memory_size_in_bytes value=0
- elasticsearch_indices_query_cache_evictions value=0
- elasticsearch_indices_query_cache_hit_count value=0
- elasticsearch_indices_query_cache_miss_count value=0
- elasticsearch_indices_store_size_in_bytes value=37715234
- elasticsearch_indices_store_throttle_time_in_millis value=215
- elasticsearch_indices_merges_current_docs value=0
- elasticsearch_indices_merges_current_size_in_bytes value=0
- elasticsearch_indices_merges_total value=133
- elasticsearch_indices_merges_total_time_in_millis value=21060
- elasticsearch_indices_merges_total_docs value=203672
- elasticsearch_indices_merges_total_size_in_bytes value=142900226
- elasticsearch_indices_merges_current value=0
- elasticsearch_indices_filter_cache_memory_size_in_bytes value=7384
- elasticsearch_indices_filter_cache_evictions value=0
- elasticsearch_indices_indexing_index_total value=84790
- elasticsearch_indices_indexing_index_time_in_millis value=29680
- elasticsearch_indices_indexing_index_current value=0
- elasticsearch_indices_indexing_noop_update_total value=0
- elasticsearch_indices_indexing_throttle_time_in_millis value=0
- elasticsearch_indices_indexing_delete_tota value=13879
- elasticsearch_indices_indexing_delete_time_in_millis value=1139
- elasticsearch_indices_indexing_delete_current value=0
- elasticsearch_indices_get_exists_time_in_millis value=0
- elasticsearch_indices_get_missing_total value=1
- elasticsearch_indices_get_missing_time_in_millis value=2
- elasticsearch_indices_get_current value=0
- elasticsearch_indices_get_total value=1
- elasticsearch_indices_get_time_in_millis value=2
- elasticsearch_indices_get_exists_total value=0
- elasticsearch_indices_refresh_total value=1076
- elasticsearch_indices_refresh_total_time_in_millis value=20078
- elasticsearch_indices_percolate_current value=0
- elasticsearch_indices_percolate_memory_size_in_bytes value=-1
- elasticsearch_indices_percolate_queries value=0
- elasticsearch_indices_percolate_total value=0
- elasticsearch_indices_percolate_time_in_millis value=0
- elasticsearch_indices_translog_operations value=17702
- elasticsearch_indices_translog_size_in_bytes value=17
- elasticsearch_indices_recovery_current_as_source value=0
- elasticsearch_indices_recovery_current_as_target value=0
- elasticsearch_indices_recovery_throttle_time_in_millis value=0
- elasticsearch_indices_docs_count value=29652
- elasticsearch_indices_docs_deleted value=5229
- elasticsearch_indices_flush_total_time_in_millis value=2401
- elasticsearch_indices_flush_total value=115
- elasticsearch_indices_fielddata_memory_size_in_bytes value=12996
- elasticsearch_indices_fielddata_evictions value=0
- elasticsearch_indices_search_fetch_current value=0
- elasticsearch_indices_search_open_contexts value=0
- elasticsearch_indices_search_query_total value=1452
- elasticsearch_indices_search_query_time_in_millis value=5695
- elasticsearch_indices_search_query_current value=0
- elasticsearch_indices_search_fetch_total value=414
- elasticsearch_indices_search_fetch_time_in_millis value=146
- elasticsearch_indices_warmer_current value=0
- elasticsearch_indices_warmer_total value=2319
- elasticsearch_indices_warmer_total_time_in_millis value=448
- elasticsearch_indices_segments_count value=134
- elasticsearch_indices_segments_memory_in_bytes value=1285212
- elasticsearch_indices_segments_index_writer_memory_in_bytes value=0
- elasticsearch_indices_segments_index_writer_max_memory_in_bytes value=172368955
- elasticsearch_indices_segments_version_map_memory_in_bytes value=611844
- elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes value=0

HTTP connection measurement names:
- elasticsearch_http_current_open value=3
- elasticsearch_http_total_opened value=3

JVM stats, memory pool information, garbage collection, buffer pools measurement names:
- elasticsearch_jvm_timestamp value=1436460392945
- elasticsearch_jvm_uptime_in_millis value=202245
- elasticsearch_jvm_mem_non_heap_used_in_bytes value=39634576
- elasticsearch_jvm_mem_non_heap_committed_in_bytes value=40841216
- elasticsearch_jvm_mem_pools_young_max_in_bytes value=279183360
- elasticsearch_jvm_mem_pools_young_peak_used_in_bytes value=71630848
- elasticsearch_jvm_mem_pools_young_peak_max_in_bytes value=279183360
- elasticsearch_jvm_mem_pools_young_used_in_bytes value=32685760
- elasticsearch_jvm_mem_pools_survivor_peak_used_in_bytes value=8912888
- elasticsearch_jvm_mem_pools_survivor_peak_max_in_bytes value=34865152
- elasticsearch_jvm_mem_pools_survivor_used_in_bytes value=8912880
- elasticsearch_jvm_mem_pools_survivor_max_in_bytes value=34865152
- elasticsearch_jvm_mem_pools_old_peak_max_in_bytes value=724828160
- elasticsearch_jvm_mem_pools_old_used_in_bytes value=11110928
- elasticsearch_jvm_mem_pools_old_max_in_bytes value=724828160
- elasticsearch_jvm_mem_pools_old_peak_used_in_bytes value=14354608
- elasticsearch_jvm_mem_heap_used_in_bytes value=52709568
- elasticsearch_jvm_mem_heap_used_percent value=5
- elasticsearch_jvm_mem_heap_committed_in_bytes value=259522560
- elasticsearch_jvm_mem_heap_max_in_bytes value=1038876672
- elasticsearch_jvm_threads_peak_count value=45
- elasticsearch_jvm_threads_count value=44
- elasticsearch_jvm_gc_collectors_young_collection_count value=2
- elasticsearch_jvm_gc_collectors_young_collection_time_in_millis value=98
- elasticsearch_jvm_gc_collectors_old_collection_count value=1
- elasticsearch_jvm_gc_collectors_old_collection_time_in_millis value=24
- elasticsearch_jvm_buffer_pools_direct_count value=40
- elasticsearch_jvm_buffer_pools_direct_used_in_bytes value=6304239
- elasticsearch_jvm_buffer_pools_direct_total_capacity_in_bytes value=6304239
- elasticsearch_jvm_buffer_pools_mapped_count value=0
- elasticsearch_jvm_buffer_pools_mapped_used_in_bytes value=0
- elasticsearch_jvm_buffer_pools_mapped_total_capacity_in_bytes value=0

TCP information measurement names:
- elasticsearch_network_tcp_in_errs value=0
- elasticsearch_network_tcp_passive_opens value=16
- elasticsearch_network_tcp_curr_estab value=29
- elasticsearch_network_tcp_in_segs value=113
- elasticsearch_network_tcp_out_segs value=97
- elasticsearch_network_tcp_retrans_segs value=0
- elasticsearch_network_tcp_attempt_fails value=0
- elasticsearch_network_tcp_active_opens value=13
- elasticsearch_network_tcp_estab_resets value=0
- elasticsearch_network_tcp_out_rsts value=0

Operating system stats, load average, cpu, mem, swap measurement names:
- elasticsearch_os_swap_used_in_bytes value=0
- elasticsearch_os_swap_free_in_bytes value=487997440
- elasticsearch_os_timestamp value=1436460392944
- elasticsearch_os_uptime_in_millis value=25092
- elasticsearch_os_cpu_sys value=0
- elasticsearch_os_cpu_user value=0
- elasticsearch_os_cpu_idle value=99
- elasticsearch_os_cpu_usage value=0
- elasticsearch_os_cpu_stolen value=0
- elasticsearch_os_mem_free_percent value=74
- elasticsearch_os_mem_used_percent value=25
- elasticsearch_os_mem_actual_free_in_bytes value=1565470720
- elasticsearch_os_mem_actual_used_in_bytes value=534159360
- elasticsearch_os_mem_free_in_bytes value=477761536
- elasticsearch_os_mem_used_in_bytes value=1621868544

Process statistics, memory consumption, cpu usage, open file descriptors measurement names:
- elasticsearch_process_mem_resident_in_bytes value=246382592
- elasticsearch_process_mem_share_in_bytes value=18747392
- elasticsearch_process_mem_total_virtual_in_bytes value=4747890688
- elasticsearch_process_timestamp value=1436460392945
- elasticsearch_process_open_file_descriptors value=160
- elasticsearch_process_cpu_total_in_millis value=15480
- elasticsearch_process_cpu_percent value=2
- elasticsearch_process_cpu_sys_in_millis value=1870
- elasticsearch_process_cpu_user_in_millis value=13610

Statistics about each thread pool, including current size, queue and rejected tasks measurement names:
- elasticsearch_thread_pool_merge_threads value=6
- elasticsearch_thread_pool_merge_queue value=4
- elasticsearch_thread_pool_merge_active value=5
- elasticsearch_thread_pool_merge_rejected value=2
- elasticsearch_thread_pool_merge_largest value=5
- elasticsearch_thread_pool_merge_completed value=1
- elasticsearch_thread_pool_bulk_threads value=4
- elasticsearch_thread_pool_bulk_queue value=5
- elasticsearch_thread_pool_bulk_active value=7
- elasticsearch_thread_pool_bulk_rejected value=3
- elasticsearch_thread_pool_bulk_largest value=1
- elasticsearch_thread_pool_bulk_completed value=4
- elasticsearch_thread_pool_warmer_threads value=2
- elasticsearch_thread_pool_warmer_queue value=7
- elasticsearch_thread_pool_warmer_active value=3
- elasticsearch_thread_pool_warmer_rejected value=2
- elasticsearch_thread_pool_warmer_largest value=3
- elasticsearch_thread_pool_warmer_completed value=1
- elasticsearch_thread_pool_get_largest value=2
- elasticsearch_thread_pool_get_completed value=1
- elasticsearch_thread_pool_get_threads value=1
- elasticsearch_thread_pool_get_queue value=8
- elasticsearch_thread_pool_get_active value=4
- elasticsearch_thread_pool_get_rejected value=3
- elasticsearch_thread_pool_index_threads value=6
- elasticsearch_thread_pool_index_queue value=8
- elasticsearch_thread_pool_index_active value=4
- elasticsearch_thread_pool_index_rejected value=2
- elasticsearch_thread_pool_index_largest value=3
- elasticsearch_thread_pool_index_completed value=6
- elasticsearch_thread_pool_suggest_threads value=2
- elasticsearch_thread_pool_suggest_queue value=7
- elasticsearch_thread_pool_suggest_active value=2
- elasticsearch_thread_pool_suggest_rejected value=1
- elasticsearch_thread_pool_suggest_largest value=8
- elasticsearch_thread_pool_suggest_completed value=3
- elasticsearch_thread_pool_fetch_shard_store_queue value=7
- elasticsearch_thread_pool_fetch_shard_store_active value=4
- elasticsearch_thread_pool_fetch_shard_store_rejected value=2
- elasticsearch_thread_pool_fetch_shard_store_largest value=4
- elasticsearch_thread_pool_fetch_shard_store_completed value=1
- elasticsearch_thread_pool_fetch_shard_store_threads value=1
- elasticsearch_thread_pool_management_threads value=2
- elasticsearch_thread_pool_management_queue value=3
- elasticsearch_thread_pool_management_active value=1
- elasticsearch_thread_pool_management_rejected value=6
- elasticsearch_thread_pool_management_largest value=2
- elasticsearch_thread_pool_management_completed value=22
- elasticsearch_thread_pool_percolate_queue value=23
- elasticsearch_thread_pool_percolate_active value=13
- elasticsearch_thread_pool_percolate_rejected value=235
- elasticsearch_thread_pool_percolate_largest value=23
- elasticsearch_thread_pool_percolate_completed value=33
- elasticsearch_thread_pool_percolate_threads value=123
- elasticsearch_thread_pool_listener_active value=4
- elasticsearch_thread_pool_listener_rejected value=8
- elasticsearch_thread_pool_listener_largest value=1
- elasticsearch_thread_pool_listener_completed value=1
- elasticsearch_thread_pool_listener_threads value=1
- elasticsearch_thread_pool_listener_queue value=2
- elasticsearch_thread_pool_search_rejected value=7
- elasticsearch_thread_pool_search_largest value=2
- elasticsearch_thread_pool_search_completed value=4
- elasticsearch_thread_pool_search_threads value=5
- elasticsearch_thread_pool_search_queue value=7
- elasticsearch_thread_pool_search_active value=2
- elasticsearch_thread_pool_fetch_shard_started_threads value=3
- elasticsearch_thread_pool_fetch_shard_started_queue value=1
- elasticsearch_thread_pool_fetch_shard_started_active value=5
- elasticsearch_thread_pool_fetch_shard_started_rejected value=6
- elasticsearch_thread_pool_fetch_shard_started_largest value=4
- elasticsearch_thread_pool_fetch_shard_started_completed value=54
- elasticsearch_thread_pool_refresh_rejected value=4
- elasticsearch_thread_pool_refresh_largest value=8
- elasticsearch_thread_pool_refresh_completed value=3
- elasticsearch_thread_pool_refresh_threads value=23
- elasticsearch_thread_pool_refresh_queue value=7
- elasticsearch_thread_pool_refresh_active value=3
- elasticsearch_thread_pool_optimize_threads value=3
- elasticsearch_thread_pool_optimize_queue value=4
- elasticsearch_thread_pool_optimize_active value=1
- elasticsearch_thread_pool_optimize_rejected value=2
- elasticsearch_thread_pool_optimize_largest value=7
- elasticsearch_thread_pool_optimize_completed value=3
- elasticsearch_thread_pool_snapshot_largest value=1
- elasticsearch_thread_pool_snapshot_completed value=0
- elasticsearch_thread_pool_snapshot_threads value=8
- elasticsearch_thread_pool_snapshot_queue value=5
- elasticsearch_thread_pool_snapshot_active value=6
- elasticsearch_thread_pool_snapshot_rejected value=2
- elasticsearch_thread_pool_generic_threads value=1
- elasticsearch_thread_pool_generic_queue value=4
- elasticsearch_thread_pool_generic_active value=6
- elasticsearch_thread_pool_generic_rejected value=3
- elasticsearch_thread_pool_generic_largest value=2
- elasticsearch_thread_pool_generic_completed value=27
- elasticsearch_thread_pool_flush_threads value=3
- elasticsearch_thread_pool_flush_queue value=8
- elasticsearch_thread_pool_flush_active value=0
- elasticsearch_thread_pool_flush_rejected value=1
- elasticsearch_thread_pool_flush_largest value=5
- elasticsearch_thread_pool_flush_completed value=3

Transport statistics about sent and received bytes in cluster communication measurement names:
- elasticsearch_transport_server_open value=13
- elasticsearch_transport_rx_count value=6
- elasticsearch_transport_rx_size_in_bytes value=1380
- elasticsearch_transport_tx_count value=6
- elasticsearch_transport_tx_size_in_bytes value=1380
