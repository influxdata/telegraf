# Ceph Storage Input Plugin

Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster.

Ceph has introduced a Telegraf and Influx plugin in the 13.x Mimic release. The Telegraf module sends to a Telegraf configured with a socket_listener. [Learn more in their docs](http://docs.ceph.com/docs/mimic/mgr/telegraf/)

*Admin Socket Stats*

This gatherer works by scanning the configured SocketDir for OSD and MON socket files.  When it finds
a MON socket, it runs **ceph --admin-daemon $file perfcounters_dump**. For OSDs it runs **ceph --admin-daemon $file perf dump**

The resulting JSON is parsed and grouped into collections, based on top-level key.  Top-level keys are
used as collection tags, and all sub-keys are flattened. For example:

```
 {
   "paxos": {
     "refresh": 9363435,
     "refresh_latency": {
       "avgcount": 9363435,
       "sum": 5378.794002000
     }
   }
 }
```

Would be parsed into the following metrics, all of which would be tagged with collection=paxos:

 - refresh = 9363435
 - refresh_latency.avgcount: 9363435
 - refresh_latency.sum: 5378.794002000


*Cluster Stats*

This gatherer works by invoking ceph commands against the cluster thus only requires the ceph client, valid
ceph configuration and an access key to function (the ceph_config and ceph_user configuration variables work
in conjunction to specify these prerequisites). It may be run on any server you wish which has access to
the cluster.  The currently supported commands are:

* ceph status
* ceph df
* ceph osd pool stats

### Configuration:

```
# Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster.
[[inputs.ceph]]
  ## This is the recommended interval to poll.  Too frequent and you will lose
  ## data points due to timeouts during rebalancing and recovery
  interval = '1m'

  ## All configuration values are optional, defaults are shown below

  ## location of ceph binary
  ceph_binary = "/usr/bin/ceph"

  ## directory in which to look for socket files
  socket_dir = "/var/run/ceph"

  ## prefix of MON and OSD socket files, used to determine socket type
  mon_prefix = "ceph-mon"
  osd_prefix = "ceph-osd"

  ## suffix used to identify socket files
  socket_suffix = "asok"

  ## Ceph user to authenticate as, ceph will search for the corresponding keyring
  ## e.g. client.admin.keyring in /etc/ceph, or the explicit path defined in the
  ## client section of ceph.conf for example:
  ##
  ##     [client.telegraf]
  ##         keyring = /etc/ceph/client.telegraf.keyring
  ##
  ## Consult the ceph documentation for more detail on keyring generation.
  ceph_user = "client.admin"

  ## Ceph configuration to use to locate the cluster
  ceph_config = "/etc/ceph/ceph.conf"

  ## Whether to gather statistics via the admin socket
  gather_admin_socket_stats = true

  ## Whether to gather statistics via ceph commands, requires ceph_user and ceph_config
  ## to be specified
  gather_cluster_stats = false
```

### Metrics:

*Admin Socket Stats*

All fields are collected under the **ceph** measurement and stored as float64s. For a full list of fields, see the sample perf dumps in ceph_test.go.

All admin measurements will have the following tags:

- type: either 'osd' or 'mon' to indicate which type of node was queried
- id: a unique string identifier, parsed from the socket file name for the node
- collection: the top-level key under which these fields were reported. Possible values are:
  - for MON nodes:
    - cluster
    - leveldb
    - mon
    - paxos
    - throttle-mon_client_bytes
    - throttle-mon_daemon_bytes
    - throttle-msgr_dispatch_throttler-mon
  - for OSD nodes:
    - WBThrottle
    - filestore
    - leveldb
    - mutex-FileJournal::completions_lock
    - mutex-FileJournal::finisher_lock
    - mutex-FileJournal::write_lock
    - mutex-FileJournal::writeq_lock
    - mutex-JOS::ApplyManager::apply_lock
    - mutex-JOS::ApplyManager::com_lock
    - mutex-JOS::SubmitManager::lock
    - mutex-WBThrottle::lock
    - objecter
    - osd
    - recoverystate_perf
    - throttle-filestore_bytes
    - throttle-filestore_ops
    - throttle-msgr_dispatch_throttler-client
    - throttle-msgr_dispatch_throttler-cluster
    - throttle-msgr_dispatch_throttler-hb_back_server
    - throttle-msgr_dispatch_throttler-hb_front_serve
    - throttle-msgr_dispatch_throttler-hbclient
    - throttle-msgr_dispatch_throttler-ms_objecter
    - throttle-objecter_bytes
    - throttle-objecter_ops
    - throttle-osd_client_bytes
    - throttle-osd_client_messages

*Cluster Stats*

+ ceph_health
  - fields:
    - status
    - overall_status

- ceph_osdmap
  - fields:
    - epoch (float)
    - num_osds (float)
    - num_up_osds (float)
    - num_in_osds (float)
    - full (bool)
    - nearfull (bool)
    - num_remapped_pgs (float)

+ ceph_pgmap
  - fields:
    - version (float)
    - num_pgs (float)
    - data_bytes (float)
    - bytes_used (float)
    - bytes_avail (float)
    - bytes_total (float)
    - read_bytes_sec (float)
    - write_bytes_sec (float)
    - op_per_sec (float, exists only in ceph <10)
    - read_op_per_sec (float)
    - write_op_per_sec (float)

- ceph_pgmap_state
  - tags:
    - state
  - fields:
    - count (float)

+ ceph_usage
  - fields:
    - total_bytes (float)
    - total_used_bytes (float)
    - total_avail_bytes (float)
    - total_space (float, exists only in ceph <0.84)
    - total_used (float, exists only in ceph <0.84)
    - total_avail (float, exists only in ceph <0.84)

- ceph_pool_usage
  - tags:
    - name
  - fields:
    - kb_used (float)
    - bytes_used (float)
    - objects (float)
    - percent_used (float)
    - max_avail (float)

+ ceph_pool_stats
  - tags:
    - name
  - fields:
    - read_bytes_sec (float)
    - write_bytes_sec (float)
    - op_per_sec (float, exists only in ceph <10)
    - read_op_per_sec (float)
    - write_op_per_sec (float)
    - recovering_objects_per_sec (float)
    - recovering_bytes_per_sec (float)
    - recovering_keys_per_sec (float)


### Example Output:

*Cluster Stats*

```
ceph_pool_stats,name=telegraf recovering_keys_per_sec=0,read_bytes_sec=0,write_bytes_sec=0,read_op_per_sec=0,write_op_per_sec=0,recovering_objects_per_sec=0,recovering_bytes_per_sec=0 1550658911000000000
ceph_pool_usage,name=telegraf kb_used=0,bytes_used=0,objects=0 1550658911000000000
ceph_pgmap_state,state=undersized+peered count=30 1550658910000000000
ceph_pgmap bytes_total=10733223936,read_op_per_sec=0,write_op_per_sec=0,num_pgs=30,data_bytes=0,bytes_avail=9654697984,read_bytes_sec=0,write_bytes_sec=0,version=0,bytes_used=1078525952 1550658910000000000
ceph_osdmap num_up_osds=1,num_in_osds=1,full=false,nearfull=false,num_remapped_pgs=0,epoch=34,num_osds=1 1550658910000000000
ceph_health status="HEALTH_WARN",overall_status="HEALTH_WARN" 1550658910000000000
```

*Admin Socket Stats*

```
ceph,collection=recoverystate_perf,id=0,type=osd reprecovering_latency.avgtime=0,repwaitrecoveryreserved_latency.avgcount=0,waitlocalbackfillreserved_latency.sum=0,reset_latency.avgtime=0.000090333,peering_latency.avgtime=0.824434333,stray_latency.avgtime=0.000030502,waitlocalrecoveryreserved_latency.sum=0,backfilling_latency.avgtime=0,reprecovering_latency.avgcount=0,incomplete_latency.avgtime=0,down_latency.avgtime=0,recovered_latency.sum=0.009692406,peering_latency.avgcount=40,notrecovering_latency.sum=0,waitremoterecoveryreserved_latency.sum=0,reprecovering_latency.sum=0,waitlocalbackfillreserved_latency.avgtime=0,started_latency.sum=9066.701648888,backfilling_latency.sum=0,waitactingchange_latency.avgcount=0,start_latency.avgtime=0.000030178,recovering_latency.avgtime=0,notbackfilling_latency.avgcount=0,waitremotebackfillreserved_latency.avgtime=0,incomplete_latency.avgcount=0,replicaactive_latency.sum=0,getinfo_latency.avgtime=0.000025945,down_latency.sum=0,recovered_latency.avgcount=40,waitactingchange_latency.avgtime=0,notrecovering_latency.avgcount=0,waitupthru_latency.sum=32.970965509,waitupthru_latency.avgtime=0.824274137,waitlocalrecoveryreserved_latency.avgcount=0,waitremoterecoveryreserved_latency.avgcount=0,activating_latency.avgcount=40,activating_latency.sum=0.83428466,activating_latency.avgtime=0.020857116,start_latency.avgcount=50,waitremotebackfillreserved_latency.avgcount=0,down_latency.avgcount=0,started_latency.avgcount=10,getlog_latency.avgcount=40,stray_latency.avgcount=10,notbackfilling_latency.sum=0,reset_latency.sum=0.00451665,active_latency.avgtime=906.505839265,repwaitbackfillreserved_latency.sum=0,waitactingchange_latency.sum=0,stray_latency.sum=0.000305022,waitremotebackfillreserved_latency.sum=0,repwaitrecoveryreserved_latency.avgtime=0,replicaactive_latency.avgtime=0,clean_latency.avgcount=10,waitremoterecoveryreserved_latency.avgtime=0,active_latency.avgcount=10,primary_latency.sum=9066.700828729,initial_latency.avgtime=0.000379351,waitlocalbackfillreserved_latency.avgcount=0,getinfo_latency.sum=0.001037815,reset_latency.avgcount=50,getlog_latency.sum=0.003079344,getlog_latency.avgtime=0.000076983,primary_latency.avgcount=10,repnotrecovering_latency.avgcount=0,initial_latency.sum=0.015174072,repwaitrecoveryreserved_latency.sum=0,replicaactive_latency.avgcount=0,clean_latency.avgtime=906.495755946,waitupthru_latency.avgcount=40,repnotrecovering_latency.sum=0,incomplete_latency.sum=0,active_latency.sum=9065.058392651,peering_latency.sum=32.977373355,repnotrecovering_latency.avgtime=0,notrecovering_latency.avgtime=0,waitlocalrecoveryreserved_latency.avgtime=0,repwaitbackfillreserved_latency.avgtime=0,recovering_latency.sum=0,getmissing_latency.sum=0.000902014,getmissing_latency.avgtime=0.00002255,clean_latency.sum=9064.957559467,getinfo_latency.avgcount=40,started_latency.avgtime=906.670164888,getmissing_latency.avgcount=40,notbackfilling_latency.avgtime=0,initial_latency.avgcount=40,recovered_latency.avgtime=0.00024231,repwaitbackfillreserved_latency.avgcount=0,backfilling_latency.avgcount=0,start_latency.sum=0.001508937,primary_latency.avgtime=906.670082872,recovering_latency.avgcount=0 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-hb_back_server,id=0,type=osd put_sum=0,wait.avgtime=0,put=0,get_or_fail_success=0,wait.avgcount=0,val=0,get_sum=0,take=0,take_sum=0,max=104857600,get=0,get_or_fail_fail=0,wait.sum=0,get_started=0 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-hb_front_client,id=0,type=osd wait.sum=0,val=0,take_sum=0,put=0,get_or_fail_success=0,put_sum=0,get=0,get_or_fail_fail=0,get_started=0,get_sum=0,wait.avgcount=0,wait.avgtime=0,max=104857600,take=0 1550658950000000000
ceph,collection=bluefs,id=0,type=osd slow_used_bytes=0,wal_total_bytes=0,gift_bytes=1048576,log_compactions=0,logged_bytes=221184,files_written_sst=1,slow_total_bytes=0,bytes_written_wal=619403,bytes_written_sst=1517,reclaim_bytes=0,db_total_bytes=1086324736,wal_used_bytes=0,log_bytes=319488,num_files=10,files_written_wal=1,db_used_bytes=12582912 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-ms_objecter,id=0,type=osd val=0,put=0,get=0,take=0,put_sum=0,get_started=0,take_sum=0,get_sum=0,wait.sum=0,wait.avgtime=0,get_or_fail_fail=0,get_or_fail_success=0,wait.avgcount=0,max=104857600 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-client,id=0,type=osd put=100,max=104857600,wait.sum=0,wait.avgtime=0,get_or_fail_fail=0,take_sum=0,val=0,wait.avgcount=0,get_sum=48561,get_or_fail_success=100,take=0,put_sum=48561,get_started=0,get=100 1550658950000000000
ceph,collection=mutex-OSDShard.2::sdata_wait_lock,id=0,type=osd wait.sum=0,wait.avgtime=0,wait.avgcount=0 1550658950000000000
ceph,collection=throttle-objecter_ops,id=0,type=osd get_or_fail_fail=0,max=1024,get_sum=0,take=0,val=0,wait.avgtime=0,get_or_fail_success=0,wait.sum=0,put_sum=0,get=0,take_sum=0,put=0,wait.avgcount=0,get_started=0 1550658950000000000
ceph,collection=AsyncMessenger::Worker-1,id=0,type=osd msgr_send_messages=266,msgr_recv_bytes=49074,msgr_active_connections=1,msgr_running_recv_time=0.136317251,msgr_running_fast_dispatch_time=0,msgr_created_connections=5,msgr_send_bytes=41569,msgr_running_send_time=0.514432253,msgr_recv_messages=81,msgr_running_total_time=0.766790051 1550658950000000000
ceph,collection=throttle-bluestore_throttle_deferred_bytes,id=0,type=osd get_started=0,wait.sum=0,wait.avgcount=0,take_sum=0,val=12134038,max=201326592,take=0,get_or_fail_fail=0,put_sum=0,wait.avgtime=0,get_or_fail_success=18,get=18,get_sum=12134038,put=0 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-hb_front_server,id=0,type=osd get=0,put_sum=0,val=0,get_or_fail_fail=0,get_or_fail_success=0,take=0,max=104857600,get_started=0,wait.sum=0,wait.avgtime=0,get_sum=0,take_sum=0,put=0,wait.avgcount=0 1550658950000000000
ceph,collection=mutex-OSDShard.1::sdata_wait_lock,id=0,type=osd wait.avgcount=0,wait.sum=0,wait.avgtime=0 1550658950000000000
ceph,collection=finisher-defered_finisher,id=0,type=osd queue_len=0,complete_latency.avgcount=0,complete_latency.sum=0,complete_latency.avgtime=0 1550658950000000000
ceph,collection=mutex-OSDShard.3::shard_lock,id=0,type=osd wait.avgtime=0,wait.avgcount=0,wait.sum=0 1550658950000000000
ceph,collection=mutex-OSDShard.0::shard_lock,id=0,type=osd wait.avgcount=0,wait.sum=0,wait.avgtime=0 1550658950000000000
ceph,collection=throttle-osd_client_bytes,id=0,type=osd get_or_fail_fail=0,get=22,get_sum=6262,take=0,max=524288000,put=31,wait.sum=0,val=0,get_started=0,put_sum=6262,get_or_fail_success=22,take_sum=0,wait.avgtime=0,wait.avgcount=0 1550658950000000000
ceph,collection=rocksdb,id=0,type=osd submit_latency.sum=0.019985172,rocksdb_write_pre_and_post_time.sum=0,rocksdb_write_wal_time.avgtime=0,rocksdb_write_delay_time.avgtime=0,rocksdb_write_pre_and_post_time.avgtime=0,rocksdb_write_pre_and_post_time.avgcount=0,submit_sync_latency.sum=0.559604552,compact=0,compact_queue_len=0,get_latency.avgcount=140,submit_latency.avgtime=0.000095622,submit_transaction=209,compact_range=0,rocksdb_write_wal_time.avgcount=0,submit_sync_latency.avgtime=0.011906479,compact_queue_merge=0,rocksdb_write_memtable_time.avgtime=0,get_latency.sum=0.013135139,submit_latency.avgcount=209,submit_sync_latency.avgcount=47,submit_transaction_sync=47,rocksdb_write_wal_time.sum=0,rocksdb_write_delay_time.avgcount=0,rocksdb_write_memtable_time.avgcount=0,rocksdb_write_memtable_time.sum=0,get=140,get_latency.avgtime=0.000093822,rocksdb_write_delay_time.sum=0 1550658950000000000
ceph,collection=mutex-OSDShard.1::shard_lock,id=0,type=osd wait.avgcount=0,wait.sum=0,wait.avgtime=0 1550658950000000000
ceph,collection=osd,id=0,type=osd subop_latency.avgtime=0,copyfrom=0,osd_pg_info=140,subop_push_latency.avgtime=0,subop_pull=0,op_rw_process_latency.sum=0,stat_bytes=10733223936,numpg_removing=0,op_latency.avgtime=0,op_w_process_latency.avgtime=0,op_rw_in_bytes=0,osd_map_cache_miss=0,loadavg=144,map_messages=31,op_w_latency.avgtime=0,op_prepare_latency.avgcount=0,op_r=0,op_latency.avgcount=0,osd_map_cache_hit=225,op_w_prepare_latency.sum=0,numpg_primary=30,op_rw_out_bytes=0,subop_w_latency.avgcount=0,subop_push_latency.avgcount=0,op_r_process_latency.avgcount=0,op_w_in_bytes=0,op_rw_latency.avgtime=0,subop_w_latency.avgtime=0,osd_map_cache_miss_low_avg.sum=0,agent_wake=0,op_before_queue_op_lat.avgtime=0.000065043,op_w_prepare_latency.avgcount=0,tier_proxy_write=0,op_rw_prepare_latency.avgtime=0,op_rw_process_latency.avgtime=0,op_in_bytes=0,op_cache_hit=0,tier_whiteout=0,op_w_prepare_latency.avgtime=0,heartbeat_to_peers=0,object_ctx_cache_hit=0,buffer_bytes=0,stat_bytes_avail=9654697984,op_w_latency.avgcount=0,tier_dirty=0,tier_flush_fail=0,op_rw_prepare_latency.avgcount=0,agent_flush=0,osd_tier_promote_lat.sum=0,subop_w_latency.sum=0,tier_promote=0,op_before_dequeue_op_lat.avgcount=22,push=0,tier_flush=0,osd_pg_biginfo=90,tier_try_flush_fail=0,subop_push_in_bytes=0,op_before_dequeue_op_lat.sum=0.00266744,osd_map_cache_miss_low=0,numpg=30,op_prepare_latency.avgtime=0,subop_pull_latency.avgtime=0,op_rw_latency.avgcount=0,subop_latency.avgcount=0,op=0,osd_tier_promote_lat.avgcount=0,cached_crc=0,op_r_prepare_latency.sum=0,subop_pull_latency.sum=0,op_before_dequeue_op_lat.avgtime=0.000121247,history_alloc_Mbytes=0,subop_push_latency.sum=0,subop_in_bytes=0,op_w_process_latency.sum=0,osd_map_cache_miss_low_avg.avgcount=0,subop=0,tier_clean=0,osd_tier_r_lat.avgtime=0,op_r_process_latency.avgtime=0,op_r_prepare_latency.avgcount=0,op_w_process_latency.avgcount=0,numpg_stray=0,op_r_prepare_latency.avgtime=0,object_ctx_cache_total=0,op_process_latency.avgtime=0,op_r_process_latency.sum=0,op_r_latency.sum=0,subop_w_in_bytes=0,op_rw=0,messages_delayed_for_map=4,map_message_epoch_dups=30,osd_map_bl_cache_miss=33,op_r_latency.avgtime=0,op_before_queue_op_lat.sum=0.001430955,map_message_epochs=64,agent_evict=0,op_out_bytes=0,op_process_latency.sum=0,osd_tier_flush_lat.sum=0,stat_bytes_used=1078525952,op_prepare_latency.sum=0,op_wip=0,osd_tier_flush_lat.avgtime=0,missed_crc=0,op_rw_latency.sum=0,op_r_latency.avgcount=0,pull=0,op_w_latency.sum=0,op_before_queue_op_lat.avgcount=22,tier_try_flush=0,numpg_replica=0,subop_push=0,osd_tier_r_lat.sum=0,op_latency.sum=0,push_out_bytes=0,op_w=0,osd_tier_promote_lat.avgtime=0,subop_latency.sum=0,osd_pg_fastinfo=0,tier_delay=0,op_rw_prepare_latency.sum=0,osd_tier_flush_lat.avgcount=0,osd_map_bl_cache_hit=0,op_r_out_bytes=0,subop_pull_latency.avgcount=0,op_process_latency.avgcount=0,tier_evict=0,tier_proxy_read=0,agent_skip=0,subop_w=0,history_alloc_num=0,osd_tier_r_lat.avgcount=0,recovery_ops=0,cached_crc_adjusted=0,op_rw_process_latency.avgcount=0 1550658950000000000
ceph,collection=finisher-finisher-0,id=0,type=osd complete_latency.sum=0.015491438,complete_latency.avgtime=0.000174061,complete_latency.avgcount=89,queue_len=0 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-hb_back_client,id=0,type=osd wait.avgtime=0,wait.avgcount=0,max=104857600,get_sum=0,take=0,get_or_fail_fail=0,val=0,get=0,get_or_fail_success=0,wait.sum=0,put=0,take_sum=0,get_started=0,put_sum=0 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-cluster,id=0,type=osd get_sum=0,take=0,val=0,max=104857600,get_or_fail_success=0,put=0,put_sum=0,wait.sum=0,wait.avgtime=0,get_started=0,get_or_fail_fail=0,take_sum=0,wait.avgcount=0,get=0 1550658950000000000
ceph,collection=mutex-OSDShard.0::sdata_wait_lock,id=0,type=osd wait.avgcount=0,wait.sum=0,wait.avgtime=0 1550658950000000000
ceph,collection=throttle-bluestore_throttle_bytes,id=0,type=osd get_sum=140287253,put_sum=140287253,get=209,put=47,val=0,get_started=209,wait.sum=0,wait.avgcount=0,wait.avgtime=0,max=67108864,get_or_fail_fail=0,take=0,take_sum=0,get_or_fail_success=0 1550658950000000000
ceph,collection=objecter,id=0,type=osd map_inc=15,op_w=0,osd_session_close=0,op=0,osdop_writefull=0,osdop_tmap_up=0,command_resend=0,poolstat_resend=0,osdop_setxattr=0,osdop_append=0,osdop_delete=0,op_rmw=0,poolstat_send=0,op_active=0,osdop_tmap_put=0,osdop_clonerange=0,osdop_rmxattr=0,op_send=0,op_resend=0,osdop_resetxattrs=0,osdop_call=0,osdop_pgls=0,poolstat_active=0,linger_resend=0,osdop_stat=0,op_reply=0,op_laggy=0,statfs_send=0,osdop_getxattr=0,osdop_pgls_filter=0,osdop_notify=0,linger_active=0,osdop_other=0,poolop_resend=0,statfs_active=0,command_active=0,map_epoch=34,osdop_create=0,osdop_watch=0,op_r=0,map_full=0,osdop_src_cmpxattr=0,omap_rd=0,osd_session_open=0,osdop_sparse_read=0,osdop_truncate=0,linger_ping=0,osdop_mapext=0,poolop_send=0,osdop_cmpxattr=0,osd_laggy=0,osdop_writesame=0,osd_sessions=0,osdop_tmap_get=0,op_pg=0,command_send=0,osdop_read=0,op_send_bytes=0,statfs_resend=0,omap_del=0,poolop_active=0,osdop_write=0,osdop_zero=0,omap_wr=0,linger_send=0 1550658950000000000
ceph,collection=mutex-OSDShard.4::shard_lock,id=0,type=osd wait.avgtime=0,wait.avgcount=0,wait.sum=0 1550658950000000000
ceph,collection=AsyncMessenger::Worker-0,id=0,type=osd msgr_recv_messages=112,msgr_recv_bytes=14550,msgr_created_connections=15,msgr_running_recv_time=0.026754699,msgr_active_connections=11,msgr_send_messages=11,msgr_running_fast_dispatch_time=0.003373472,msgr_send_bytes=2090,msgr_running_total_time=0.041323592,msgr_running_send_time=0.000441856 1550658950000000000
ceph,collection=mutex-OSDShard.2::shard_lock,id=0,type=osd wait.sum=0,wait.avgtime=0,wait.avgcount=0 1550658950000000000
ceph,collection=bluestore,id=0,type=osd submit_lat.avgcount=209,kv_flush_lat.avgtime=0.000002175,bluestore_write_big_bytes=0,bluestore_txc=209,kv_commit_lat.avgcount=47,kv_commit_lat.sum=0.585164754,bluestore_buffer_miss_bytes=511,commit_lat.avgcount=209,bluestore_buffer_bytes=0,bluestore_onodes=102,state_kv_queued_lat.sum=1.439223859,deferred_write_bytes=0,bluestore_write_small_bytes=60279,decompress_lat.sum=0,state_kv_done_lat.avgcount=209,submit_lat.sum=0.055637603,state_prepare_lat.avgcount=209,bluestore_write_big=0,read_wait_aio_lat.avgcount=17,bluestore_write_small_deferred=18,kv_lat.sum=0.585267001,kv_flush_lat.sum=0.000102247,bluestore_buffers=0,state_prepare_lat.sum=0.051411998,bluestore_write_small_pre_read=18,state_deferred_queued_lat.sum=0,decompress_lat.avgtime=0,state_kv_done_lat.avgtime=0.000000629,bluestore_write_small_unused=0,read_lat.avgcount=34,bluestore_onode_shard_misses=0,bluestore_blobs=72,bluestore_read_eio=0,bluestore_blob_split=0,bluestore_onode_shard_hits=0,state_kv_commiting_lat.avgcount=209,bluestore_onode_hits=153,state_kv_commiting_lat.sum=2.477385041,read_onode_meta_lat.avgcount=51,state_finishing_lat.avgtime=0.000000489,bluestore_compressed_original=0,state_kv_queued_lat.avgtime=0.006886238,bluestore_gc_merged=0,throttle_lat.avgtime=0.000001247,state_aio_wait_lat.avgtime=0.000001326,bluestore_onode_reshard=0,state_done_lat.avgcount=191,bluestore_compressed_allocated=0,write_penalty_read_ops=0,bluestore_extents=72,compress_lat.avgtime=0,state_aio_wait_lat.avgcount=209,state_io_done_lat.avgtime=0.000000519,bluestore_write_big_blobs=0,state_kv_queued_lat.avgcount=209,kv_flush_lat.avgcount=47,state_finishing_lat.sum=0.000093565,state_io_done_lat.avgcount=209,kv_lat.avgtime=0.012452489,bluestore_buffer_hit_bytes=20750,read_wait_aio_lat.avgtime=0.000038077,bluestore_allocated=4718592,state_deferred_cleanup_lat.avgtime=0,compress_lat.avgcount=0,write_pad_bytes=304265,throttle_lat.sum=0.000260785,read_onode_meta_lat.avgtime=0.000038702,compress_success_count=0,state_deferred_aio_wait_lat.sum=0,decompress_lat.avgcount=0,state_deferred_aio_wait_lat.avgtime=0,bluestore_stored=51133,state_finishing_lat.avgcount=191,bluestore_onode_misses=132,deferred_write_ops=0,read_wait_aio_lat.sum=0.000647315,csum_lat.avgcount=1,state_kv_done_lat.sum=0.000131531,state_prepare_lat.avgtime=0.00024599,state_deferred_cleanup_lat.avgcount=0,state_deferred_queued_lat.avgcount=0,bluestore_reads_with_retries=0,state_kv_commiting_lat.avgtime=0.011853516,kv_commit_lat.avgtime=0.012450313,read_lat.sum=0.003031418,throttle_lat.avgcount=209,bluestore_write_small_new=71,state_deferred_queued_lat.avgtime=0,bluestore_extent_compress=0,bluestore_write_small=89,state_deferred_cleanup_lat.sum=0,submit_lat.avgtime=0.000266208,bluestore_fragmentation_micros=0,state_aio_wait_lat.sum=0.000277323,commit_lat.avgtime=0.018987901,compress_lat.sum=0,bluestore_compressed=0,state_done_lat.sum=0.000206953,csum_lat.avgtime=0.000023281,state_deferred_aio_wait_lat.avgcount=0,compress_rejected_count=0,kv_lat.avgcount=47,read_onode_meta_lat.sum=0.001973812,read_lat.avgtime=0.000089159,csum_lat.sum=0.000023281,state_io_done_lat.sum=0.00010855,state_done_lat.avgtime=0.000001083,commit_lat.sum=3.96847136 1550658950000000000
ceph,collection=mutex-OSDShard.3::sdata_wait_lock,id=0,type=osd wait.avgcount=0,wait.sum=0,wait.avgtime=0 1550658950000000000
ceph,collection=AsyncMessenger::Worker-2,id=0,type=osd msgr_running_fast_dispatch_time=0,msgr_recv_bytes=246,msgr_created_connections=5,msgr_active_connections=1,msgr_running_recv_time=0.001392218,msgr_running_total_time=1.934101301,msgr_running_send_time=1.781171967,msgr_recv_messages=3,msgr_send_bytes=26504031,msgr_send_messages=15409 1550658950000000000
ceph,collection=finisher-objecter-finisher-0,id=0,type=osd complete_latency.avgcount=0,complete_latency.sum=0,complete_latency.avgtime=0,queue_len=0 1550658950000000000
ceph,collection=mutex-OSDShard.4::sdata_wait_lock,id=0,type=osd wait.avgcount=0,wait.sum=0,wait.avgtime=0 1550658950000000000
ceph,collection=throttle-objecter_bytes,id=0,type=osd take=0,get_sum=0,put_sum=0,put=0,val=0,get=0,get_or_fail_fail=0,wait.avgcount=0,get_or_fail_success=0,wait.sum=0,wait.avgtime=0,get_started=0,max=104857600,take_sum=0 1550658950000000000
ceph,collection=throttle-mon_client_bytes,id=test,type=monitor get_or_fail_fail=0,take_sum=0,wait.avgtime=0,wait.avgcount=0,get_sum=64607,take=0,get_started=0,put=950,val=240,wait.sum=0,max=104857600,get_or_fail_success=953,put_sum=64367,get=953 1550658950000000000
ceph,collection=mon,id=test,type=monitor election_win=1,election_lose=0,num_sessions=3,session_add=199,session_rm=196,session_trim=0,num_elections=1,election_call=0 1550658950000000000
ceph,collection=cluster,id=test,type=monitor num_pg_active=0,num_mon=1,osd_bytes_avail=9654697984,num_object=0,num_osd_in=1,osd_bytes_used=1078525952,num_bytes=0,num_osd=1,num_pg_peering=0,num_pg_active_clean=0,num_pg=30,num_mon_quorum=1,num_object_degraded=0,osd_bytes=10733223936,num_object_unfound=0,num_osd_up=1,num_pool=1,num_object_misplaced=0,osd_epoch=34 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-mon-mgrc,id=test,type=monitor get=2,put=2,get_sum=16,take_sum=0,wait.avgtime=0,val=0,wait.avgcount=0,get_or_fail_success=2,put_sum=16,max=104857600,get_started=0,take=0,get_or_fail_fail=0,wait.sum=0 1550658950000000000
ceph,collection=rocksdb,id=test,type=monitor rocksdb_write_memtable_time.avgtime=0,submit_sync_latency.avgtime=0.013689071,submit_transaction_sync=39173,rocksdb_write_pre_and_post_time.avgtime=0,get_latency.avgcount=724581,submit_latency.avgtime=0,submit_sync_latency.avgcount=39173,rocksdb_write_wal_time.avgtime=0,rocksdb_write_pre_and_post_time.sum=0,compact_range=231,compact_queue_merge=0,rocksdb_write_memtable_time.avgcount=0,submit_sync_latency.sum=536.242007888,compact=0,rocksdb_write_delay_time.sum=0,get_latency.sum=9.578173532,rocksdb_write_delay_time.avgcount=0,rocksdb_write_delay_time.avgtime=0,compact_queue_len=0,get_latency.avgtime=0.000013218,submit_latency.sum=0,get=724581,rocksdb_write_wal_time.avgcount=0,submit_transaction=0,rocksdb_write_wal_time.sum=0,submit_latency.avgcount=0,rocksdb_write_pre_and_post_time.avgcount=0,rocksdb_write_memtable_time.sum=0 1550658950000000000
ceph,collection=finisher-mon_finisher,id=test,type=monitor complete_latency.avgtime=0,complete_latency.avgcount=0,complete_latency.sum=0,queue_len=0 1550658950000000000
ceph,collection=paxos,id=test,type=monitor share_state_keys.sum=0,collect_keys.avgcount=0,collect=0,store_state_latency.avgtime=0,begin_latency.sum=338.90900364,collect_keys.sum=0,collect_bytes.avgcount=0,accept_timeout=0,new_pn_latency.avgcount=0,new_pn_latency.sum=0,commit_keys.sum=116820,share_state_bytes.sum=0,refresh_latency.avgcount=19576,store_state=0,collect_timeout=0,lease_ack_timeout=0,collect_latency.avgcount=0,store_state_keys.avgcount=0,commit_bytes.sum=38478195,refresh_latency.sum=8.341938952,collect_uncommitted=0,commit_latency.avgcount=19576,share_state=0,begin_latency.avgtime=0.017312474,commit_latency.avgtime=0.009926797,begin_keys.sum=58728,start_peon=0,commit_keys.avgcount=19576,begin_latency.avgcount=19576,store_state_latency.avgcount=0,start_leader=1,begin_keys.avgcount=19576,collect_bytes.sum=0,begin_bytes.avgcount=19576,store_state_bytes.sum=0,commit=19576,begin_bytes.sum=41771257,new_pn_latency.avgtime=0,refresh_latency.avgtime=0.00042613,commit_latency.sum=194.326980684,new_pn=0,refresh=19576,collect_latency.sum=0,collect_latency.avgtime=0,lease_timeout=0,begin=19576,share_state_bytes.avgcount=0,share_state_keys.avgcount=0,store_state_keys.sum=0,store_state_bytes.avgcount=0,store_state_latency.sum=0,commit_bytes.avgcount=19576,restart=2 1550658950000000000
ceph,collection=finisher-monstore,id=test,type=monitor complete_latency.avgcount=19576,complete_latency.sum=208.300976568,complete_latency.avgtime=0.01064063,queue_len=0 1550658950000000000
ceph,collection=AsyncMessenger::Worker-2,id=test,type=monitor msgr_created_connections=1,msgr_send_bytes=0,msgr_running_send_time=0,msgr_recv_bytes=0,msgr_send_messages=1,msgr_recv_messages=0,msgr_running_total_time=0.003026541,msgr_running_recv_time=0,msgr_running_fast_dispatch_time=0,msgr_active_connections=1 1550658950000000000
ceph,collection=throttle-msgr_dispatch_throttler-mon,id=test,type=monitor take=0,take_sum=0,put=39933,get=39933,put_sum=56745184,wait.avgtime=0,get_or_fail_success=39933,wait.sum=0,get_sum=56745184,get_or_fail_fail=0,wait.avgcount=0,val=0,max=104857600,get_started=0 1550658950000000000
ceph,collection=throttle-mon_daemon_bytes,id=test,type=monitor max=419430400,get_started=0,wait.avgtime=0,take_sum=0,get=262,take=0,put_sum=21212,wait.avgcount=0,get_or_fail_success=262,get_or_fail_fail=0,put=262,wait.sum=0,val=0,get_sum=21212 1550658950000000000
ceph,collection=AsyncMessenger::Worker-1,id=test,type=monitor msgr_send_messages=1071,msgr_running_total_time=0.703589077,msgr_active_connections=146,msgr_send_bytes=3887863,msgr_running_send_time=0.361602994,msgr_running_recv_time=0.328218119,msgr_running_fast_dispatch_time=0,msgr_recv_messages=978,msgr_recv_bytes=142209,msgr_created_connections=197 1550658950000000000
ceph,collection=AsyncMessenger::Worker-0,id=test,type=monitor msgr_created_connections=54,msgr_recv_messages=38957,msgr_active_connections=47,msgr_running_fast_dispatch_time=0,msgr_send_bytes=25338946,msgr_running_total_time=9.190267622,msgr_running_send_time=3.124663809,msgr_running_recv_time=13.03937269,msgr_send_messages=15973,msgr_recv_bytes=59558181 1550658950000000000
```
