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
