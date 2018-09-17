# Internal Input Plugin

The `internal` plugin collects metrics about the telegraf agent itself.

Note that some metrics are aggregates across all instances of one type of
plugin.

### Configuration:

```toml
# Collect statistics about itself
[[inputs.internal]]
  ## If true, collect telegraf memory stats.
  # collect_memstats = true
```

### Measurements & Fields:

memstats are taken from the Go runtime: https://golang.org/pkg/runtime/#MemStats

- internal\_memstats
    - alloc\_bytes
    - frees
    - heap\_alloc\_bytes
    - heap\_idle\_bytes
    - heap\_in\_use\_bytes
    - heap\_objects\_bytes
    - heap\_released\_bytes
    - heap\_sys\_bytes
    - mallocs
    - num\_gc
    - pointer\_lookups
    - sys\_bytes
    - total\_alloc\_bytes

agent stats collect aggregate stats on all telegraf plugins.

- internal\_agent
    - gather\_errors
    - metrics\_dropped
    - metrics\_gathered
    - metrics\_written

internal\_gather stats collect aggregate stats on all input plugins
that are of the same input type. They are tagged with `input=<plugin_name>`.

- internal\_gather
    - gather\_time\_ns
    - metrics\_gathered

internal\_write stats collect aggregate stats on all output plugins
that are of the same input type. They are tagged with `output=<plugin_name>`.


- internal\_write
    - buffer\_limit
    - buffer\_size
    - metrics\_written
    - metrics\_filtered
    - write\_time\_ns

internal\_\<plugin\_name\> are metrics which are defined on a per-plugin basis, and
usually contain tags which differentiate each instance of a particular type of
plugin.

- internal\_\<plugin\_name\>
    - individual plugin-specific fields, such as requests counts.

### Tags:

All measurements for specific plugins are tagged with information relevant
to each particular plugin.

### Example Output:

```
internal_memstats,host=tyrion alloc_bytes=4457408i,sys_bytes=10590456i,pointer_lookups=7i,mallocs=17642i,frees=7473i,heap_sys_bytes=6848512i,heap_idle_bytes=1368064i,heap_in_use_bytes=5480448i,heap_released_bytes=0i,total_alloc_bytes=6875560i,heap_alloc_bytes=4457408i,heap_objects_bytes=10169i,num_gc=2i 1480682800000000000
internal_agent,host=tyrion metrics_written=18i,metrics_dropped=0i,metrics_gathered=19i,gather_errors=0i 1480682800000000000
internal_write,output=file,host=tyrion buffer_limit=10000i,write_time_ns=636609i,metrics_written=18i,buffer_size=0i 1480682800000000000
internal_gather,input=internal,host=tyrion metrics_gathered=19i,gather_time_ns=442114i 1480682800000000000
internal_gather,input=http_listener,host=tyrion metrics_gathered=0i,gather_time_ns=167285i 1480682800000000000
internal_http_listener,address=:8186,host=tyrion queries_received=0i,writes_received=0i,requests_received=0i,buffers_created=0i,requests_served=0i,pings_received=0i,bytes_received=0i,not_founds_served=0i,pings_served=0i,queries_served=0i,writes_served=0i 1480682800000000000
```
