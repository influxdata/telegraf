# Riak Input Plugin

The Riak plugin gathers metrics from one or more riak instances.

## Configuration

```toml @sample.conf
# Read metrics one or many Riak servers
[[inputs.riak]]
  # Specify a list of one or more riak http servers
  servers = ["http://localhost:8098"]
```

## Metrics

Riak provides one measurement named "riak", with the following fields:

- cpu_{avg1, avg5, avg15}
- cpu_nprocs
- clusteraae_fsm_{active, create, create_error}
- connected_nodes
- consistent_get_objsize_{100, 99, 95, mean, median}
- consistent_get_time_{100, 99, 95, mean, median}
- consistent_gets
- consistent_gets_total
- consistent_put_objsize_{100, 99, 95, mean, median}
- consistent_put_time_{100, 99, 95, mean, median}
- consistent_puts
- consistent_puts_total
- converge_delay_{last, max, min, mean}
- coord_local_{soft_loaded, unloaded}_total
- coord_redir_{least_loaded, loaded_local, unloaded}_total
- coord_redirs_total
- counter_actor_counts_{100, 99, 95, mean, median}
- dropped_vnode_requests_totals
- executing_mappers
- gossip_received
- handoff_timeouts
- hll_bytes
- hll_bytes_{100, 99, 95, mean, median, total}
- ignored_gossip_total
- index_fsm_{active, complete, create, create_error}
- index_fsm_results_{100, 99, 95, mean, median}
- index_fsm_time_{100, 99, 95, mean, median}
- late_put_fsm_coordinator_ack
- leveldb_read_block_error
- list_fsm_{active, create, create_total, create_error, create_error_total}
- map_actor_counts_{100, 99, 95, mean, median}
- mem_{allocated, total}
- memory_{atom, atom_used, binary, code, ets, processes, processes_used, system, total}
- ngrfetch_{nofetch, nofetch_total, prefetch, prefetch_total, tofetch, tofetch_total}
- ngrrepl_{empty, empty_total, error, error_total, object, object_total, srcdiscard, srcdiscard_total}
- node_get_fsm_{active, active_60s}
- node_get_fsm_counter_objsize_{100, 99, 95, mean, median}
- node_get_fsm_counter_siblings_{100, 99, 95, mean, median}
- node_get_fsm_counter_time_{100, 99, 95, mean, median}
- node_get_fsm_{errors, errors_total}
- node_get_fsm_hll_objsize_{100, 99, 95, mean, median}
- node_get_fsm_hll_siblings_{100, 99, 95, mean, median}
- node_get_fsm_hll_time_{100, 99, 95, mean, median}
- node_get_fsm_in_rate
- node_get_fsm_map_objsize_{100, 99, 95, mean, median}
- node_get_fsm_map_siblings_{100, 99, 95, mean, median}
- node_get_fsm_map_time_{100, 99, 95, mean, median}
- node_get_fsm_objsize_{100, 99, 95, mean, median}
- node_get_fsm_out_rate
- node_get_fsm_{rejected, rejected_60s, rejected_total}
- node_get_fsm_set_objsize_{100, 99, 95, mean, median}
- node_get_fsm_set_siblings_{100, 99, 95, mean, median}
- node_get_fsm_set_time_{100, 99, 95, mean, median}
- node_get_fsm_siblings_{100, 99, 95, mean, median}
- node_get_fsm_time_{100, 99, 95, mean, median}
- node_gets
- node_gets_counter
- node_gets_counter_total
- node_gets_hll
- node_gets_hll_total
- node_gets_map
- node_gets_map_total
- node_gets_set
- node_gets_set_total
- node_gets_total
- node_put_fsm_active
- node_put_fsm_active_60s
- node_put_fsm_counter_time_{100, 99, 95, mean, median}
- node_put_fsm_hll_time_{100, 99, 95, mean, median}
- node_put_fsm_in_rate
- node_put_fsm_map_time_{100, 99, 95, mean, median}
- node_put_fsm_out_rate
- node_put_fsm_rejected
- node_put_fsm_rejected_60s
- node_put_fsm_rejected_total
- node_put_fsm_set_time_{100, 99, 95, mean, median}
- node_put_fsm_time_{100, 99, 95, mean, median}
- node_puts
- node_puts_counter
- node_puts_counter_total
- node_puts_hll
- node_puts_hll_total
- node_puts_map
- node_puts_map_total
- node_puts_set
- node_puts_set_total
- node_puts_total
- nodename
- object_counter_merge
- object_counter_merge_time_{100, 99, 95, mean, median}
- object_counter_merge_total
- object_hll_merge
- object_hll_merge_time_{100, 99, 95, mean, median}
- object_hll_merge_total
- object_map_merge
- object_map_merge_time_{100, 99, 95, mean, median}
- object_map_merge_total
- object_merge
- object_merge_time_{100, 99, 95, mean, median}
- object_merge_total
- object_set_merge
- object_set_merge_time_{100, 99, 95, mean, median}
- object_set_merge_total
- pbc_{active, connects, connects_total}
- pipeline_{active, create_count, create_one, create_error_count, create_error_one}
- postcommit_fail
- precommit_fail
- read_repairs
- read_repairs_{counter, counter_total}
- read_repairs_fallback_{notfound_count, notfound_one, outofdate_count, outofdate_one}
- read_repairs_{hll, hll_total}
- read_repairs_{map, map_total}
- read_repairs_primary_{notfound_count, notfound_one, outofdate_count, outofdate_one}
- read_repairs_{set, set_total}
- read_repairs_total
- rebalance_delay_{last, max, min, mean}
- rejected_handoffs
- riak_kv_vnodeq_{max, min, mean, median, total}
- riak_kv_vnodes_running
- riak_pipe_vnodeq_{max, min, mean, median, total}
- riak_pipe_vnodes_running
- ring_creation_size
- ring_{members, num_partition, ownership}
- rings_{reconciled, reconciled_total}
- set_actor_counts_{100, 99, 95, mean, median}
- skipped_read_repairs
- skipped_read_repairs_total
- soft_loaded_vnode_mbox_total
- storage_backend
- sys_driver_version
- sys_global_heaps_size
- sys_heap_type
- sys_logical_processors
- sys_monitor_count
- sys_otp_release
- sys_port_count
- sys_process_count
- sys_smp_support
- sys_system_architecture
- sys_system_version
- sys_thread_pool_size
- sys_threads_enabled
- sys_wordsize
- tictacaae_branch_compare
- tictacaae_branch_compare_total
- tictacaae_bucket
- tictacaae_bucket_total
- tictacaae_clock_compare
- tictacaae_clock_compare_total
- tictacaae_error
- tictacaae_error_total
- tictacaae_exchange
- tictacaae_exchange_total
- tictacaae_modtime
- tictacaae_modtime_total
- tictacaae_not_supported
- tictacaae_not_supported_total
- tictacaae_queue_microsec__max
- tictacaae_queue_microsec_mean
- tictacaae_root_compare
- tictacaae_root_compare_total
- tictacaae_timeout
- tictacaae_timeout_total
- ttaaefs_allcheck_total
- ttaaefs_daycheck_total
- ttaaefs_fail_time_100
- ttaaefs_fail_total
- ttaaefs_hourcheck_total
- ttaaefs_nosync_time_100
- ttaaefs_nosync_total
- ttaaefs_rangecheck_total
- ttaaefs_snk_ahead_total
- ttaaefs_src_ahead_total
- ttaaefs_sync_time_100
- ttaaefs_sync_total
- vnode_counter_update
- vnode_counter_update_time_{100, 99, 95, mean, median}
- vnode_counter_update_total
- vnode_get_fsm_time_{100, 99, 95, mean, median}
- vnode_gets
- vnode_gets_total
- vnode_head_fsm_time_{100, 99, 95, mean, median}
- vnode_heads
- vnode_heads_total
- vnode_hll_update
- vnode_hll_update_time_{100, 99, 95, mean, median}
- vnode_hll_update_total
- vnode_index_deletes
- vnode_index_deletes_postings
- vnode_index_deletes_postings_total
- vnode_index_deletes_total
- vnode_index_reads
- vnode_index_reads_total
- vnode_index_refreshes
- vnode_index_refreshes_total
- vnode_index_writes
- vnode_index_writes_postings
- vnode_index_writes_postings_total
- vnode_index_writes_total
- vnode_map_update
- vnode_map_update_time_{100, 99, 95, mean, median}
- vnode_map_update_total
- vnode_mbox_check_timeout_total
- vnode_put_fsm_time_{100, 99, 95, mean, median}
- vnode_puts
- vnode_puts_total
- vnode_set_update
- vnode_set_update_time_{100, 99, 95, mean, median}
- vnode_set_update_total
- worker_af1_pool_queuetime_{100, mean}
- worker_af1_pool_total
- worker_af1_pool_worktime_{100, mean}
- worker_af2_pool_queuetime_{100, mean}
- worker_af2_pool_total
- worker_af2_pool_worktime_{100, mean}
- worker_af3_pool_queuetime_{100, mean}
- worker_af3_pool_total
- worker_af3_pool_worktime_{100, mean}
- worker_af4_pool_queuetime_{100, mean}
- worker_af4_pool_total
- worker_af4_pool_worktime_{100, mean}
- worker_be_pool_queuetime_{100, mean}
- worker_be_pool_total
- worker_be_pool_worktime_{100, mean}
- worker_node_worker_pool_queuetime_{100, mean}
- worker_node_worker_pool_total
- worker_node_worker_pool_worktime_{100, mean}
- worker_unregistered_queuetime_{100, mean}
- worker_unregistered_total
- worker_unregistered_worktime_{100, mean}
- worker_vnode_pool_queuetime_{100, mean}
- worker_vnode_pool_total
- worker_vnode_pool_worktime_{100, mean}
- write_once_merge
- write_once_put_objsize_{100, 99, 95, mean, median}
- write_once_put_time_{100, 99, 95, mean, median}
- write_once_puts
- write_once_puts_total

