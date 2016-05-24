# Ceph Storage Input Plugin

Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster.  

The plugin works by scanning the configured SocketDir for OSD and MON socket files.  When it finds
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


### Configuration:

```
# Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster.
[[inputs.ceph]]
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
```

### Measurements & Fields:

All fields are collected under the **ceph** measurement and stored as float64s. For a full list of fields, see the sample perf dumps in ceph_test.go. 


### Tags:

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
 

### Example Output:

<pre>
telegraf -test -config /etc/telegraf/telegraf.conf -config-directory /etc/telegraf/telegraf.d  -input-filter ceph
* Plugin: ceph, Collection 1
> ceph,collection=paxos, id=node-2,role=openstack,type=mon accept_timeout=0,begin=14931264,begin_bytes.avgcount=14931264,begin_bytes.sum=180309683362,begin_keys.avgcount=0,begin_keys.sum=0,begin_latency.avgcount=14931264,begin_latency.sum=9293.29589,collect=1,collect_bytes.avgcount=1,collect_bytes.sum=24,collect_keys.avgcount=1,collect_keys.sum=1,collect_latency.avgcount=1,collect_latency.sum=0.00028,collect_timeout=0,collect_uncommitted=0,commit=14931264,commit_bytes.avgcount=0,commit_bytes.sum=0,commit_keys.avgcount=0,commit_keys.sum=0,commit_latency.avgcount=0,commit_latency.sum=0,lease_ack_timeout=0,lease_timeout=0,new_pn=0,new_pn_latency.avgcount=0,new_pn_latency.sum=0,refresh=14931264,refresh_latency.avgcount=14931264,refresh_latency.sum=8706.98498,restart=4,share_state=0,share_state_bytes.avgcount=0,share_state_bytes.sum=0,share_state_keys.avgcount=0,share_state_keys.sum=0,start_leader=0,start_peon=1,store_state=14931264,store_state_bytes.avgcount=14931264,store_state_bytes.sum=353119959211,store_state_keys.avgcount=14931264,store_state_keys.sum=289807523,store_state_latency.avgcount=14931264,store_state_latency.sum=10952.835724 1462821234814535148
> ceph,collection=throttle-mon_client_bytes,id=node-2,type=mon get=1413017,get_or_fail_fail=0,get_or_fail_success=0,get_sum=71211705,max=104857600,put=1413013,put_sum=71211459,take=0,take_sum=0,val=246,wait.avgcount=0,wait.sum=0 1462821234814737219
> ceph,collection=throttle-mon_daemon_bytes,id=node-2,type=mon get=4058121,get_or_fail_fail=0,get_or_fail_success=0,get_sum=6027348117,max=419430400,put=4058121,put_sum=6027348117,take=0,take_sum=0,val=0,wait.avgcount=0,wait.sum=0 1462821234814815661
> ceph,collection=throttle-msgr_dispatch_throttler-mon,id=node-2,type=mon get=54276277,get_or_fail_fail=0,get_or_fail_success=0,get_sum=370232877040,max=104857600,put=54276277,put_sum=370232877040,take=0,take_sum=0,val=0,wait.avgcount=0,wait.sum=0 1462821234814872064
</pre>
