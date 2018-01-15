# Logstash Input Plugin

This plugin reads metrics exposed by [Logstash Monitoring API](https://www.elastic.co/guide/en/logstash/current/monitoring-logstash.html).

### Configuration:

```toml
  ## This plugin reads metrics exposed by Logstash Monitoring API.
  # https://www.elastic.co/guide/en/logstash/current/monitoring.html
  #
  # url = "http://localhost:9600"
  url = "http://localhost:9600"
```

### Measurements & Fields:

- logstash_jvm
    - threads_peak_count
    - mem_pools_survivor_peak_max_in_bytes
    - mem_pools_survivor_max_in_bytes
    - mem_pools_old_peak_used_in_bytes
    - mem_pools_young_used_in_bytes
    - mem_non_heap_committed_in_bytes
    - threads_count
    - mem_pools_old_committed_in_bytes
    - mem_pools_young_peak_max_in_bytes
    - mem_heap_used_percent
    - gc_collectors_young_collection_time_in_millis
    - mem_pools_survivor_peak_used_in_bytes
    - mem_pools_young_committed_in_bytes
    - gc_collectors_old_collection_time_in_millis
    - gc_collectors_old_collection_count
    - mem_pools_survivor_used_in_bytes
    - mem_pools_old_used_in_bytes
    - mem_pools_young_max_in_bytes
    - mem_heap_max_in_bytes
    - mem_non_heap_used_in_bytes
    - mem_pools_survivor_committed_in_bytes
    - mem_pools_old_max_in_bytes
    - mem_heap_committed_in_bytes
    - mem_pools_old_peak_max_in_bytes
    - mem_pools_young_peak_used_in_bytes
    - mem_heap_used_in_bytes
    - gc_collectors_young_collection_count
    - uptime_in_millis

- logstash_process
    - open_file_descriptors
    - cpu_load_average_1m
    - cpu_load_average_5m
    - cpu_load_average_15m
    - cpu_total_in_millis
    - cpu_percent
    - peak_open_file_descriptors
    - max_file_descriptors
    - mem_total_virtual_in_bytes
    - mem_total_virtual_in_bytes

- logstash_events
    - queue_push_duration_in_millis
    - duration_in_millis
    - in
    - filtered
    - out

- logstash_plugins
  There are 3 categories, input, filter, output. For each will be separate measurement consisting:
      - tags
        - type (input|filter|output)
        - plugin (name of the plugin, eg:beats, stdout)
      - fields
        - queue_push_duration_in_millis (for input plugins only)
        - duration_in_millis
        - in
        - out

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter logstash -test
> logstash_jvm,node_id=04f508ba-8ad5-466b-9b23-02ec71cba42e,host=laptop mem_non_heap_committed_in_bytes=99561472,mem_pools_young_max_in_bytes=279183360,mem_pools_old_committed_in_bytes=724828160,mem_heap_used_percent=21,mem_pools_survivor_peak_used_in_bytes=34865152,mem_pools_old_used_in_bytes=37868216,gc_collectors_old_collection_time_in_millis=568,threads_peak_count=33,mem_pools_young_peak_used_in_bytes=279183360,mem_heap_committed_in_bytes=1038876672,mem_non_heap_used_in_bytes=92957192,mem_pools_survivor_used_in_bytes=34865152,mem_pools_young_committed_in_bytes=279183360,gc_collectors_young_collection_time_in_millis=1605,mem_heap_max_in_bytes=1038876672,mem_heap_used_in_bytes=218198056,mem_pools_old_max_in_bytes=724828160,gc_collectors_young_collection_count=5,mem_pools_old_peak_max_in_bytes=724828160,mem_pools_young_used_in_bytes=145464688,mem_pools_survivor_committed_in_bytes=34865152,mem_pools_old_peak_used_in_bytes=85451288,mem_pools_young_peak_max_in_bytes=279183360,gc_collectors_old_collection_count=2,uptime_in_millis=119806,threads_count=33,mem_pools_survivor_peak_max_in_bytes=34865152,mem_pools_survivor_max_in_bytes=34865152 1512382629000000000
> logstash_process,node_id=04f508ba-8ad5-466b-9b23-02ec71cba42e,host=laptop cpu_load_average_5m=1.05,open_file_descriptors=110,mem_total_virtual_in_bytes=4857896960,cpu_percent=5,cpu_load_average_1m=1.01,peak_open_file_descriptors=110,max_file_descriptors=1048576,cpu_total_in_millis=97420,cpu_load_average_15m=0.91 1512382629000000000
> logstash_events,node_id=04f508ba-8ad5-466b-9b23-02ec71cba42e,host=laptop duration_in_millis=2,in=3,out=3,filtered=0,queue_push_duration_in_millis=1 1512382629000000000
> logstash_plugins,type=input,plugin=beats,host=laptop queue_push_duration_in_millis=2,duration_in_millis=0,in=5,out=5 1512382629000000000
> logstash_plugins,host=laptop,type=output,plugin=stdout in=6,out=6,duration_in_millis=3 1512382629000000000
```