Measurements of time (such as node_get_fsm_time_mean) are measured in
nanoseconds.

### Tags

All measurements have the following tags:

- server (the host:port of the given server address, ex. `127.0.0.1:8087`)
- nodename (the internal node name received, ex. `riak@127.0.0.1`)

## Example Output

```shell
$ ./telegraf --config telegraf.conf --input-filter riak --test
> riak,nodename=riak@127.0.0.1,server=localhost:8098 cpu_avg1=31i,cpu_avg15=69i,cpu_avg5=51i,memory_code=11563738i,memory_ets=5925872i,memory_processes=30236069i,memory_system=93074971i,memory_total=123311040i,node_get_fsm_objsize_100=0i,node_get_fsm_objsize_95=0i,node_get_fsm_objsize_99=0i,node_get_fsm_objsize_mean=0i,node_get_fsm_objsize_median=0i,node_get_fsm_siblings_100=0i,node_get_fsm_siblings_95=0i,node_get_fsm_siblings_99=0i,node_get_fsm_siblings_mean=0i,node_get_fsm_siblings_median=0i,node_get_fsm_time_100=0i,node_get_fsm_time_95=0i,node_get_fsm_time_99=0i,node_get_fsm_time_mean=0i,node_get_fsm_time_median=0i,node_gets=0i,node_gets_total=19i,node_put_fsm_time_100=0i,node_put_fsm_time_95=0i,node_put_fsm_time_99=0i,node_put_fsm_time_mean=0i,node_put_fsm_time_median=0i,node_puts=0i,node_puts_total=0i,pbc_active=0i,pbc_connects=0i,pbc_connects_total=20i,vnode_gets=0i,vnode_gets_total=57i,vnode_index_reads=0i,vnode_index_reads_total=0i,vnode_index_writes=0i,vnode_index_writes_total=0i,vnode_puts=0i,vnode_puts_total=0i,read_repair=0i,read_repairs_total=0i 1455913392622482332
```
