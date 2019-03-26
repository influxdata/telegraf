# Elasticsearch input plugin

The [elasticsearch](https://www.elastic.co/) plugin queries endpoints to obtain
[node](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-nodes-stats.html)
and optionally [cluster-health](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-health.html)
or [cluster-stats](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-stats.html) metrics.

### Configuration:

```
[[inputs.elasticsearch]]
  ## specify a list of one or more Elasticsearch servers
  servers = ["http://localhost:9200"]

  ## Timeout for HTTP requests to the elastic search server(s)
  http_timeout = "5s"

  ## When local is true (the default), the node will read only its own stats.
  ## Set local to false when you want to read the node stats from all nodes
  ## of the cluster. 
  local = true

  ## Set cluster_health to true when you want to also obtain cluster health stats
  cluster_health = false

  ## Adjust cluster_health_level when you want to also obtain detailed health stats
  ## The options are
  ##  - indices (default)
  ##  - cluster
  # cluster_health_level = "indices"

  ## Set cluster_stats to true when you want to also obtain cluster stats.
  cluster_stats = false

  ## Only gather cluster_stats from the master node. To work this require local = true
  cluster_stats_only_from_master = true

  ## node_stats is a list of sub-stats that you want to have gathered. Valid options
  ## are "indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http",
  ## "breaker". Per default, all stats are gathered.
  # node_stats = ["jvm", "http"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Status mappings

When reporting health (green/yellow/red), additional field `status_code`
is reported. Field contains mapping from status:string to status_code:int
with following rules:

* `green` - 1
* `yellow` - 2
* `red` - 3
* `unknown` - 0

### Measurements & Fields:

field data circuit breaker measurement names:
- elasticsearch_breakers
  - fielddata_estimated_size_in_bytes value=0
  - fielddata_overhead value=1.03
  - fielddata_tripped value=0
  - fielddata_limit_size_in_bytes value=623326003
  - request_estimated_size_in_bytes value=0
  - request_overhead value=1.0
  - request_tripped value=0
  - request_limit_size_in_bytes value=415550668
  - parent_overhead value=1.0
  - parent_tripped value=0
  - parent_limit_size_in_bytes value=727213670
  - parent_estimated_size_in_bytes value=0

File system information, data path, free disk space, read/write measurement names:
- elasticsearch_fs
  - timestamp value=1436460392946
  - total_free_in_bytes value=16909316096
  - total_available_in_bytes value=15894814720
  - total_total_in_bytes value=19507089408

indices size, document count, indexing and deletion times, search times,
field cache size, merges and flushes measurement names:
- elasticsearch_indices
  - id_cache_memory_size_in_bytes value=0
  - completion_size_in_bytes value=0
  - suggest_total value=0
  - suggest_time_in_millis value=0
  - suggest_current value=0
  - query_cache_memory_size_in_bytes value=0
  - query_cache_evictions value=0
  - query_cache_hit_count value=0
  - query_cache_miss_count value=0
  - store_size_in_bytes value=37715234
  - store_throttle_time_in_millis value=215
  - merges_current_docs value=0
  - merges_current_size_in_bytes value=0
  - merges_total value=133
  - merges_total_time_in_millis value=21060
  - merges_total_docs value=203672
  - merges_total_size_in_bytes value=142900226
  - merges_current value=0
  - filter_cache_memory_size_in_bytes value=7384
  - filter_cache_evictions value=0
  - indexing_index_total value=84790
  - indexing_index_time_in_millis value=29680
  - indexing_index_current value=0
  - indexing_noop_update_total value=0
  - indexing_throttle_time_in_millis value=0
  - indexing_delete_tota value=13879
  - indexing_delete_time_in_millis value=1139
  - indexing_delete_current value=0
  - get_exists_time_in_millis value=0
  - get_missing_total value=1
  - get_missing_time_in_millis value=2
  - get_current value=0
  - get_total value=1
  - get_time_in_millis value=2
  - get_exists_total value=0
  - refresh_total value=1076
  - refresh_total_time_in_millis value=20078
  - percolate_current value=0
  - percolate_memory_size_in_bytes value=-1
  - percolate_queries value=0
  - percolate_total value=0
  - percolate_time_in_millis value=0
  - translog_operations value=17702
  - translog_size_in_bytes value=17
  - recovery_current_as_source value=0
  - recovery_current_as_target value=0
  - recovery_throttle_time_in_millis value=0
  - docs_count value=29652
  - docs_deleted value=5229
  - flush_total_time_in_millis value=2401
  - flush_total value=115
  - fielddata_memory_size_in_bytes value=12996
  - fielddata_evictions value=0
  - search_fetch_current value=0
  - search_open_contexts value=0
  - search_query_total value=1452
  - search_query_time_in_millis value=5695
  - search_query_current value=0
  - search_fetch_total value=414
  - search_fetch_time_in_millis value=146
  - warmer_current value=0
  - warmer_total value=2319
  - warmer_total_time_in_millis value=448
  - segments_count value=134
  - segments_memory_in_bytes value=1285212
  - segments_index_writer_memory_in_bytes value=0
  - segments_index_writer_max_memory_in_bytes value=172368955
  - segments_version_map_memory_in_bytes value=611844
  - segments_fixed_bit_set_memory_in_bytes value=0

HTTP connection measurement names:
- elasticsearch_http
  - current_open value=3
  - total_opened value=3

JVM stats, memory pool information, garbage collection, buffer pools measurement names:
- elasticsearch_jvm
  - timestamp value=1436460392945
  - uptime_in_millis value=202245
  - mem_non_heap_used_in_bytes value=39634576
  - mem_non_heap_committed_in_bytes value=40841216
  - mem_pools_young_max_in_bytes value=279183360
  - mem_pools_young_peak_used_in_bytes value=71630848
  - mem_pools_young_peak_max_in_bytes value=279183360
  - mem_pools_young_used_in_bytes value=32685760
  - mem_pools_survivor_peak_used_in_bytes value=8912888
  - mem_pools_survivor_peak_max_in_bytes value=34865152
  - mem_pools_survivor_used_in_bytes value=8912880
  - mem_pools_survivor_max_in_bytes value=34865152
  - mem_pools_old_peak_max_in_bytes value=724828160
  - mem_pools_old_used_in_bytes value=11110928
  - mem_pools_old_max_in_bytes value=724828160
  - mem_pools_old_peak_used_in_bytes value=14354608
  - mem_heap_used_in_bytes value=52709568
  - mem_heap_used_percent value=5
  - mem_heap_committed_in_bytes value=259522560
  - mem_heap_max_in_bytes value=1038876672
  - threads_peak_count value=45
  - threads_count value=44
  - gc_collectors_young_collection_count value=2
  - gc_collectors_young_collection_time_in_millis value=98
  - gc_collectors_old_collection_count value=1
  - gc_collectors_old_collection_time_in_millis value=24
  - buffer_pools_direct_count value=40
  - buffer_pools_direct_used_in_bytes value=6304239
  - buffer_pools_direct_total_capacity_in_bytes value=6304239
  - buffer_pools_mapped_count value=0
  - buffer_pools_mapped_used_in_bytes value=0
  - buffer_pools_mapped_total_capacity_in_bytes value=0

TCP information measurement names:
- elasticsearch_network
  - tcp_in_errs value=0
  - tcp_passive_opens value=16
  - tcp_curr_estab value=29
  - tcp_in_segs value=113
  - tcp_out_segs value=97
  - tcp_retrans_segs value=0
  - tcp_attempt_fails value=0
  - tcp_active_opens value=13
  - tcp_estab_resets value=0
  - tcp_out_rsts value=0

Operating system stats, load average, cpu, mem, swap measurement names:
- elasticsearch_os
  - swap_used_in_bytes value=0
  - swap_free_in_bytes value=487997440
  - timestamp value=1436460392944
  - uptime_in_millis value=25092
  - cpu_sys value=0
  - cpu_user value=0
  - cpu_idle value=99
  - cpu_usage value=0
  - cpu_stolen value=0
  - mem_free_percent value=74
  - mem_used_percent value=25
  - mem_actual_free_in_bytes value=1565470720
  - mem_actual_used_in_bytes value=534159360
  - mem_free_in_bytes value=477761536
  - mem_used_in_bytes value=1621868544

Process statistics, memory consumption, cpu usage, open file descriptors measurement names:
- elasticsearch_process
  - mem_resident_in_bytes value=246382592
  - mem_share_in_bytes value=18747392
  - mem_total_virtual_in_bytes value=4747890688
  - timestamp value=1436460392945
  - open_file_descriptors value=160
  - cpu_total_in_millis value=15480
  - cpu_percent value=2
  - cpu_sys_in_millis value=1870
  - cpu_user_in_millis value=13610

Statistics about each thread pool, including current size, queue and rejected tasks measurement names:
- elasticsearch_thread_pool
  - merge_threads value=6
  - merge_queue value=4
  - merge_active value=5
  - merge_rejected value=2
  - merge_largest value=5
  - merge_completed value=1
  - bulk_threads value=4
  - bulk_queue value=5
  - bulk_active value=7
  - bulk_rejected value=3
  - bulk_largest value=1
  - bulk_completed value=4
  - warmer_threads value=2
  - warmer_queue value=7
  - warmer_active value=3
  - warmer_rejected value=2
  - warmer_largest value=3
  - warmer_completed value=1
  - get_largest value=2
  - get_completed value=1
  - get_threads value=1
  - get_queue value=8
  - get_active value=4
  - get_rejected value=3
  - index_threads value=6
  - index_queue value=8
  - index_active value=4
  - index_rejected value=2
  - index_largest value=3
  - index_completed value=6
  - suggest_threads value=2
  - suggest_queue value=7
  - suggest_active value=2
  - suggest_rejected value=1
  - suggest_largest value=8
  - suggest_completed value=3
  - fetch_shard_store_queue value=7
  - fetch_shard_store_active value=4
  - fetch_shard_store_rejected value=2
  - fetch_shard_store_largest value=4
  - fetch_shard_store_completed value=1
  - fetch_shard_store_threads value=1
  - management_threads value=2
  - management_queue value=3
  - management_active value=1
  - management_rejected value=6
  - management_largest value=2
  - management_completed value=22
  - percolate_queue value=23
  - percolate_active value=13
  - percolate_rejected value=235
  - percolate_largest value=23
  - percolate_completed value=33
  - percolate_threads value=123
  - listener_active value=4
  - listener_rejected value=8
  - listener_largest value=1
  - listener_completed value=1
  - listener_threads value=1
  - listener_queue value=2
  - search_rejected value=7
  - search_largest value=2
  - search_completed value=4
  - search_threads value=5
  - search_queue value=7
  - search_active value=2
  - fetch_shard_started_threads value=3
  - fetch_shard_started_queue value=1
  - fetch_shard_started_active value=5
  - fetch_shard_started_rejected value=6
  - fetch_shard_started_largest value=4
  - fetch_shard_started_completed value=54
  - refresh_rejected value=4
  - refresh_largest value=8
  - refresh_completed value=3
  - refresh_threads value=23
  - refresh_queue value=7
  - refresh_active value=3
  - optimize_threads value=3
  - optimize_queue value=4
  - optimize_active value=1
  - optimize_rejected value=2
  - optimize_largest value=7
  - optimize_completed value=3
  - snapshot_largest value=1
  - snapshot_completed value=0
  - snapshot_threads value=8
  - snapshot_queue value=5
  - snapshot_active value=6
  - snapshot_rejected value=2
  - generic_threads value=1
  - generic_queue value=4
  - generic_active value=6
  - generic_rejected value=3
  - generic_largest value=2
  - generic_completed value=27
  - flush_threads value=3
  - flush_queue value=8
  - flush_active value=0
  - flush_rejected value=1
  - flush_largest value=5
  - flush_completed value=3

Transport statistics about sent and received bytes in cluster communication measurement names:
- elasticsearch_transport
  - server_open value=13
  - rx_count value=6
  - rx_size_in_bytes value=1380
  - tx_count value=6
  - tx_size_in_bytes value=1380
