## Telegraf Plugin: Aerospike

#### Plugin arguments:
- **servers** string array: List of aerospike servers to query (def: 127.0.0.1:3000)

#### Description

The aerospike plugin queries aerospike server(s) and get node statistics. It also collects stats for
all the configured namespaces.

For what the measurements mean, please consult the [Aerospike Metrics Reference Docs](http://www.aerospike.com/docs/reference/metrics).

The metric names, to make it less complicated in querying, have replaced all `-` with `_` as Aerospike metrics come in both forms (no idea why).

# Measurements:
#### Aerospike Statistics [values]:

Meta:
- units: Integer

Measurement names:
- batch_index_queue
- batch_index_unused_buffers
- batch_queue
- batch_tree_count
- client_connections
- data_used_bytes_memory
- index_used_bytes_memory
- info_queue
- migrate_progress_recv
- migrate_progress_send
- migrate_rx_objs
- migrate_tx_objs
- objects
- ongoing_write_reqs
- partition_absent
- partition_actual
- partition_desync
- partition_object_count
- partition_ref_count
- partition_replica
- proxy_in_progress
- query_agg_avg_rec_count
- query_avg_rec_count
- query_lookup_avg_rec_count
- queue
- record_locks
- record_refs
- sindex_used_bytes_memory
- sindex_gc_garbage_cleaned
- system_free_mem_pct
- total_bytes_disk
- total_bytes_memory
- tree_count
- scans_active
- uptime
- used_bytes_disk
- used_bytes_memory
- cluster_size
- waiting_transactions

#### Aerospike Statistics [cumulative]:

Meta:
- units: Integer

Measurement names:
- batch_errors
- batch_index_complete
- batch_index_errors
- batch_index_initiate
- batch_index_timeout
- batch_initiate
- batch_timeout
- err_duplicate_proxy_request
- err_out_of_space
- err_replica_non_null_node
- err_replica_null_node
- err_rw_cant_put_unique
- err_rw_pending_limit
- err_rw_request_not_found
- err_storage_queue_full
- err_sync_copy_null_master
- err_sync_copy_null_node
- err_tsvc_requests
- err_write_fail_bin_exists
- err_write_fail_generation
- err_write_fail_generation_xdr
- err_write_fail_incompatible_type
- err_write_fail_key_exists
- err_write_fail_key_mismatch
- err_write_fail_not_found
- err_write_fail_noxdr
- err_write_fail_parameter
- err_write_fail_prole_delete
- err_write_fail_prole_generation
- err_write_fail_prole_unknown
- err_write_fail_unknown
- fabric_msgs_rcvd
- fabric_msgs_sent
- heartbeat_received_foreign
- heartbeat_received_self
- migrate_msgs_recv
- migrate_msgs_sent
- migrate_num_incoming_accepted
- migrate_num_incoming_refused
- proxy_action
- proxy_initiate
- proxy_retry
- proxy_retry_new_dest
- proxy_retry_q_full
- proxy_retry_same_dest
- proxy_unproxy
- query_abort
- query_agg
- query_agg_abort
- query_agg_err
- query_agg_success
- query_bad_records
- query_fail
- query_long_queue_full
- query_long_running
- query_lookup_abort
- query_lookup_err
- query_lookups
- query_lookup_success
- query_reqs
- query_short_queue_full
- query_short_running
- query_success
- query_tracked
- read_dup_prole
- reaped_fds
- rw_err_ack_badnode
- rw_err_ack_internal
- rw_err_ack_nomatch
- rw_err_dup_cluster_key
- rw_err_dup_internal
- rw_err_dup_send
- rw_err_write_cluster_key
- rw_err_write_internal
- rw_err_write_send
- sindex_ucgarbage_found
- sindex_gc_locktimedout
- sindex_gc_inactivity_dur
- sindex_gc_activity_dur
- sindex_gc_list_creation_time
- sindex_gc_list_deletion_time
- sindex_gc_objects_validated
- sindex_gc_garbage_found
- stat_cluster_key_err_ack_dup_trans_reenqueue
- stat_cluster_key_err_ack_rw_trans_reenqueue
- stat_cluster_key_prole_retry
- stat_cluster_key_regular_processed
- stat_cluster_key_trans_to_proxy_retry
- stat_deleted_set_object
- stat_delete_success
- stat_duplicate_operation
- stat_evicted_objects
- stat_evicted_objects_time
- stat_evicted_set_objects
- stat_expired_objects
- stat_nsup_deletes_not_shipped
- stat_proxy_errs
- stat_proxy_reqs
- stat_proxy_reqs_xdr
- stat_proxy_success
- stat_read_errs_notfound
- stat_read_errs_other
- stat_read_reqs
- stat_read_reqs_xdr
- stat_read_success
- stat_rw_timeout
- stat_slow_trans_queue_batch_pop
- stat_slow_trans_queue_pop
- stat_slow_trans_queue_push
- stat_write_errs
- stat_write_errs_notfound
- stat_write_errs_other
- stat_write_reqs
- stat_write_reqs_xdr
- stat_write_success
- stat_xdr_pipe_miss
- stat_xdr_pipe_writes
- stat_zero_bin_records
- storage_defrag_corrupt_record
- storage_defrag_wait
- transactions
- basic_scans_succeeded
- basic_scans_failed
- aggr_scans_succeeded
- aggr_scans_failed
- udf_bg_scans_succeeded
- udf_bg_scans_failed
- udf_delete_err_others
- udf_delete_reqs
- udf_delete_success
- udf_lua_errs
- udf_query_rec_reqs
- udf_read_errs_other
- udf_read_reqs
- udf_read_success
- udf_replica_writes
- udf_scan_rec_reqs
- udf_write_err_others
- udf_write_reqs
- udf_write_success
- write_master
- write_prole

#### Aerospike Statistics [percentage]:

Meta:
- units: percent (out of 100)

Measurement names:
- free_pct_disk
- free_pct_memory

# Measurements:
#### Aerospike Namespace Statistics [values]:

Meta:
- units: Integer
- tags: `namespace=<namespace>`

Measurement names:
- available_bin_names
- available_pct
- current_time
- data_used_bytes_memory
- index_used_bytes_memory
- master_objects
- max_evicted_ttl
- max_void_time
- non_expirable_objects
- objects
- prole_objects
- sindex_used_bytes_memory
- total_bytes_disk
- total_bytes_memory
- used_bytes_disk
- used_bytes_memory

#### Aerospike Namespace Statistics [cumulative]:

Meta:
- units: Integer
- tags: `namespace=<namespace>`

Measurement names:
- evicted_objects
- expired_objects
- set_deleted_objects
- set_evicted_objects

#### Aerospike Namespace Statistics [percentage]:

Meta:
- units: percent (out of 100)
- tags: `namespace=<namespace>`

Measurement names:
- free_pct_disk
- free_pct_memory
