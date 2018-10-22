# Logstash Input Plugin

This plugin reads metrics exposed by
[Logstash Monitoring API](https://www.elastic.co/guide/en/logstash/current/monitoring-logstash.html).

### Configuration:

```toml
  ## This plugin reads metrics exposed by Logstash Monitoring API.
  ## https://www.elastic.co/guide/en/logstash/current/monitoring.html

  ## The URL of the exposed Logstash API endpoint
  url = "http://127.0.0.1:9600"

  ## Enable Logstash 6+ multi-pipeline statistics support
  multi_pipeline = true

  ## Should the general process statistics be gathered
  collect_process_stats = true

  ## Should the JVM specific statistics be gathered
  collect_jvm_stats = true

  ## Should the event pipelines statistics be gathered
  collect_pipelines_stats = true

  ## Should the plugin statistics be gathered
  collect_plugins_stats = true

  ## Should the queue statistics be gathered
  collect_queue_stats = true

  ## HTTP method
  # method = "GET"

  ## Optional HTTP headers
  # headers = {"X-Special-Header" = "Special-Value"}

  ## Override HTTP "Host" header
  # host_header = "logstash.example.com"

  ## Timeout for HTTP requests
  timeout = "5s"

  ## Optional HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Measurements & Fields:

- **logstash_jvm**
  * Fields:
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
  * Tags:
    - node_id
    - node_name
    - node_host
  	- node_version

- **logstash_process**
  * Fields:
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
  * Tags:
    - node_id
    - node_name
    - node_host
  	- node_version

- **logstash_events**
  * Fields:
    - queue_push_duration_in_millis
    - duration_in_millis
    - in
    - filtered
    - out
  * Tags:
    - node_id
    - node_name
    - node_host
  	- node_version
  	- pipeline (for Logstash 6 only)

- **logstash_plugins**
  * Fields:
    - queue_push_duration_in_millis (for input plugins only)
    - duration_in_millis
    - in
    - out
  * Tags:
    - node_id
    - node_name
    - node_host
  	- node_version
  	- pipeline (for Logstash 6 only)
  	- plugin_id
  	- plugin_name
  	- plugin_type

- **logstash_queue**
  * Fields:
    - events
    - free_space_in_bytes
    - max_queue_size_in_bytes
    - max_unread_events
    - page_capacity_in_bytes
    - queue_size_in_bytes
  * Tags:
    - node_id
    - node_name
    - node_host
    - node_version
    - pipeline (for Logstash 6 only)
    - queue_type  

### Tags description

- node_id - The uuid of the logstash node. Randomly generated.
- node_name - The name of the logstash node. Can be defined in the *logstash.yml* or defaults to the hostname.
  Can be used to break apart metrics from different logstash instances of the same host.
- node_host - The hostname of the logstash node.
  Can be different from the telegraf's host if a remote connection to logstash instance is used.
- node_version - The version of logstash service running on this node.
- pipeline (for Logstash 6 only) - The name of a pipeline if multi-pipeline is configured.
  Will defaults to "main" if there is only one pipeline and will be missing for logstash 5.
- plugin_id - The unique id of this plugin.
  It will be a randomly generated string unless it's defined in the logstash pipeline config file.
- plugin_name - The name of this plugin. i.e. file, elasticsearch, date, mangle. 
- plugin_type - The type of this plugin i.e. input/filter/output.
- queue_type - The type of the event queue (memory/persisted).

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter logstash -test

> logstash_jvm,host=node-6,node_host=node-6,node_id=3044f675-21ce-4335-898a-8408aa678245,node_name=node-6-test,node_version=6.4.2
  gc_collectors_old_collection_count=5,gc_collectors_old_collection_time_in_millis=702,gc_collectors_young_collection_count=95,gc_collectors_young_collection_time_in_millis=4772,mem_heap_committed_in_bytes=360804352,mem_heap_max_in_bytes=8389328896,mem_heap_used_in_bytes=252629768,mem_heap_used_percent=3,mem_non_heap_committed_in_bytes=212144128,mem_non_heap_used_in_bytes=188726024,mem_pools_old_committed_in_bytes=280260608,mem_pools_old_max_in_bytes=6583418880,mem_pools_old_peak_max_in_bytes=6583418880,mem_pools_old_peak_used_in_bytes=235352976,mem_pools_old_used_in_bytes=194754608,mem_pools_survivor_committed_in_bytes=8912896,mem_pools_survivor_max_in_bytes=200605696,mem_pools_survivor_peak_max_in_bytes=200605696,mem_pools_survivor_peak_used_in_bytes=8912896,mem_pools_survivor_used_in_bytes=4476680,mem_pools_young_committed_in_bytes=71630848,mem_pools_young_max_in_bytes=1605304320,mem_pools_young_peak_max_in_bytes=1605304320,mem_pools_young_peak_used_in_bytes=71630848,mem_pools_young_used_in_bytes=53398480,threads_count=60,threads_peak_count=62,uptime_in_millis=10469014 1540289864000000000
> logstash_process,host=node-6,node_host=node-6,node_id=3044f675-21ce-4335-898a-8408aa678245,node_name=node-6-test,node_version=6.4.2
  cpu_load_average_15m=39.84,cpu_load_average_1m=32.87,cpu_load_average_5m=39.23,cpu_percent=0,cpu_total_in_millis=389920,max_file_descriptors=262144,mem_total_virtual_in_bytes=17921355776,open_file_descriptors=132,peak_open_file_descriptors=140 1540289864000000000
> logstash_events,host=node-6,node_host=node-6,node_id=3044f675-21ce-4335-898a-8408aa678245,node_name=node-6-test,node_version=6.4.2,pipeline=main
  duration_in_millis=175144,filtered=4543,in=4543,out=4543,queue_push_duration_in_millis=19 1540289864000000000
> logstash_plugins,host=node-6,node_host=node-6,node_id=3044f675-21ce-4335-898a-8408aa678245,node_name=node-6-test,node_version=6.4.2,pipeline=main,plugin_id=input-kafka,plugin_name=kafka,plugin_type=input
  duration_in_millis=0,in=0,out=4543,queue_push_duration_in_millis=19 1540289864000000000
```
