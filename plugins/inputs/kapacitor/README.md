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

  ## Time limit for http requests
  timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Measurements & Fields

- kapacitor
    - num_enabled_tasks, integer
    - num_subscriptions, integer
    - num_tasks, integer
- kapacitor_edges
    - collected, integer
    - emitted, integer
- kapacitor_ingress
    - points_received, integer
- kapacitor_memstats
    - alloc_bytes, integer
    - buck_hash_sys_bytes, integer
    - frees, integer
    - gcc_pu_fraction, float
    - gc_sys_bytes, integer
    - heap_alloc_bytes, integer
    - heap_idle_bytes, integer
    - heap_inuse_bytes, integer
    - heap_objects, integer
    - heap_released_bytes, integer
    - heap_sys_bytes, integer
    - last_gc_ns, integer
    - lookups, integer
    - mallocs, integer
    - mcache_in_use_bytes, integer
    - mcache_sys_bytes, integer
    - mspan_in_use_bytes, integer
    - mspan_sys_bytes, integer
    - next_gc_ns, integer
    - num_gc, integer
    - other_sys_bytes, integer
    - pause_total_ns, integer
    - stack_in_use_bytes, integer
    - stack_sys_bytes, integer
    - sys_bytes, integer
    - total_alloc_bytes, integer
- kapacitor_nodes
    - alerts_triggered, integer
    - avg_exec_time_ns, integer
    - batches_queried, integer
    - crits_triggered, integer
    - eval_errors, integer
    - fields_defaulted, integer
    - infos_triggered, integer
    - oks_triggered, integer
    - points_queried, integer
    - points_written, integer
    - query_errors, integer
    - tags_defaulted, integer
    - warns_triggered, integer
    - write_errors, integer

*Note:* The Kapacitor variables `host`, `cluster_id`, and `server_id`
are currently not recorded due to the potential high cardinality of
these values.

### Example Output:

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
