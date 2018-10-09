# Ceph Storage Input Plugin

Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster.

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

  ## prefix of MON, OSD or RGW socket files, used to determine socket type
  mon_prefix = "ceph-mon"
  osd_prefix = "ceph-osd"
  rgw_prefix = "ceph-rgw"

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

### Measurements & Fields:

*Admin Socket Stats*

All fields are collected under the **ceph** measurement and stored as float64s. For a full list of fields, see the sample perf dumps in ceph_test.go.

*Cluster Stats*

* ceph\_osdmap
  * epoch (float)
  * full (boolean)
  * nearfull (boolean)
  * num\_in\_osds (float)
  * num\_osds (float)
  * num\_remremapped\_pgs (float)
  * num\_up\_osds (float)

* ceph\_pgmap
  * bytes\_avail (float)
  * bytes\_total (float)
  * bytes\_used (float)
  * data\_bytes (float)
  * num\_pgs (float)
  * op\_per\_sec (float)
  * read\_bytes\_sec (float)
  * version (float)
  * write\_bytes\_sec (float)
  * recovering\_bytes\_per\_sec (float)
  * recovering\_keys\_per\_sec (float)
  * recovering\_objects\_per\_sec (float)

* ceph\_pgmap\_state
  * count (float)

* ceph\_usage
  * bytes\_used (float)
  * kb\_used (float)
  * max\_avail (float)
  * objects (float)

* ceph\_pool\_usage
  * bytes\_used (float)
  * kb\_used (float)
  * max\_avail (float)
  * objects (float)

* ceph\_pool\_stats
  * op\_per\_sec (float)
  * read\_bytes\_sec (float)
  * write\_bytes\_sec (float)
  * recovering\_object\_per\_sec (float)
  * recovering\_bytes\_per\_sec (float)
  * recovering\_keys\_per\_sec (float)

### Tags:

*Admin Socket Stats*

All measurements will have the following tags:

- type: either 'osd', 'mon' or 'rgw' to indicate which type of node was queried
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
  - for RGW nodes:
    - rgw
    - filestore
    - cct
    - WBThrottle
    - objecter
    - finisher-radosclient
    - throttle-msgr_dispatch_throttler-radosclient
    - throttle-objecter_bytes
    - throttle-objecter_ops
    - throttle-rgw_async_rados_ops


*Cluster Stats*

* ceph\_pgmap\_state has the following tags:
  * state (state for which the value applies e.g. active+clean, active+remapped+backfill)
* ceph\_pool\_usage has the following tags:
  * id
  * name
* ceph\_pool\_stats has the following tags:
  * id
  * name

### Example Output:

*Admin Socket Stats*

