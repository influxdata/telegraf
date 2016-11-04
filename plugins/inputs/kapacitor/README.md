# Kapacitor Plugin

The Kapacitor plugin will collect metrics from the given Kapacitor instances.

### Configuration:

```toml
[[inputs.kapacitor]]
  ## Multiple URLs from which to read Kapacitor-formatted JSON
  ## Default is "http://localhost:9092/kapacitor/v1/debug/vars".
  urls = [
    "http://localhost:9092/kapacitor/v1/debug/vars"
  ]
```

### Measurements & Fields

- kapacitor
    - num_enabled_tasks,integer
    - num_subscriptions,integer
    - num_tasks,integer
- kapacitor_edges
    - collected,float
    - emitted,float
- kapacitor_ingress
    - points_received,float
- kapacitor_memstats
    - alloc,integer
    - buck_hash_sys,integer
    - frees,integer
    - gc_sys,integer
    - gcc_pu_fraction,float
    - heap_alloc,integer
    - heap_idle,integer
    - heap_inuse,integer
    - heap_objects,integer
    - heap_released,integer
    - heap_sys,integer
    - last_gc,integer
    - lookups,integer
    - mallocs,integer
    - mcache_inuse,integer
    - mcache_sys,integer
    - mspan_inuse,integer
    - mspan_sys,integer
    - next_gc,integer
    - num_gc,integer
    - other_sys,integer
    - pause_total_ns,integer
    - stack_inuse,integer
    - stack_sys,integer
    - sys,integer
    - total_alloc, integer
- kapacitor_nodes
    - alerts_triggered,float
    - avg_exec_time_ns,integer
    - batches_queried,float
    - crits_triggered,float
    - eval_errors,float
    - fields_defaulted,float
    - infos_triggered,float
    - oks_triggered,float
    - points_queried,float
    - points_written,float
    - query_errors,float
    - tags_defaulted,float
    - warns_triggered,float
    - write_errors,float

*Note:* The Kapacitor variables `host`, `cluster_id`, and `server_id`
are currently not recorded due to the potential high cardinality of
these values.

### Example Output:

```
$ telegraf -config /etc/telegraf.conf -input-filter kapacitor -test
* Plugin: inputs.kapacitor, Collection 1
> kapacitor_memstats,host=hostname.local,kap_version=1.1.0~rc2,url=http://localhost:9092/kapacitor/v1/debug/vars alloc=6974808i,buck_hash_sys=1452609i,frees=207281i,gc_sys=802816i,gcc_pu_fraction=0.00004693548939673313,heap_alloc=6974808i,heap_idle=6742016i,heap_inuse=9183232i,heap_objects=23216i,heap_released=0i,heap_sys=15925248i,last_gc=1478791460012676997i,lookups=88i,mallocs=230497i,mcache_inuse=9600i,mcache_sys=16384i,mspan_inuse=98560i,mspan_sys=131072i,next_gc=11467528i,num_gc=8i,other_sys=2236087i,pause_total_ns=2994110i,stack_inuse=1900544i,stack_sys=1900544i,sys=22464760i,total_alloc=35023600i 1478791462000000000
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
