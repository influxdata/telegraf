# Couchbase Input Plugin

Couchbase is a distributed NoSQL database.
This plugin gets metrics for each Couchbase node, as well as detailed metrics for each bucket, for a given couchbase server.

## Configuration

```toml
# Read per-node and per-bucket metrics from Couchbase
[[inputs.couchbase]]
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    http://couchbase-0.example.com/
  ##    http://admin:secret@couchbase-0.example.com:8091/
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no protocol is specified, HTTP is used.
  ## If no port is specified, 8091 is used.
  servers = ["http://localhost:8091"]

  ## Filter bucket fields to include only here.
  # bucket_stats_included = ["quota_percent_used", "ops_per_sec", "disk_fetches", "item_count", "disk_used", "data_used", "mem_used"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification (defaults to false)
  ## If set to false, tls_cert and tls_key are required
  # insecure_skip_verify = false
```

## Measurements

### couchbase_node

Tags:

- cluster: sanitized string from `servers` configuration field e.g.: `http://user:password@couchbase-0.example.com:8091/endpoint` -> `http://couchbase-0.example.com:8091/endpoint`
- hostname: Couchbase's name for the node and port, e.g., `172.16.10.187:8091`

Fields:

- memory_free (unit: bytes, example: 23181365248.0)
- memory_total (unit: bytes, example: 64424656896.0)

### couchbase_bucket

Tags:

- cluster: whatever you called it in `servers` in the configuration, e.g.: `http://couchbase-0.example.com/`)
- bucket: the name of the couchbase bucket, e.g., `blastro-df`

Default bucket fields:

- quota_percent_used (unit: percent, example: 68.85424936294555)
- ops_per_sec (unit: count, example: 5686.789686789687)
- disk_fetches (unit: count, example: 0.0)
- item_count (unit: count, example: 943239752.0)
- disk_used (unit: bytes, example: 409178772321.0)
- data_used (unit: bytes, example: 212179309111.0)
- mem_used (unit: bytes, example: 202156957464.0)

Additional fields that can be configured with the `bucket_stats_included` option:

