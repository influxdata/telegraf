# Kapacitor Plugin

The Kapacitor plugin will collect metrics from the given Kapacitor instances.

## Configuration:

```toml
[[inputs.kapacitor]]
  ## Multiple URLs from which to read Kapacitor-formatted JSON
  ## Default is "http://localhost:9092/kapacitor/v1/debug/vars".
  urls = [
    "http://localhost:9092/kapacitor/v1/debug/vars"
  ]

  ## Time limit for http requests
  timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Measurements & Fields

- [kapacitor](#kapacitor)
    - [num_enabled_tasks](#num_enabled_tasks)
    - [num_subscriptions](#num_subscriptions)
    - [num_tasks](#num_tasks)
- [kapacitor_edges](#kapacitor_edges)
    - [collected](#collected)
    - [emitted](#emitted)
- [kapacitor_ingress](#kapacitor_ingress)
    - [points_received](#points_received)
- [kapacitor_load](#kapacitor_load)
    - [errors](#errors)
- [kapacitor_memstats](#kapacitor_memstats)
    - [alloc_bytes](#alloc_bytes)
    - [buck_hash_sys_bytes](#buck_hash_sys_bytes)
    - [frees](#frees)
    - [gc_sys_bytes](#gc_sys_bytes)
    - [gcc_pu_fraction](#gcc_pu_fraction)
    - [heap_alloc_bytes](#heap_alloc_bytes)
    - [heap_idle_bytes](#heap_idle_bytes)
    - [heap_in_use_bytes](#heap_in_use_bytes)
    - [heap_objects](#heap_objects)
    - [heap_released_bytes](#heap_released_bytes)
    - [heap_sys_bytes](#heap_sys_bytes)
    - [last_gc_ns](#last_gc_ns)
    - [lookups](#lookups)
    - [mallocs](#mallocs)
    - [mcache_in_use_bytes](#mcache_in_use_bytes)
    - [mcache_sys_bytes](#mcache_sys_bytes)
    - [mspan_in_use_bytes](#mspan_in_use_bytes)
    - [mspan_sys_bytes](#mspan_sys_bytes)
    - [next_gc_ns](#next_gc_ns)
    - [num_gc](#num_gc)
    - [other_sys_bytes](#other_sys_bytes)
    - [pause_total_ns](#pause_total_ns)
    - [stack_in_use_bytes](#stack_in_use_bytes)
    - [stack_sys_bytes](#stack_sys_bytes)
    - [sys_bytes](#sys_bytes)
    - [total_alloc_bytes](#total_alloc_bytes)
- [kapacitor_nodes](#kapacitor_nodes)
    - [alerts_inhibited](#alerts_inhibited)
    - [alerts_triggered](#alerts_triggered)
    - [avg_exec_time_ns](#avg_exec_time_ns)
    - [crits_triggered](#crits_triggered)
    - [errors](#errors)
    - [infos_triggered](#infos_triggered)
    - [oks_triggered](#oks_triggered)
    - [points_written](#points_written)
    - [warns_triggered](#warns_triggered)
    - [write_errors](#write_errors)
- [kapacitor_topics](#kapacitor_topics)
    - [collected](#collected)


---

### kapacitor
The `kapacitor` measurement stores fields with information related to
[Kapacitor tasks](https://docs.influxdata.com/kapacitor/latest/introduction/getting-started/#kapacitor-tasks)
and [subscriptions](https://docs.influxdata.com/kapacitor/latest/administration/subscription-management/).

#### num_enabled_tasks
The number of enabled Kapacitor tasks.  
_**Data type:** integer_

#### num_subscriptions
The number of Kapacitor/InfluxDB subscriptions.  
_**Data type:** integer_

#### num_tasks
The total number of Kapacitor tasks.  
_**Data type:** integer_

---

### kapacitor_edges
The `kapacitor_edges` measurement stores fields with information related to
[edges](https://docs.influxdata.com/kapacitor/latest/tick/introduction/#pipelines)
in Kapacitor TICKscripts.

#### collected
The number of messages collected by TICKscript edges.  
_**Data type:** integer_

#### emitted
The number of messages emitted by TICKscript edges.  
_**Data type:** integer_

---

### kapacitor_ingress
The `kapacitor_ingress` measurement stores fields with information related to data
coming into Kapacitor.

#### points_received
The number of points received by Kapacitor.  
_**Data type:** integer_

---

### kapacitor_load
The `kapacitor_load` measurement stores fields with information related to the
[Kapacitor Load Directory service](https://docs.influxdata.com/kapacitor/latest/guides/load_directory/).

#### errors
The number of errors reported from the load directory service.  
_**Data type:** integer_

---

### kapacitor_memstats
The `kapacitor_memstats` measurement stores fields related to Kapacitor memory usage.

#### alloc_bytes
The number of bytes of memory allocated by Kapacitor that are still in use.  
_**Data type:** integer_

#### buck_hash_sys_bytes
The number of bytes of memory used by the profiling bucket hash table.  
_**Data type:** integer_

#### frees
The number of heap objects freed.  
_**Data type:** integer_

#### gc_sys_bytes
The number of bytes of memory used for garbage collection system metadata.  
_**Data type:** integer_

#### gcc_pu_fraction
The fraction of Kapacitor's available CPU time used by garbage collection since
Kapacitor started.  
_**Data type:** float_

#### heap_alloc_bytes
The number of reachable and unreachable heap objects garbage collection has
not freed.  
_**Data type:** integer_

#### heap_idle_bytes
The number of heap bytes waiting to be used.  
_**Data type:** integer_

#### heap_in_use_bytes
The number of heap bytes in use.  
_**Data type:** integer_

#### heap_objects
The number of allocated objects.  
_**Data type:** integer_

#### heap_released_bytes
The number of heap bytes released to the operating system.  
_**Data type:** integer_

#### heap_sys_bytes
The number of heap bytes obtained from `system` .    
_**Data type:** integer_

#### last_gc_ns
The nanosecond epoch time of the last garbage collection.  
_**Data type:** integer_

#### lookups
The total number of pointer lookups.  
_**Data type:** integer_

#### mallocs
The total number of mallocs.  
_**Data type:** integer_

#### mcache_in_use_bytes
The number of bytes in use by mcache structures.  
_**Data type:** integer_

#### mcache_sys_bytes
The number of bytes used for mcache structures obtained from `system` .  
_**Data type:** integer_

#### mspan_in_use_bytes
The number of bytes in use by mspan structures.  
_**Data type:** integer_

#### mspan_sys_bytes
The number of bytes used for mspan structures obtained from `system` .  
_**Data type:** integer_

#### next_gc_ns
The nanosecond epoch time of the next garbage collection.  
_**Data type:** integer_

#### num_gc
The number of completed garbage collection cycles.  
_**Data type:** integer_

#### other_sys_bytes
The number of bytes used for other system allocations.  
_**Data type:** integer_

#### pause_total_ns
The totoal number of nanoseconds spent in garbage collection "stop-the-world"
pauses since Kapacitor started.  
_**Data type:** integer_

#### stack_in_use_bytes
The number of bytes in use by the stack allocator.  
_**Data type:** integer_

#### stack_sys_bytes
The number of bytes obtained from `system` for stack allocator.  
_**Data type:** integer_

#### sys_bytes
The number of bytes of memory obtained from `system` .  
_**Data type:** integer_

#### total_alloc_bytes
The total number of bytes allocated, even if freed.  
_**Data type:** integer_

---

### kapacitor_nodes
The `kapacitor_nodes` measurement stores fields related to events that occur in
[TICKscript nodes](https://docs.influxdata.com/kapacitor/latest/nodes/).

#### alerts_inhibited
The total number of alerts inhibited by TICKscripts.  
_**Data type:** integer_

#### alerts_triggered
The total number of alerts triggered by TICKscripts.  
_**Data type:** integer_

#### avg_exec_time_ns
The average execution time of TICKscripts in nanoseconds.  
_**Data type:** integer_

#### crits_triggered
The number of critical (`crit`) alerts triggered by TICKscripts.  
_**Data type:** integer_

#### errors
The number of errors caused caused by TICKscripts.  
_**Data type:** integer_

#### infos_triggered
The number of info (`info`) alerts triggered by TICKscripts.  
_**Data type:** integer_

#### oks_triggered
The number of ok (`ok`) alerts triggered by TICKscripts.  
_**Data type:** integer_

#### points_written
The number of points written to InfluxDB or back to Kapacitor.  
_**Data type:** integer_

#### warns_triggered
The number of warning (`warn`) alerts triggered by TICKscripts.  
_**Data type:** integer_

#### working_cardinality
The total number of unique series processed.  
_**Data type:** integer_

#### write_errors
The number of errors that occurred when writing to InfluxDB or to a given Kafka
topic and cluster.  
_**Data type:** integer_

---

### kapacitor_topics
The `kapacitor_topics` measurement stores fields related to
Kapacitor topics](https://docs.influxdata.com/kapacitor/latest/working/using_alert_topics/).

#### collected
The number of events collected by Kapacitor topics.  
_**Data type:** integer_  

---

*Note:* The Kapacitor variables `host`, `cluster_id`, and `server_id`
are currently not recorded due to the potential high cardinality of
these values.

## Example Output

```
$ telegraf --config /etc/telegraf.conf --input-filter kapacitor --test
* Plugin: inputs.kapacitor, Collection 1
> kapacitor_memstats,host=hostname.local,kap_version=1.1.0~rc2,url=http://localhost:9092/kapacitor/v1/debug/vars alloc_bytes=6974808i,buck_hash_sys_bytes=1452609i,frees=207281i,gc_sys_bytes=802816i,gcc_pu_fraction=0.00004693548939673313,heap_alloc_bytes=6974808i,heap_idle_bytes=6742016i,heap_in_use_bytes=9183232i,heap_objects=23216i,heap_released_bytes=0i,heap_sys_bytes=15925248i,last_gc_ns=1478791460012676997i,lookups=88i,mallocs=230497i,mcache_in_use_bytes=9600i,mcache_sys_bytes=16384i,mspan_in_use_bytes=98560i,mspan_sys_bytes=131072i,next_gc_ns=11467528i,num_gc=8i,other_sys_bytes=2236087i,pause_total_ns=2994110i,stack_in_use_bytes=1900544i,stack_sys_bytes=1900544i,sys_bytes=22464760i,total_alloc_bytes=35023600i 1478791462000000000
> kapacitor,host=hostname.local,kap_version=1.1.0~rc2,url=http://localhost:9092/kapacitor/v1/debug/vars num_enabled_tasks=5i,num_subscriptions=5i,num_tasks=5i 1478791462000000000
> kapacitor_edges,child=stream0,host=hostname.local,parent=stream,task=deadman-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=shard,retention_policy=monitor,task_master=main points_received=120 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=subscriber,retention_policy=monitor,task_master=main points_received=60 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=http_out,node=http_out3,task=sys-stats,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_edges,child=window6,host=hostname.local,parent=derivative5,task=deadman-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=from,node=from1,task=sys-stats,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=stream,node=stream0,task=test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=window,node=window6,task=deadman-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=cq,retention_policy=monitor,task_master=main points_received=10 1478791462000000000
> kapacitor_edges,child=http_out3,host=hostname.local,parent=window2,task=sys-stats,type=batch collected=0,emitted=0 1478791462000000000
> kapacitor_edges,child=mean4,host=hostname.local,parent=log3,task=deadman-test,type=batch collected=0,emitted=0 1478791462000000000
> kapacitor_ingress,database=_kapacitor,host=hostname.local,measurement=nodes,retention_policy=autogen,task_master=main points_received=207 1478791462000000000
> kapacitor_edges,child=stream0,host=hostname.local,parent=stream,task=sys-stats,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_edges,child=log6,host=hostname.local,parent=sum5,task=derivative-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_edges,child=from1,host=hostname.local,parent=stream0,task=sys-stats,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=alert,node=alert2,task=test,type=stream alerts_triggered=0,avg_exec_time_ns=0i,crits_triggered=0,infos_triggered=0,oks_triggered=0,warns_triggered=0 1478791462000000000
> kapacitor_edges,child=log3,host=hostname.local,parent=derivative2,task=derivative-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_ingress,database=_kapacitor,host=hostname.local,measurement=runtime,retention_policy=autogen,task_master=main points_received=9 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=tsm1_filestore,retention_policy=monitor,task_master=main points_received=120 1478791462000000000
> kapacitor_edges,child=derivative2,host=hostname.local,parent=from1,task=derivative-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=stream,node=stream0,task=derivative-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=queryExecutor,retention_policy=monitor,task_master=main points_received=10 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=tsm1_wal,retention_policy=monitor,task_master=main points_received=120 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=log,node=log6,task=derivative-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_edges,child=stream,host=hostname.local,parent=stats,task=task_master:main,type=stream collected=598,emitted=598 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=write,retention_policy=monitor,task_master=main points_received=10 1478791462000000000
> kapacitor_edges,child=stream0,host=hostname.local,parent=stream,task=derivative-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=log,node=log3,task=deadman-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=from,node=from1,task=deadman-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_ingress,database=_kapacitor,host=hostname.local,measurement=ingress,retention_policy=autogen,task_master=main points_received=148 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=eval,node=eval4,task=derivative-test,type=stream avg_exec_time_ns=0i,eval_errors=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=derivative,node=derivative2,task=derivative-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=runtime,retention_policy=monitor,task_master=main points_received=10 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=httpd,retention_policy=monitor,task_master=main points_received=10 1478791462000000000
> kapacitor_edges,child=sum5,host=hostname.local,parent=eval4,task=derivative-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_ingress,database=_kapacitor,host=hostname.local,measurement=kapacitor,retention_policy=autogen,task_master=main points_received=9 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=from,node=from1,task=test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=tsm1_engine,retention_policy=monitor,task_master=main points_received=120 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=window,node=window2,task=deadman-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=stream,node=stream0,task=deadman-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_edges,child=influxdb_out4,host=hostname.local,parent=http_out3,task=sys-stats,type=batch collected=0,emitted=0 1478791462000000000
> kapacitor_edges,child=window2,host=hostname.local,parent=from1,task=deadman-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=from,node=from1,task=derivative-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_edges,child=from1,host=hostname.local,parent=stream0,task=deadman-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=database,retention_policy=monitor,task_master=main points_received=40 1478791462000000000
> kapacitor_edges,child=stream,host=hostname.local,parent=write_points,task=task_master:main,type=stream collected=750,emitted=750 1478791462000000000
> kapacitor_edges,child=log7,host=hostname.local,parent=window6,task=deadman-test,type=batch collected=0,emitted=0 1478791462000000000
> kapacitor_edges,child=window2,host=hostname.local,parent=from1,task=sys-stats,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=log,node=log7,task=deadman-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_ingress,database=_kapacitor,host=hostname.local,measurement=edges,retention_policy=autogen,task_master=main points_received=225 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=derivative,node=derivative5,task=deadman-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_edges,child=from1,host=hostname.local,parent=stream0,task=test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_edges,child=alert2,host=hostname.local,parent=from1,task=test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=log,node=log3,task=derivative-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=influxdb_out,node=influxdb_out4,task=sys-stats,type=stream avg_exec_time_ns=0i,points_written=0,write_errors=0 1478791462000000000
> kapacitor_edges,child=stream0,host=hostname.local,parent=stream,task=test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_edges,child=log3,host=hostname.local,parent=window2,task=deadman-test,type=batch collected=0,emitted=0 1478791462000000000
> kapacitor_edges,child=derivative5,host=hostname.local,parent=mean4,task=deadman-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=stream,node=stream0,task=sys-stats,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=window,node=window2,task=sys-stats,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=mean,node=mean4,task=deadman-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_edges,child=from1,host=hostname.local,parent=stream0,task=derivative-test,type=stream collected=0,emitted=0 1478791462000000000
> kapacitor_ingress,database=_internal,host=hostname.local,measurement=tsm1_cache,retention_policy=monitor,task_master=main points_received=120 1478791462000000000
> kapacitor_nodes,host=hostname.local,kind=sum,node=sum5,task=derivative-test,type=stream avg_exec_time_ns=0i 1478791462000000000
> kapacitor_edges,child=eval4,host=hostname.local,parent=log3,task=derivative-test,type=stream collected=0,emitted=0 1478791462000000000
```
