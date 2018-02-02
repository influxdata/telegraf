# Filebeat Plugin

The Filebeat plugin will collect metrics from the given Filebeat instances.

### Configuration:

```toml
[[inputs.filebeat]]
  ## Multiple URLs from which to read Filebeat-formatted JSON
  ## Default is "http://localhost:9602/debug/vars".
  urls = [
    "http://localhost:9602/debug/vars"
  ]
  ## Time limit for http requests
  timeout = "5s"
```

### Measurements & Fields

- filebeat
    - publish_events, integer
    - registrar_states_cleanup, integer
    - registrar_states_current, integer
    - registrar_states_update, integer
    - registrar_writes, integer
    - harvester_closed, integer
    - harvester_files_truncated, integer
    - harvester_open_files, integer
    - harvester_running, integer
    - harvester_skipped, integer
    - harvester_started, integer
    - prospector_log_files_renamed, integer
    - prospector_log_files_truncated, integer
- filebeat_memstats
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
- libbeat
    - config_module_running, integer
    - config_module_starts, integer
    - config_module_stops, integer
    - config_reloads, integer
    - es_call_count_publish_events, integer
    - es_publish_read_bytes, integer
    - es_publish_read_errors, integer
    - es_publish_write_bytes, integer
    - es_publish_write_errors, integer
    - es_published_and_acked_events, integer
    - es_published_but_not_acked_events, integer
    - kafka_call_count_publishevents, integer
    - kafka_published_and_acked_events, integer
    - kafka_published_but_not_acked_events, integer
    - logstash_call_count_publishevents, integer
    - logstash_publish_read_bytes, integer
    - logstash_publish_read_errors, integer
    - logstash_publish_write_bytes, integer
    - logstash_publish_write_errors, integer
    - logstash_published_and_acked_events, integer
    - logstash_published_but_not_acked_events, integer
    - outputs_messages_dropped, integer
    - publisher_messages_in_worker_queues, integer
    - publisher_published_events, integer
    - redis_publish_read_bytes, integer
    - redis_publish_read_errors, integer
    - redis_publish_write_bytes, integer
    - redis_publish_write_errors, integer


### Example Output:

```
$ /opt/OSAGtelegraf/bin/telegraf -config /etc/filebeat.conf --test
* Plugin: inputs.filebeat, Collection 1
> filebeat_memstats,url=http://localhost:9602/debug/vars,host=tau heap_sys_bytes=5636096i,mcache_in_use_bytes=4800i,stack_in_use_bytes=655360i,last_gc_ns=1505118613012035994i,pause_total_ns=3820190595i,total_alloc_bytes=1500537416i,buck_hash_sys_bytes=1473183i,heap_alloc_bytes=1030208i,heap_idle_bytes=3678208i,heap_released_bytes=2990080i,mallocs=4345197i,stack_sys_bytes=655360i,frees=4338330i,lookups=139i,mspan_in_use_bytes=33600i,mcache_sys_bytes=16384i,num_gc=2052i,sys_bytes=9509112i,gc_sys_bytes=419840i,heap_objects=6867i,next_gc_ns=4194304i,other_sys_bytes=1226329i,alloc_bytes=1030208i,gcc_pu_fraction=0.000020812596115449476,heap_in_use_bytes=1957888i,mspan_sys_bytes=81920i 1505118691000000000
> filebeat,host=tau,url=http://localhost:9602/debug/vars registrar_states_cleanup=4i,prospector_log_files_renamed=0i,registrar_states_current=1i,harvester_open_files=0i,prospector_log_files_truncated=0i,publish_events=13i,registrar_states_update=13i,harvester_closed=4i,harvester_files_truncated=0i,harvester_skipped=0i,registrar_writes=11i,harvester_running=0i,harvester_started=4i 1505118691000000000
> libbeat,url=http://localhost:9602/debug/vars,host=tau logstash_published_and_acked_events=0i,publisher_messages_in_worker_queues=0i,redis_publish_read_bytes=0i,config_module_running=0i,es_call_count_publish_events=0i,es_published_but_not_acked_events=0i,logstash_publish_write_errors=0i,es_published_and_acked_events=0i,logstash_publish_read_errors=0i,config_reloads=0i,logstash_call_count_publishevents=0i,logstash_published_but_not_acked_events=0i,outputs_messages_dropped=0i,kafka_published_and_acked_events=0i,logstash_publish_write_bytes=0i,redis_publish_write_bytes=0i,redis_publish_write_errors=0i,config_module_starts=0i,kafka_call_count_publishevents=0i,redis_publish_read_errors=0i,es_publish_write_bytes=0i,es_publish_write_errors=0i,logstash_publish_read_bytes=0i,publisher_published_events=0i,config_module_stops=0i,es_publish_read_bytes=0i,kafka_published_but_not_acked_events=0i,es_publish_read_errors=0i 1505118691000000000
```