<pre>
telegraf --config /etc/telegraf/telegraf.conf --config-directory /etc/telegraf/telegraf.d --input-filter ceph --test
* Plugin: ceph, Collection 1
> ceph,collection=paxos, id=node-2,role=openstack,type=mon accept_timeout=0,begin=14931264,begin_bytes.avgcount=14931264,begin_bytes.sum=180309683362,begin_keys.avgcount=0,begin_keys.sum=0,begin_latency.avgcount=14931264,begin_latency.sum=9293.29589,collect=1,collect_bytes.avgcount=1,collect_bytes.sum=24,collect_keys.avgcount=1,collect_keys.sum=1,collect_latency.avgcount=1,collect_latency.sum=0.00028,collect_timeout=0,collect_uncommitted=0,commit=14931264,commit_bytes.avgcount=0,commit_bytes.sum=0,commit_keys.avgcount=0,commit_keys.sum=0,commit_latency.avgcount=0,commit_latency.sum=0,lease_ack_timeout=0,lease_timeout=0,new_pn=0,new_pn_latency.avgcount=0,new_pn_latency.sum=0,refresh=14931264,refresh_latency.avgcount=14931264,refresh_latency.sum=8706.98498,restart=4,share_state=0,share_state_bytes.avgcount=0,share_state_bytes.sum=0,share_state_keys.avgcount=0,share_state_keys.sum=0,start_leader=0,start_peon=1,store_state=14931264,store_state_bytes.avgcount=14931264,store_state_bytes.sum=353119959211,store_state_keys.avgcount=14931264,store_state_keys.sum=289807523,store_state_latency.avgcount=14931264,store_state_latency.sum=10952.835724 1462821234814535148
> ceph,collection=throttle-mon_client_bytes,id=node-2,type=mon get=1413017,get_or_fail_fail=0,get_or_fail_success=0,get_sum=71211705,max=104857600,put=1413013,put_sum=71211459,take=0,take_sum=0,val=246,wait.avgcount=0,wait.sum=0 1462821234814737219
> ceph,collection=throttle-mon_daemon_bytes,id=node-2,type=mon get=4058121,get_or_fail_fail=0,get_or_fail_success=0,get_sum=6027348117,max=419430400,put=4058121,put_sum=6027348117,take=0,take_sum=0,val=0,wait.avgcount=0,wait.sum=0 1462821234814815661
> ceph,collection=throttle-msgr_dispatch_throttler-mon,id=node-2,type=mon get=54276277,get_or_fail_fail=0,get_or_fail_success=0,get_sum=370232877040,max=104857600,put=54276277,put_sum=370232877040,take=0,take_sum=0,val=0,wait.avgcount=0,wait.sum=0 1462821234814872064
</pre>

*Cluster Stats*

<pre>
> ceph_osdmap,host=ceph-mon-0 epoch=170772,full=false,nearfull=false,num_in_osds=340,num_osds=340,num_remapped_pgs=0,num_up_osds=340 1468841037000000000
> ceph_pgmap,host=ceph-mon-0 bytes_avail=634895531270144,bytes_total=812117151809536,bytes_used=177221620539392,data_bytes=56979991615058,num_pgs=22952,op_per_sec=15869,read_bytes_sec=43956026,version=39387592,write_bytes_sec=165344818 1468841037000000000
> ceph_pgmap_state,host=ceph-mon-0,state=active+clean count=22952 1468928660000000000
> ceph_pgmap_state,host=ceph-mon-0,state=active+degraded count=16 1468928660000000000
> ceph_usage,host=ceph-mon-0 total_avail_bytes=634895514791936,total_bytes=812117151809536,total_used_bytes=177221637017600 1468841037000000000
> ceph_pool_usage,host=ceph-mon-0,id=150,name=cinder.volumes bytes_used=12648553794802,kb_used=12352103316,max_avail=154342562489244,objects=3026295 1468841037000000000
> ceph_pool_usage,host=ceph-mon-0,id=182,name=cinder.volumes.flash bytes_used=8541308223964,kb_used=8341121313,max_avail=39388593563936,objects=2075066 1468841037000000000
> ceph_pool_stats,host=ceph-mon-0,id=150,name=cinder.volumes op_per_sec=1706,read_bytes_sec=28671674,write_bytes_sec=29994541 1468841037000000000
> ceph_pool_stats,host=ceph-mon-0,id=182,name=cinder.volumes.flash op_per_sec=9748,read_bytes_sec=9605524,write_bytes_sec=45593310 1468841037000000000
</pre>

*RGW Stats*

