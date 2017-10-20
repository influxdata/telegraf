# Logstash Input Plugin

This plugin reads metrics exposed by [Logstash Monitoring API](https://www.elastic.co/guide/en/logstash/current/monitoring-logstash.html).

### Configuration:

```toml
  ## This plugin reads metrics exposed by Logstash Monitoring API.
  # https://www.elastic.co/guide/en/logstash/current/monitoring.html
  #
  # logstashURL = "http://localhost:9600"
  logstashURL = "http://localhost:9600"
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
    - duration_in_millis
    - in
    - filtered
    - out

- logstash_plugins
  There are 3 categories, input, filter, output. For each will be separate measurement with name and 3 fields.
  Example for stdout
    - logstash_plugin_output_stdout
      - name
      - duration_in_millis
      - in
      - out

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter logstash -test
logstash_jvm,host=ThinkPad-T440s threads_peak_count=26,mem_pools_survivor_peak_max_in_bytes=34865152,mem_pools_survivor_max_in_bytes=69730304,mem_pools_old_peak_used_in_bytes=124331072,mem_pools_young_used_in_bytes=79756008,mem_non_heap_committed_in_bytes=214155264,threads_count=25,mem_pools_old_committed_in_bytes=357957632,mem_pools_young_peak_max_in_bytes=279183360,mem_heap_used_percent=16,gc_collectors_young_collection_time_in_millis=1031,mem_pools_survivor_peak_used_in_bytes=8912896,mem_pools_young_committed_in_bytes=143261696,gc_collectors_old_collection_time_in_millis=114,gc_collectors_old_collection_count=2,mem_pools_survivor_used_in_bytes=9292032,mem_pools_old_used_in_bytes=248662144,mem_pools_young_max_in_bytes=558366720,mem_heap_max_in_bytes=2077753344,mem_non_heap_used_in_bytes=199046736,mem_pools_survivor_committed_in_bytes=17825792,mem_pools_old_max_in_bytes=1449656320,mem_heap_committed_in_bytes=519045120,mem_pools_old_peak_max_in_bytes=724828160,mem_pools_young_peak_used_in_bytes=71630848,mem_heap_used_in_bytes=337710184,gc_collectors_young_collection_count=55,uptime_in_millis=801834 1492681500000000000
logstash_process,host=ThinkPad-T440s open_file_descriptors=83,cpu_load_average_1m=0.86,cpu_load_average_5m=0.67,cpu_load_average_15m=0.49,cpu_total_in_millis=97500000000,cpu_percent=1,peak_open_file_descriptors=83,max_file_descriptors=1048576,mem_total_virtual_in_bytes=4788379648 1492681500000000000
logstash_events,host=ThinkPad-T440s duration_in_millis=1151,in=1269,filtered=1269,out=1269 1492681500000000000
logstash_plugin_output_s3,host=ThinkPad-T440s name="s3",duration_in_millis=228i,in=1269i,out=1269i 1492681500000000000
logstash_plugin_output_stdout,host=ThinkPad-T440s name="stdout",duration_in_millis=360i,in=1269i,out=1269i 1492681500000000000
```
