# Oracle Input Plugin

The Oracle input plugin gathers metrics from an Oracle database.

The following versions have been tested and are currently supported by this plugin:
- Oracle 11g EE and XE
- Oracle 12c EE and XE

**Note:** This plugin makes use of the [goracle](https://github.com/go-goracle/goracle) SQL driver. In order for metric collection to be successful, the Oracle Instant Client must be installed on the same host that the agent is running on.

### Configuration:

```toml
[[inputs.oracle]]
    ## Username used to connect to Oracle.
    username = "telegraf"
    ## Password used to connect to Oracle.
    password = "telegraf"
    ## SID used to connect to Oracle.
    sid = "localhost/sid"
    ## Minimum number of database connections that the connection pool can contain. Defaults to 10.
    min_sessions = 10
    ## Maximum number of database connections that the connection pool can contain. Defaults to 20.
    max_sessions = 20
    ## Increment by which the connection pool capacity is expanded. Defaults to 1.
    pool_increment = 1
    ## Maximum amount of time a connection may be reused. Defaults to 0s or forever.
    max_lifetime = "0s"

    ## Collect instance state metrics from V$INSTANCE. Defaults to true.
    instance_state_metrics = true
    ## Collect system metrics from V$SYSMETRIC. Defaults to true.
    system_metrics = true
    ## Collect tablespace metrics from DBA_TABLESPACE_USAGE_METRICS. Defaults to true.
    tablespace_metrics = true
    ## Collect wait event metrics from V$EVENTMETRIC. Defaults to true.
    wait_event_metrics = true
    ## Collect wait class metrics from V$WAITCLASSMETRIC. Defaults to true.
    wait_class_metrics = true
```

### Metrics:

Metrics are generated dynamically based on the contents of [V$INSTANCE](https://docs.oracle.com/database/121/REFRN/GUID-6A0C9B51-1714-4223-B166-9D54C4E65D67.htm#REFRN30105), [V$SYSMETRIC](https://docs.oracle.com/database/121/REFRN/GUID-623748C3-F765-4149-8378-F5CDAD59909A.htm#REFRN30343), [DBA_TABLESPACE_USAGE_METRICS](https://docs.oracle.com/database/121/REFRN/GUID-FE479528-BB37-4B55-92CF-9EC19EDF4F46.htm#REFRN23496), [V$EVENTMETRIC](https://docs.oracle.com/database/121/REFRN/GUID-33AB057F-1588-42BE-A407-ADB1B880583F.htm), and [V$WAITCLASSMETRIC](https://docs.oracle.com/database/121/REFRN/GUID-A73F50B3-67F4-4F34-B332-402CC29A8011.htm#REFRN30348).

#### All metrics
**Tags:**
- database_name
- instance_name
- host_name
- version
- instance_role

#### oracle_instance_state
**Fields:**
- active_state_normal
- active_state_quiescing
- active_state_quiesced
- archiver_started
- archiver_stopped
- archiver_failed
- database_status_active
- database_status_suspended
- database_status_instance_recovery
- logins_allowed
- logins_restricted
- shutdown_pending
- status_started
- status_mounted
- status_open
- status_open_migrate

#### oracle_system
 **Fields:**
- executions_per_user_call
- io_requests_per_second
- background_time_per_sec
- redo_writes_per_txn
- rows_per_sort
- execute_without_parse_ratio
- gc_cr_block_received_per_second
- physical_write_total_io_requests_per_sec
- physical_reads_direct_per_sec
- logons_per_txn
- global_cache_blocks_lost
- database_cpu_time_ratio
- row_cache_hit_ratio
- host_cpu_usage_per_sec
- physical_writes_direct_lobs_per_txn
- hard_parse_count_per_sec
- enqueue_waits_per_sec
- executions_per_txn
- queries_parallelized_per_sec
- disk_sort_per_sec
- host_cpu_utilization_percent
- physical_write_io_requests_per_sec
- total_pga_allocated
- db_block_gets_per_txn
- px_downgraded_50_to_75percent_per_sec
- cr_blocks_created_per_txn
- physical_write_bytes_per_sec
- streams_pool_usage_percentage
- user_calls_per_txn
- total_parse_count_per_sec
- open_cursors_per_txn
- user_rollback_undorec_applied_per_sec
- enqueue_requests_per_txn
- consistent_read_gets_per_sec
- user_limit_percent
- txns_per_logon
- io_megabytes_per_second
- user_transaction_per_sec
- soft_parse_ratio
- captured_user_calls
- shared_pool_free_percent
- physical_read_io_requests_per_sec
- db_block_gets_per_user_call
- total_sorts_per_user_call
- average_synchronous_single_block_read_latency
- background_cpu_usage_per_sec
- gc_current_block_received_per_second
- database_wait_time_ratio
- full_index_scans_per_sec
- db_block_changes_per_txn
- px_downgraded_75_to_99percent_per_sec
- logical_reads_per_user_call
- total_pga_used_by_sql_workareas
- vm_in_bytes_per_sec
- enqueue_waits_per_txn
- db_block_gets_per_sec
- ddl_statements_parallelized_per_sec
- redo_generated_per_txn
- global_cache_average_cr_get_time
- branch_node_splits_per_sec
- response_time_per_txn
- physical_write_total_bytes_per_sec
- average_active_sessions
- physical_reads_direct_lobs_per_sec
- cr_blocks_created_per_sec
- long_table_scans_per_txn
- cursor_cache_hit_ratio
- user_calls_ratio
- global_cache_average_current_get_time
- logons_per_sec
- recursive_calls_per_sec
- physical_read_total_bytes_per_sec
- user_commits_percentage
- user_rollback_undo_records_applied_per_txn
- total_index_scans_per_txn
- db_block_changes_per_sec
- cr_undo_records_applied_per_sec
- physical_read_bytes_per_sec
- physical_writes_per_txn
- dbwr_checkpoints_per_sec
- row_cache_miss_ratio
- physical_reads_per_sec
- open_cursors_per_sec
- total_table_scans_per_txn
- total_index_scans_per_sec
- total_parse_count_per_txn
- px_operations_not_downgraded_per_sec
- workload_capture_and_replay_status
- vm_out_bytes_per_sec
- physical_writes_per_sec
- physical_reads_direct_lobs_per_txn
- enqueue_requests_per_sec
- leaf_node_splits_per_txn
- physical_writes_direct_per_txn
- total_table_scans_per_sec
- long_table_scans_per_sec
- hard_parse_count_per_txn
- cpu_usage_per_sec
- gc_current_block_received_per_txn
- active_parallel_sessions
- buffer_cache_hit_ratio
- recursive_calls_per_txn
- full_index_scans_per_txn
- disk_sort_per_txn
- current_os_load
- logical_reads_per_txn
- redo_writes_per_sec
- library_cache_miss_ratio
- replayed_user_calls
- cell_physical_io_interconnect_bytes
- redo_allocation_hit_ratio
- physical_writes_direct_per_sec
- px_downgraded_to_serial_per_sec
- physical_read_total_io_requests_per_sec
- background_checkpoints_per_sec
- consistent_read_changes_per_sec
- global_cache_blocks_corrupted
- library_cache_hit_ratio
- temp_space_used
- enqueue_deadlocks_per_txn
- gc_cr_block_received_per_txn
- redo_generated_per_sec
- user_rollbacks_per_sec
- cr_undo_records_applied_per_txn
- px_downgraded_1_to_25percent_per_sec
- current_logons_count
- current_open_cursors_count
- memory_sorts_ratio
- physical_reads_per_txn
- pga_cache_hit_percent
- executions_per_sec
- user_calls_per_sec
- parse_failure_count_per_txn
- db_block_changes_per_user_call
- user_commits_per_sec
- user_rollbacks_percentage
- sql_service_response_time
- database_time_per_sec
- pq_qc_session_count
- pq_slave_session_count
- run_queue_per_sec
- enqueue_timeouts_per_txn
- leaf_node_splits_per_sec
- process_limit_percent
- session_limit_percent
- total_table_scans_per_user_call
- dml_statements_parallelized_per_sec
- consistent_read_changes_per_txn
- cpu_usage_per_txn
- logical_reads_per_sec
- active_serial_sessions
- physical_writes_direct_lobs_per_sec
- branch_node_splits_per_txn
- network_traffic_volume_per_sec
- enqueue_deadlocks_per_sec
- consistent_read_gets_per_txn
- px_downgraded_25_to_50percent_per_sec
- session_count
- physical_reads_direct_per_txn
- parse_failure_count_per_sec

#### oracle_tablespace
**Tags:**
- tablespace

**Fields:**
- used_space
- tablespace_size
- used_percent

#### oracle_wait_event
**Tags:**
- class
- event

**Fields:**
- num_sess_waiting
- time_waited
- wait_count
- time_waited_fg
- wait_count_fg

#### oracle_wait_class
**Tags:**
- event

**Fields:**
- average_waiter_count
- dbtime_in_wait
- time_waited
- wait_count
- time_waited_fg
- wait_count_fg

### Example Output:
```
oracle_tablespace,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,tablespace=SYSAUX,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e used_space=80976,tablespace_size=4194302,used_percent=1.9306192067237886 1519083464000000000
oracle_tablespace,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,tablespace=SYSTEM,host=localhost,database_name=XE used_space=100112,tablespace_size=4194302,used_percent=2.386857217243775 1519083464000000000
oracle_tablespace,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,tablespace=TEMP,host=localhost used_percent=0,used_space=0,tablespace_size=4194302 1519083464000000000
oracle_tablespace,tablespace=UNDOTBS1,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE used_space=160,tablespace_size=4194302,used_percent=0.003814699084615271 1519083464000000000
oracle_tablespace,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,tablespace=USERS,host=localhost tablespace_size=4194302,used_percent=0.005340578718461379,used_space=224 1519083464000000000
oracle_wait_class,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,class=other,host=localhost average_waiter_count=0.000356035469107552,dbtime_in_wait=23.3972762389917,time_waited=2.4894,wait_count=12,time_waited_fg=0,wait_count_fg=0 1519083464000000000
oracle_wait_class,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,class=application,host=localhost,database_name=XE,instance_name=xe time_waited=0,wait_count=0,time_waited_fg=0,wait_count_fg=0,average_waiter_count=0,dbtime_in_wait=0 1519083464000000000
oracle_wait_class,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,class=configuration,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e dbtime_in_wait=0,time_waited=0,wait_count=0,time_waited_fg=0,wait_count_fg=0,average_waiter_count=0 1519083464000000000
oracle_wait_class,instance_role=PRIMARY_INSTANCE,class=concurrency,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 time_waited=0.4392,wait_count=11,time_waited_fg=0.4166,wait_count_fg=0,average_waiter_count=0.0000628146453089245,dbtime_in_wait=4.12793593804337 1519083464000000000
oracle_wait_class,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,class=commit average_waiter_count=0.000241661899313501,dbtime_in_wait=15.8810868727502,time_waited=1.6897,wait_count=2,time_waited_fg=1.6897,wait_count_fg=0 1519083464000000000
oracle_wait_class,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,class=idle,host=localhost,database_name=XE average_waiter_count=28.7055997854691,dbtime_in_wait=0,time_waited=200709.5537,wait_count=724,time_waited_fg=2.7593,wait_count_fg=0 1519083464000000000
oracle_wait_class,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,class=network dbtime_in_wait=0.553587037228493,time_waited=0.0589,wait_count=28,time_waited_fg=0.0077,wait_count_fg=0,average_waiter_count=0.00000842391304347826 1519083464000000000
oracle_wait_class,instance_role=PRIMARY_INSTANCE,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 average_waiter_count=0.00000117276887871854,dbtime_in_wait=0.0770698421947987,time_waited=0.0082,wait_count=2,time_waited_fg=0,wait_count_fg=0 1519083464000000000
oracle_wait_class,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e average_waiter_count=0.00958135011441647,dbtime_in_wait=100,time_waited=66.9928,wait_count=313,time_waited_fg=0,wait_count_fg=0 1519083464000000000
oracle_wait_class,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,class=scheduler,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e average_waiter_count=0,dbtime_in_wait=0,time_waited=0,wait_count=0,time_waited_fg=0,wait_count_fg=0 1519083464000000000
oracle_system,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,host=localhost,database_name=XE,instance_name=xe buffer_cache_hit_ratio=100,physical_read_total_bytes_per_sec=67719.9084668192,sql_service_response_time=0.0108901740020471,database_cpu_time_ratio=46.9938062163407,background_time_per_sec=0.0187648169336384,enqueue_deadlocks_per_txn=0,response_time_per_txn=5.31985,active_serial_sessions=5,captured_user_calls=0,temp_space_used=0,physical_reads_direct_per_txn=0,logical_reads_per_sec=8.78759398496241,total_table_scans_per_sec=1.31578947368421,enqueue_waits_per_txn=0,enqueue_requests_per_txn=485,pq_qc_session_count=0,average_active_sessions=0.00499985902255639,run_queue_per_sec=0,user_commits_per_sec=0.028604118993135,long_table_scans_per_txn=0,consistent_read_gets_per_txn=178,physical_write_bytes_per_sec=0,total_sorts_per_user_call=3.5625,user_calls_ratio=1.63766632548618,enqueue_waits_per_sec=0,leaf_node_splits_per_txn=0,streams_pool_usage_percentage=0,replayed_user_calls=0,physical_reads_direct_lobs_per_txn=0,full_index_scans_per_sec=0,full_index_scans_per_txn=0,executions_per_sec=1.45676691729323,physical_writes_direct_lobs_per_txn=0,soft_parse_ratio=100,network_traffic_volume_per_sec=123.58409610984,physical_read_total_io_requests_per_sec=4.13329519450801,workload_capture_and_replay_status=0,dml_statements_parallelized_per_sec=0,session_count=50,total_pga_used_by_sql_workareas=0,redo_writes_per_sec=0.0469924812030075,total_table_scans_per_txn=28,library_cache_hit_ratio=100,executions_per_txn=31,pq_slave_session_count=0,current_open_cursors_count=33,executions_per_user_call=3.875,queries_parallelized_per_sec=0,redo_writes_per_txn=1,total_index_scans_per_sec=0.471967963386728,disk_sort_per_txn=0,px_downgraded_50_to_75percent_per_sec=0,gc_cr_block_received_per_second=0,user_calls_per_txn=12,pga_cache_hit_percent=100,physical_write_total_bytes_per_sec=10830.2059496568,db_block_changes_per_user_call=0.75,long_table_scans_per_sec=0,consistent_read_changes_per_txn=0,ddl_statements_parallelized_per_sec=0,vm_out_bytes_per_sec=0,session_limit_percent=10.5932203389831,physical_write_io_requests_per_sec=0,recursive_calls_per_txn=480.5,enqueue_requests_per_sec=13.8729977116705,consistent_read_changes_per_sec=0,gc_current_block_received_per_second=0,database_wait_time_ratio=53.0061937836593,physical_reads_direct_lobs_per_sec=0,redo_generated_per_txn=2228,logons_per_sec=0.18796992481203,db_block_changes_per_sec=0.56390977443609,total_parse_count_per_txn=33,leaf_node_splits_per_sec=0,active_parallel_sessions=0,background_cpu_usage_per_sec=0.486270022883295,hard_parse_count_per_sec=0,cr_undo_records_applied_per_txn=0,px_downgraded_75_to_99percent_per_sec=0,gc_cr_block_received_per_txn=0,shared_pool_free_percent=7.27727768268991,cpu_usage_per_sec=0.0715102974828375,px_downgraded_1_to_25percent_per_sec=0,row_cache_miss_ratio=0,open_cursors_per_sec=0.986842105263158,user_calls_per_sec=0.56390977443609,total_parse_count_per_sec=0.943935926773455,enqueue_timeouts_per_sec=0,consistent_read_gets_per_sec=8.36466165413534,logical_reads_per_user_call=50.25,total_table_scans_per_user_call=2.8125,user_rollbacks_percentage=0,disk_sort_per_sec=0,user_rollback_undorec_applied_per_sec=0,redo_allocation_hit_ratio=100,user_rollbacks_per_sec=0,recursive_calls_per_sec=13.7442791762014,total_pga_allocated=86863872,user_transaction_per_sec=0,logical_reads_per_txn=187,txns_per_logon=0,db_block_gets_per_user_call=1.25,cell_physical_io_interconnect_bytes=1281536,branch_node_splits_per_txn=0,current_logons_count=36,physical_read_bytes_per_sec=0,physical_writes_per_sec=0,physical_writes_direct_per_sec=0,dbwr_checkpoints_per_sec=0,enqueue_timeouts_per_txn=0,db_block_gets_per_txn=9,io_megabytes_per_second=0.0715102974828375,host_cpu_usage_per_sec=0.75187969924812,cr_blocks_created_per_txn=0,database_time_per_sec=0.499985902255639,px_operations_not_downgraded_per_sec=0,physical_reads_per_sec=0,physical_writes_direct_per_txn=0,logons_per_txn=4,background_checkpoints_per_sec=0,host_cpu_utilization_percent=0.380771061399334,rows_per_sort=12.5263157894737,db_block_gets_per_sec=0.422932330827068,process_limit_percent=13,vm_in_bytes_per_sec=0,physical_read_io_requests_per_sec=0,open_cursors_per_txn=34.5,cursor_cache_hit_ratio=28.7878787878788,branch_node_splits_per_sec=0,global_cache_blocks_lost=0,library_cache_miss_ratio=0,px_downgraded_to_serial_per_sec=0,user_limit_percent=0.000000838190317349087,row_cache_hit_ratio=100,memory_sorts_ratio=100,physical_reads_per_txn=0,px_downgraded_25_to_50percent_per_sec=0,io_requests_per_second=4.84839816933638,total_index_scans_per_txn=16.5,parse_failure_count_per_txn=0,user_rollback_undo_records_applied_per_txn=0,global_cache_average_cr_get_time=0,cr_blocks_created_per_sec=0,physical_writes_direct_lobs_per_sec=0,redo_generated_per_sec=104.699248120301,parse_failure_count_per_sec=0,execute_without_parse_ratio=0,enqueue_deadlocks_per_sec=0,physical_reads_direct_per_sec=0,gc_current_block_received_per_txn=0,global_cache_blocks_corrupted=0,db_block_changes_per_txn=12,cr_undo_records_applied_per_sec=0,cpu_usage_per_txn=2.5,global_cache_average_current_get_time=0,average_synchronous_single_block_read_latency=0,user_commits_percentage=100,physical_writes_per_txn=0,hard_parse_count_per_txn=0,physical_write_total_io_requests_per_sec=0.715102974828375,current_os_load=0.1298828125 1519083464000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=null_event,class=other,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=pmon_timer,class=idle,host=localhost time_waited=6991.7468,wait_count=23i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=1i 1519083465000000000
oracle_wait_event,event=logout_restrictor,class=concurrency,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=vktm_logical_idle_wait,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=39i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=vktm_init_wait_for_gsga,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=iorm_scheduler_slave_idle_wait,class=idle,host=localhost,database_name=XE wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=acknowledge_over_pga_limit,class=scheduler,host=localhost,database_name=XE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=parameter_file_io,class=user_io,host=localhost time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=rdbms_ipc_message,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=11i,time_waited=90016.3807,wait_count=321i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=remote_db_operation,class=network,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=network,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=remote_db_file_read num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=remote_db_file_write,class=network,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=disk_file_operations_io,class=user_io,host=localhost num_sess_waiting=0i,time_waited=0.0082,wait_count=2i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=disk_file_io_calibration,class=user_io,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=disk_file_mirror_read,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=disk_file_mirror_media_repair_write,class=user_io,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=direct_path_sync,class=user_io,host=localhost,database_name=XE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=clonedb_bitmap_file_write,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=datapump_dump_file_io time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=dbms_file_transfer_io,class=user_io,host=localhost,database_name=XE wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,event=dg_broker_configuration_file_io,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=data_file_init_write num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_file_init_write,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_archive_io,class=system_io,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=rman_backup_recovery_io,class=system_io,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=standby_redo_io,class=system_io,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=network_file_transfer,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=pluggable_database_file_copy,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,event=file_copy,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,event=backup_mml_initialization,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_v1_open_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_v1_read_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=backup_mml_v1_write_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_v1_close_backup_piece,class=administrative,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_v1_query_backup_piece,class=administrative,host=localhost time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_v1_delete_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_create_a_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_commit_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=backup_mml_command_to_channel,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_shutdown,class=administrative time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_obtain_textual_error num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=backup_mml_query_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=backup_mml_extended_initialization,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,event=backup_mml_read_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_delete_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=backup_mml_restore_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_write_backup_piece,class=administrative,host=localhost,database_name=XE time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_proxy_initialize_backup,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_proxy_cancel,class=administrative,host=localhost wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,event=backup_mml_proxy_commit_backup_piece,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=backup_mml_proxy_session_end,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=backup_mml_datafile_proxy_backup,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_datafile_proxy_restore,class=administrative,host=localhost,database_name=XE time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_proxy_initialize_restore,class=administrative,host=localhost,database_name=XE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_proxy_start_data_movement,class=administrative num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_data_movement_done num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_proxy_prepare_to_start,class=administrative num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=backup_mml_obtain_a_direct_buffer,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,event=backup_mml_release_a_direct_buffer,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,event=backup_mml_get_base_address,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=backup_mml_query_for_direct_buffers wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=ofs_operation_completion,class=administrative num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=ofs_idle num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=io_done,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=i_o_slave_wait,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,event=rman_disk_slave_io,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=rman_tape_slave_io num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=dbwr_slave_io,class=system_io num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=lgwr_slave_io,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=archiver_slave_io time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=vkrm_idle num_sess_waiting=1i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=wait_for_unread_message_on_broadcast_channel,class=idle,host=localhost wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=wait_for_unread_message_on_multiple_broadcast_channels,class=idle time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,event=class_slave_wait,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0.02,wait_count=2i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=ping,class=idle,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=watchdog_main_loop,class=idle,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=process_in_prespawned_state,class=idle,host=localhost,database_name=XE,instance_name=xe time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=ba_performance_api time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=file_repopulation_write,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=diag_idle_wait time_waited_fg=0,wait_count_fg=0,num_sess_waiting=2i,time_waited=13933.38,wait_count=134i 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=ges_remote_message,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,event=gcs_remote_message,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=heartbeat_monitor_sleep,class=idle num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=gcr_sleep,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=sga_mman_sleep_for_component_shrink,class=idle,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=retry_contact_scn_lock_master,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=control_file_sequential_read,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0.3549,wait_count=289i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=control_file_single_write,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=control_file_parallel_write num_sess_waiting=0i,time_waited=64.3646,wait_count=23i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=control_file_backup_creation,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=shared_io_pool_memory,class=concurrency num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=shared_io_pool_io_completion,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=remote_log_force_commit,class=commit,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=remote_log_force_buffer_update,class=cluster num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=remote_log_force_buffer_read,class=cluster,host=localhost,database_name=XE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=remote_log_force_buffer_send,class=cluster,host=localhost,database_name=XE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=remote_log_force_scn_range,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=remote_log_force_session_cleanout,class=cluster wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=enq_pw_flush_prewarm_buffers,class=application wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,event=latch_cache_buffers_chains,class=concurrency,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=free_buffer_waits,class=configuration,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=local_write_wait,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=checkpoint_completed,class=configuration,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=write_complete_waits,class=configuration,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=write_complete_waits_flash_cache,class=configuration wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=buffer_read_retry,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=buffer_busy_waits,class=concurrency wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_buffer_busy_acquire,class=cluster num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_buffer_busy_release,class=cluster,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=read_by_other_session time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=multiple_dbwriter_suspend_resume_for_file_offline,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=recovery_read wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=pi_renounce_write_complete wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=db_flash_cache_single_block_physical_read,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=db_flash_cache_multiblock_physical_read,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=db_flash_cache_write,class=user_io time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=db_flash_cache_invalidate_wait,class=concurrency,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=db_flash_cache_dynamic_disabling_wait,class=administrative,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=enq_ro_contention,class=application,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=enq_ro_fast_object_reuse,class=application,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=application,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=enq_ko_fast_object_checkpoint time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=mrp_redo_arrival,class=idle,host=localhost,database_name=XE wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=rfs_sequential_i_o,class=system_io,host=localhost,database_name=XE,instance_name=xe wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,event=rfs_random_i_o,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=rfs_write,class=system_io,host=localhost time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=arch_wait_for_net_re_connect,class=network wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=arch_wait_for_netserver_start,class=network,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=lns_wait_on_lgwr,class=network,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=lgwr_wait_on_lns,class=network time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,class=network,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=arch_wait_for_netserver_init_2 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=network,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=arch_wait_for_flow_control num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=arch_wait_for_netserver_detach,class=network,host=localhost,database_name=XE,instance_name=xe time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=lns_async_archive_log,class=idle,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=lns_async_dest_activation,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=lns_async_end_of_log,class=idle,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=log_file_sequential_read,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_file_single_write,class=system_io,host=localhost,database_name=XE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_file_parallel_write,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e wait_count=1i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=1i,time_waited=2.2733 1519083465000000000
oracle_wait_event,class=configuration,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=latch_redo_writing wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=latch_redo_copy,class=configuration,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_buffer_space,class=configuration num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_file_switch_checkpoint_incomplete,class=configuration,host=localhost,database_name=XE wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_file_switch_private_strand_flush_incomplete,class=configuration,host=localhost,database_name=XE,instance_name=xe time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,event=log_file_switch_archiving_needed,class=configuration,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=switch_logfile_command,class=administrative,host=localhost,database_name=XE,instance_name=xe num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_file_switch_completion,class=configuration,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_file_sync,class=commit,host=localhost,database_name=XE num_sess_waiting=2i,time_waited=1.6897,wait_count=1i,time_waited_fg=1.6897,wait_count_fg=1 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=log_file_sync_scn_ordering,class=concurrency,host=localhost time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=simulated_log_write_delay,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=heartbeat_redo_informer,class=idle,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e num_sess_waiting=1i,time_waited=7007.0858,wait_count=70i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=lgwr_real_time_apply_sync,class=idle wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=lgwr_worker_group_idle,class=idle time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=2i 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=remote_log_force_log_switch_recovery,class=cluster,host=localhost wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=db_file_sequential_read,class=user_io wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=db_file_scattered_read,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,event=db_file_single_write,class=user_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,class=system_io,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=db_file_parallel_write num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=db_file_async_io_submit,class=system_io,host=localhost,database_name=XE wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=db_file_parallel_read,class=user_io,host=localhost,database_name=XE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=enq_mv_datafile_move,class=administrative time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_request,class=cluster,host=localhost time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_cr_request,class=cluster,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=gc_cr_disk_request,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,event=gc_cr_multi_block_request,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_multi_block_request time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,event=gc_block_recovery_request,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=gc_imc_multi_block_request,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,event=gc_imc_multi_block_quiesce,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=gc_cr_block_2_way,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_cr_block_3_way,class=cluster,host=localhost wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_cr_block_busy,class=cluster num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=gc_cr_block_congested,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,instance_role=PRIMARY_INSTANCE,event=gc_cr_failure,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0 wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_cr_block_lost,class=cluster,host=localhost,database_name=XE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,event=gc_cr_block_unknown,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_block_2_way,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,event=gc_current_block_3_way,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_block_busy time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i 1519083465000000000
oracle_wait_event,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_block_congested,class=cluster,host=localhost,database_name=XE time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0,wait_count=0i 1519083465000000000
oracle_wait_event,class=cluster,host=localhost,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_retry num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_block_lost,class=cluster,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_split,class=cluster,host=localhost,database_name=XE,instance_name=xe wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_current_block_unknown,class=cluster,host=localhost num_sess_waiting=0i,time_waited=0,wait_count=0i,time_waited_fg=0,wait_count_fg=0 1519083465000000000
oracle_wait_event,database_name=XE,instance_name=xe,db_host=38e94136e66e,version=12.1.0.2.0,instance_role=PRIMARY_INSTANCE,event=gc_cr_grant_2_way,class=cluster,host=localhost wait_count=0i,time_waited_fg=0,wait_count_fg=0,num_sess_waiting=0i,time_waited=0 1519083465000000000
...
```