<pre>
# telegraf --config /etc/telegraf/telegraf.conf  --test  | grep type=rgw
> ceph,collection=throttle-objecter_bytes-0x55a6473ac5a0,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw get=2,get_or_fail_fail=0,get_or_fail_success=2,get_started=0,get_sum=0,max=104857600,put=0,put_sum=0,take=0,take_sum=0,val=0,wait.avgcount=0,wait.avgtime=0,wait.sum=0 1539079060000000000
> ceph,collection=rgw,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw cache_hit=9752756,cache_miss=2359127,failed_req=135622,get=69792,get_b=12381229,get_initial_lat.avgcount=67225,get_initial_lat.avgtime=0.104983629,get_initial_lat.sum=7057.524503277,keystone_token_cache_hit=0,keystone_token_cache_miss=0,put=1770460,put_b=102378362473,put_initial_lat.avgcount=1770460,put_initial_lat.avgtime=0.008662659,put_initial_lat.sum=15336.891717227,qactive=0,qlen=0,req=2389873 1539079060000000000
> ceph,collection=cct,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw total_workers=32,unhealthy_workers=0 1539079060000000000
> ceph,collection=throttle-objecter_ops,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw get=1869628,get_or_fail_fail=0,get_or_fail_success=1869628,get_started=0,get_sum=1869628,max=1024,put=1869628,put_sum=1869628,take=0,take_sum=0,val=0,wait.avgcount=0,wait.avgtime=0,wait.sum=0 1539079060000000000
> ceph,collection=throttle-objecter_bytes,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw get=1869628,get_or_fail_fail=0,get_or_fail_success=1869628,get_started=0,get_sum=225229924728,max=104857600,put=486561,put_sum=225229924728,take=0,take_sum=0,val=0,wait.avgcount=0,wait.avgtime=0,wait.sum=0 1539079060000000000
> ceph,collection=finisher-radosclient,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw complete_latency.avgcount=440449,complete_latency.avgtime=0.000017472,complete_latency.sum=7.695941099,queue_len=0 1539079060000000000
> ceph,collection=throttle-rgw_async_rados_ops,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw get=832,get_or_fail_fail=0,get_or_fail_success=0,get_started=832,get_sum=832,max=64,put=832,put_sum=832,take=0,take_sum=0,val=0,wait.avgcount=0,wait.avgtime=0,wait.sum=0 1539079060000000000
> ceph,collection=throttle-msgr_dispatch_throttler-radosclient,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw get=2123420,get_or_fail_fail=0,get_or_fail_success=2123420,get_started=0,get_sum=16043329227,max=104857600,put=2123420,put_sum=16043329227,take=0,take_sum=0,val=0,wait.avgcount=0,wait.avgtime=0,wait.sum=0 1539079060000000000
> ceph,collection=AsyncMessenger::Worker-1,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw msgr_active_connections=48,msgr_created_connections=2637,msgr_recv_bytes=17465858940,msgr_recv_messages=3889947,msgr_running_fast_dispatch_time=228.248315303,msgr_running_recv_time=228.93186848,msgr_running_send_time=338.516899023,msgr_running_total_time=831.177209327,msgr_send_bytes=44182125553,msgr_send_messages=4029669 1539079060000000000
> ceph,collection=finisher-radosclient-0x55a6473ac500,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw complete_latency.avgcount=2,complete_latency.avgtime=0.000201649,complete_latency.sum=0.000403299,queue_len=0 1539079060000000000
> ceph,collection=objecter-0x55a6473ac780,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw command_active=0,command_resend=0,command_send=0,linger_active=1,linger_ping=67795,linger_resend=0,linger_send=1,map_epoch=11181,map_full=0,map_inc=8,omap_del=0,omap_rd=0,omap_wr=0,op=67798,op_active=0,op_laggy=0,op_pg=0,op_r=2,op_reply=67798,op_resend=0,op_rmw=0,op_send=67798,op_send_bytes=40,op_w=67796,osd_laggy=1,osd_session_close=0,osd_session_open=1,osd_sessions=1,osdop_append=0,osdop_call=0,osdop_clonerange=0,osdop_cmpxattr=0,osdop_create=0,osdop_delete=0,osdop_getxattr=0,osdop_mapext=0,osdop_notify=0,osdop_other=2,osdop_pgls=0,osdop_pgls_filter=0,osdop_read=0,osdop_resetxattrs=0,osdop_rmxattr=0,osdop_setxattr=0,osdop_sparse_read=0,osdop_src_cmpxattr=0,osdop_stat=0,osdop_tmap_get=0,osdop_tmap_put=0,osdop_tmap_up=0,osdop_truncate=0,osdop_watch=67796,osdop_write=0,osdop_writefull=0,osdop_writesame=0,osdop_zero=0,poolop_active=0,poolop_resend=0,poolop_send=0,poolstat_active=0,poolstat_resend=0,poolstat_send=0,statfs_active=0,statfs_resend=0,statfs_send=0 1539079060000000000
> ceph,collection=throttle-objecter_ops-0x55a6473ac640,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw get=2,get_or_fail_fail=0,get_or_fail_success=2,get_started=0,get_sum=2,max=1024,put=2,put_sum=2,take=0,take_sum=0,val=0,wait.avgcount=0,wait.avgtime=0,wait.sum=0 1539079060000000000
> ceph,collection=AsyncMessenger::Worker-0,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw msgr_active_connections=51,msgr_created_connections=2782,msgr_recv_bytes=64034035231,msgr_recv_messages=7435504,msgr_running_fast_dispatch_time=650.424761206,msgr_running_recv_time=464.60981172,msgr_running_send_time=595.278626436,msgr_running_total_time=1775.976613711,msgr_send_bytes=62255370946,msgr_send_messages=7463847 1539079060000000000
> ceph,collection=AsyncMessenger::Worker-2,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw msgr_active_connections=52,msgr_created_connections=2676,msgr_recv_bytes=3016239158,msgr_recv_messages=4969733,msgr_running_fast_dispatch_time=151.060846403,msgr_running_recv_time=270.232489998,msgr_running_send_time=447.379763916,msgr_running_total_time=912.514276361,msgr_send_bytes=48266414577,msgr_send_messages=5019305 1539079060000000000
> ceph,collection=throttle-msgr_dispatch_throttler-radosclient-0x55a6473ac6e0,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw get=67955,get_or_fail_fail=0,get_or_fail_success=67955,get_started=0,get_sum=13336955,max=104857600,put=67955,put_sum=13336955,take=0,take_sum=0,val=0,wait.avgcount=0,wait.avgtime=0,wait.sum=0 1539079060000000000
> ceph,collection=objecter,host=ceph-rgw1003,id=datapro1-client.rgw.ceph-rgw1003.1668.94172644032512,type=rgw command_active=0,command_resend=0,command_send=0,linger_active=9,linger_ping=133974,linger_resend=0,linger_send=39671,map_epoch=0,map_full=0,map_inc=0,omap_del=0,omap_rd=7,omap_wr=35,op=2043314,op_active=0,op_laggy=0,op_pg=0,op_r=1071920,op_reply=2043314,op_resend=0,op_rmw=0,op_send=2043314,op_send_bytes=17651693250,op_w=971394,osd_laggy=7,osd_session_close=1538,osd_session_open=1598,osd_sessions=60,osdop_append=0,osdop_call=2257546,osdop_clonerange=0,osdop_cmpxattr=52503,osdop_create=298555,osdop_delete=59364,osdop_getxattr=0,osdop_mapext=0,osdop_notify=39662,osdop_other=443501,osdop_pgls=0,osdop_pgls_filter=0,osdop_read=206401,osdop_resetxattrs=0,osdop_rmxattr=4,osdop_setxattr=1599356,osdop_sparse_read=0,osdop_src_cmpxattr=0,osdop_stat=403066,osdop_tmap_get=0,osdop_tmap_put=0,osdop_tmap_up=0,osdop_truncate=0,osdop_watch=133983,osdop_write=0,osdop_writefull=280092,osdop_writesame=0,osdop_zero=0,poolop_active=0,poolop_resend=0,poolop_send=0,poolstat_active=0,poolstat_resend=0,poolstat_send=0,statfs_active=0,statfs_resend=0,statfs_send=0 1539079060000000000
</pre>