- couch_total_disk_size
- couch_docs_fragmentation
- couch_views_fragmentation
- hit_ratio
- ep_cache_miss_rate
- ep_resident_items_rate
- vb_avg_active_queue_age
- vb_avg_replica_queue_age
- vb_avg_pending_queue_age
- vb_avg_total_queue_age
- vb_active_resident_items_ratio
- vb_replica_resident_items_ratio
- vb_pending_resident_items_ratio
- avg_disk_update_time
- avg_disk_commit_time
- avg_bg_wait_time
- avg_active_timestamp_drift
- avg_replica_timestamp_drift
- ep_dcp_views+indexes_count
- ep_dcp_views+indexes_items_remaining
- ep_dcp_views+indexes_producer_count
- ep_dcp_views+indexes_total_backlog_size
- ep_dcp_views+indexes_items_sent
- ep_dcp_views+indexes_total_bytes
- ep_dcp_views+indexes_backoff
- bg_wait_count
- bg_wait_total
- bytes_read
- bytes_written
- cas_badval
- cas_hits
- cas_misses
- cmd_get
- cmd_lookup
- cmd_set
- couch_docs_actual_disk_size
- couch_docs_data_size
- couch_docs_disk_size
- couch_spatial_data_size
- couch_spatial_disk_size
- couch_spatial_ops
- couch_views_actual_disk_size
- couch_views_data_size
- couch_views_disk_size
- couch_views_ops
- curr_connections
- curr_items
- curr_items_tot
- decr_hits
- decr_misses
- delete_hits
- delete_misses
- disk_commit_count
- disk_commit_total
- disk_update_count
- disk_update_total
- disk_write_queue
- ep_active_ahead_exceptions
- ep_active_hlc_drift
- ep_active_hlc_drift_count
- ep_bg_fetched
- ep_clock_cas_drift_threshold_exceeded
- ep_data_read_failed
- ep_data_write_failed
- ep_dcp_2i_backoff
- ep_dcp_2i_count
- ep_dcp_2i_items_remaining
- ep_dcp_2i_items_sent
- ep_dcp_2i_producer_count
- ep_dcp_2i_total_backlog_size
- ep_dcp_2i_total_bytes
- ep_dcp_cbas_backoff
- ep_dcp_cbas_count
- ep_dcp_cbas_items_remaining
- ep_dcp_cbas_items_sent
- ep_dcp_cbas_producer_count
- ep_dcp_cbas_total_backlog_size
- ep_dcp_cbas_total_bytes
- ep_dcp_eventing_backoff
- ep_dcp_eventing_count
- ep_dcp_eventing_items_remaining
- ep_dcp_eventing_items_sent
- ep_dcp_eventing_producer_count
- ep_dcp_eventing_total_backlog_size
- ep_dcp_eventing_total_bytes
- ep_dcp_fts_backoff
- ep_dcp_fts_count
- ep_dcp_fts_items_remaining
- ep_dcp_fts_items_sent
- ep_dcp_fts_producer_count
- ep_dcp_fts_total_backlog_size
- ep_dcp_fts_total_bytes
- ep_dcp_other_backoff
- ep_dcp_other_count
- ep_dcp_other_items_remaining
- ep_dcp_other_items_sent
- ep_dcp_other_producer_count
- ep_dcp_other_total_backlog_size
- ep_dcp_other_total_bytes
- ep_dcp_replica_backoff
- ep_dcp_replica_count
- ep_dcp_replica_items_remaining
- ep_dcp_replica_items_sent
- ep_dcp_replica_producer_count
- ep_dcp_replica_total_backlog_size
- ep_dcp_replica_total_bytes
- ep_dcp_views_backoff
- ep_dcp_views_count
- ep_dcp_views_items_remaining
- ep_dcp_views_items_sent
- ep_dcp_views_producer_count
- ep_dcp_views_total_backlog_size
- ep_dcp_views_total_bytes
- ep_dcp_xdcr_backoff
- ep_dcp_xdcr_count
- ep_dcp_xdcr_items_remaining
- ep_dcp_xdcr_items_sent
- ep_dcp_xdcr_producer_count
- ep_dcp_xdcr_total_backlog_size
- ep_dcp_xdcr_total_bytes
- ep_diskqueue_drain
- ep_diskqueue_fill
- ep_diskqueue_items
- ep_flusher_todo
- ep_item_commit_failed
- ep_kv_size
- ep_max_size
- ep_mem_high_wat
- ep_mem_low_wat
- ep_meta_data_memory
- ep_num_non_resident
- ep_num_ops_del_meta
- ep_num_ops_del_ret_meta
- ep_num_ops_get_meta
- ep_num_ops_set_meta
- ep_num_ops_set_ret_meta
- ep_num_value_ejects
- ep_oom_errors
- ep_ops_create
- ep_ops_update
- ep_overhead
- ep_queue_size
- ep_replica_ahead_exceptions
- ep_replica_hlc_drift
- ep_replica_hlc_drift_count
- ep_tmp_oom_errors
- ep_vb_total
- evictions
- get_hits
- get_misses
- incr_hits
- incr_misses
- mem_used
- misses
- ops
- timestamp
- vb_active_eject
- vb_active_itm_memory
- vb_active_meta_data_memory
- vb_active_num
- vb_active_num_non_resident
- vb_active_ops_create
- vb_active_ops_update
- vb_active_queue_age
- vb_active_queue_drain
- vb_active_queue_fill
- vb_active_queue_size
- vb_active_sync_write_aborted_count
- vb_active_sync_write_accepted_count
- vb_active_sync_write_committed_count
- vb_pending_curr_items
- vb_pending_eject
- vb_pending_itm_memory
- vb_pending_meta_data_memory
- vb_pending_num
- vb_pending_num_non_resident
- vb_pending_ops_create
- vb_pending_ops_update
- vb_pending_queue_age
- vb_pending_queue_drain
- vb_pending_queue_fill
- vb_pending_queue_size
- vb_replica_curr_items
- vb_replica_eject
- vb_replica_itm_memory
- vb_replica_meta_data_memory
- vb_replica_num
- vb_replica_num_non_resident
- vb_replica_ops_create
- vb_replica_ops_update
- vb_replica_queue_age
- vb_replica_queue_drain
- vb_replica_queue_fill
- vb_replica_queue_size
- vb_total_queue_age
- xdc_ops
- allocstall
- cpu_cores_available
- cpu_irq_rate
- cpu_stolen_rate
- cpu_sys_rate
- cpu_user_rate
- cpu_utilization_rate
- hibernated_requests
- hibernated_waked
- mem_actual_free
- mem_actual_used
- mem_free
- mem_limit
- mem_total
- mem_used_sys
- odp_report_failed
- rest_requests
- swap_total
- swap_used

## Example output

```shell
couchbase_node,cluster=http://localhost:8091/,hostname=172.17.0.2:8091 memory_free=7705575424,memory_total=16558182400 1547829754000000000
couchbase_bucket,bucket=beer-sample,cluster=http://localhost:8091/ quota_percent_used=27.09285736083984,ops_per_sec=0,disk_fetches=0,item_count=7303,disk_used=21662946,data_used=9325087,mem_used=28408920 1547829754000000000
```
