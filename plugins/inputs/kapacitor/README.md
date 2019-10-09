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

---

### kapacitor
The `kapacitor` measurement stores fields with information about [Kapacitor
tasks](https://docs.influxdata.com/kapacitor/latest/introduction/getting-started/#kapacitor-tasks)
and [subscriptions](https://docs.influxdata.com/kapacitor/latest/administration/subscription-management/).

- [num_enabled_tasks](#num_enabled_tasks)
- [num_subscriptions](#num_subscriptions)
- [num_tasks](#num_tasks)

##### num_enabled_tasks
The number of enabled Kapacitor tasks.  
**Data type:** integer

##### num_subscriptions
The number of Kapacitor/InfluxDB subscriptions.  
**Data type:** integer

##### num_tasks
The total number of Kapacitor tasks.  
**Data type:** integer

---

### kapacitor_edges
The `kapacitor_edges` measurement stores fields with information related to
[edges](https://docs.influxdata.com/kapacitor/v1.5/tick/introduction/#pipelines)
in Kapacitor TICKscripts.

- [collected](#collected)
- [emitted](#emitted)

##### collected
The number of messages collected by TICKscript edges.  
**Data type:** integer

##### emitted
The number of messages emitted by TICKscript edges.  
**Data type:** integer

---

### kapacitor_ingress
The `kapacitor_ingress` measurement stores fields with information about data
coming into Kapacitor.

- [points_received](#points_received)

##### points_received
The number of points received by Kapacitor.  
**Data type:** integer

---

### kapacitor_memstats
The `kapacitor_memstats` measurement stores fields related to Kapacitor memory usage.

- [alloc_bytes](#alloc_bytes)
- [buck_hash_sys_bytes](#buck_hash_sys_bytes)
- [frees](#frees)
- [gc_cpu_fraction](#gc_cpu_fraction)
- [gc_sys_bytes](#gc_sys_bytes)
- [heap_alloc_bytes](#heap_alloc_bytes)
- [heap_idle_bytes](#heap_idle_bytes)
- [heap_inuse_bytes](#heap_inuse_bytes)
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

##### alloc_bytes
The number of bytes of memory allocated by Kapacitor that are still in use.  
**Data type:** integer

##### buck_hash_sys_bytes
The number of bytes of memory used by the profiling bucket hash table.  
**Data type:** integer

##### frees
The number of heap objects freed.  
**Data type:** integer

##### gc_cpu_fraction
The fraction of Kapacitor's available CPU time used by garbage collection since
Kapacitor started.  
**Data type:** float

##### gc_sys_bytes
The number of bytes of memory used for garbage collection system metadata.  
**Data type:** integer

##### heap_alloc_bytes
The number of reachable and unreachable heap objects garbage collection has
not freed.  
**Data type:** integer

##### heap_idle_bytes
The number of heap bytes waiting to be used.  
**Data type:** integer

##### heap_inuse_bytes
The number of heap bytes in use.  
**Data type:** integer

##### heap_objects
The number of allocated objects.  
**Data type:** integer

##### heap_released_bytes
The number of heap bytes released to the operating system.  
**Data type:** integer

##### heap_sys_bytes
The number of heap bytes obtained from `system`.  
**Data type:** integer

##### last_gc_ns
The nanosecond epoch time of the last garbage collection.
**Data type:** integer

##### lookups
The total number of pointer lookups.  
**Data type:** integer

##### mallocs
The total number of mallocs.
**Data type:** integer

##### mcache_in_use_bytes
The number of bytes in use by mcache structures.  
**Data type:** integer

##### mcache_sys_bytes
The number of bytes used for mcache structures obtained from `system`.  
**Data type:** integer

##### mspan_in_use_bytes
The number of bytes in use by mspan structures.  
**Data type:** integer

##### mspan_sys_bytes
The number of bytes used for mspan structures obtained from `system`.  
**Data type:** integer

##### next_gc_ns
The nanosecond epoch time of the next garbage collection.  
**Data type:** integer

##### num_gc
The number of completed garbage collection cycles.  
**Data type:** integer

##### other_sys_bytes
The number of bytes used for other system allocations.  
**Data type:** integer

##### pause_total_ns
The totoal number of nanoseconds spent in garbage collection "stop-the-world"
pauses since Kapacitor started.  
**Data type:** integer

##### stack_in_use_bytes
The number of bytes in use by the stack allocator.  
**Data type:** integer

##### stack_sys_bytes
The number of bytes obtained from `system` for stack allocator.  
**Data type:** integer

##### sys_bytes
The number of bytes of memory obtained from `system`.  
**Data type:** integer

##### total_alloc_bytes
The total number of bytes allocated, even if freed.  
**Data type:** integer

---

### kapacitor_nodes
The `kapacitor_nodes` measurement stores fields related to events that occur in
[TICKscript nodes](https://docs.influxdata.com/kapacitor/latest/nodes/).

- [alerts_triggered](#alerts_triggered)
- [avg_exec_time_ns](#avg_exec_time_ns)
- [batches_queried](#batches_queried)
- [crits_triggered](#crits_triggered)
- [cooldown_drops](#cooldown_drops)
- [decrease_events](#decrease_events)
- [eval_errors](#eval_errors)
- [fields_defaulted](#fields_defaulted)
- [fields_deleted](#fields_deleted)
- [infos_triggered](#infos_triggered)
- [increase_events](#increase_events)
- [oks_triggered](#oks_triggered)
- [points_queried](#points_queried)
- [points_written](#points_written)
- [query_errors](#query_errors)
- [tags_defaulted](#tags_defaulted)
- [tags_deleted](#tags_deleted)
- [warns_triggered](#warns_triggered)
- [write_errors](#write_errors)

##### alerts_triggered
The total number of alerts triggered by TICKscripts.  
**Data type:** integer

##### avg_exec_time_ns
The average execution time of TICKscripts in nanoseconds.  
**Data type:** integer

##### batches_queried
The number of batches returned from queries.  
**Data type:** integer

##### crits_triggered
The number of critical (`crit`) alerts triggered by TICKscripts.  
**Data type:** integer

##### cooldown_drops
The number of times an autoscale event was dropped because of a cooldown timer.  
**Data type:** integer

##### decrease_events
The number of times an "autoscale" node decreased the replica count.  
**Data type:** integer

##### eval_errors
The number of evaluation errors caused by
[`eval` nodes](https://docs.influxdata.com/kapacitor/latest/nodes/eval_node/)
expressions in TICKscripts.  
**Data type:** integer

##### fields_defaulted
The number of missing fields.  
**Data type:** integer

##### fields_deleted
The number of fields deleted from points.
Kapacitor only increments the delete count if the field already existed.  
**Data type:** integer

##### infos_triggered
The number of info (`info`) alerts triggered by TICKscripts.  
**Data type:** integer

##### increase_events
The number of times an "autoscale" node increased the replica count.  
**Data type:** integer

##### oks_triggered
The number of ok (`ok`) alerts triggered by TICKscripts.  
**Data type:** integer

##### points_queried
The total number of points in batches.  
**Data type:** integer

##### points_written
The number of points written to InfluxDB or back to Kapacitor.
**Data type:** integer

##### query_errors
The number of errors when querying.
**Data type:** integer

##### tags_defaulted
The number of missing tags.  
**Data type:** integer

##### tags_deleted
The number of tags deleted from points.
Kapacitor only increments the delete count if the tag already existed.  
**Data type:** integer

##### warns_triggered
The number of warning (`warn`) alerts triggered by TICKscripts.  
**Data type:** integer

##### write_errors
The number of errors that occurred when writing to InfluxDB or to a given Kafka
topic and cluster.  
**Data type:** integer

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
