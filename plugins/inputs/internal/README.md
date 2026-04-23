# Telegraf Internal Input Plugin

This plugin collects metrics about the telegraf agent and its plugins.

> [!NOTE]
> Some metrics are aggregates across all instances of a plugin type.

⭐ Telegraf v1.2.0
🏷️ applications
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Collect statistics about itself
[[inputs.internal]]
  ## If true, collect telegraf memory stats.
  # collect_memstats = true

  ## If true, collect metrics from Go's runtime.metrics. For a full list see:
  ##   https://pkg.go.dev/runtime/metrics
  # collect_gostats = false

  ## Collect statistics per plugin instance and not per plugin type
  # per_instance = false
```

## Metrics

memstats are collected using the [Go runtime framework][memstats]

- internal_memstats
  - alloc_bytes
  - frees
  - heap_alloc_bytes
  - heap_idle_bytes
  - heap_in_use_bytes
  - heap_objects_bytes
  - heap_released_bytes
  - heap_sys_bytes
  - mallocs
  - num_gc
  - pointer_lookups
  - sys_bytes
  - total_alloc_bytes

agent stats collect aggregate stats on all telegraf plugins.

- internal_agent
  - gather_errors    -- number of failing collection operations
                        (excluding startup-errors)
  - gather_timeouts  -- number of times a collection took longer than the
                        defined interval
  - metrics_dropped  -- total number of metrics dropped from buffers without
                        sending
  - metrics_gathered -- total number of metrics successfully collected by inputs
  - metrics_rejected -- total number of metrics rejected by service endpoints
  - metrics_written  -- total number of metrics successfully written by outputs
  - write_errors     -- total number of failing write operations
                        (excluding startup-errors)

internal_gather stats collect aggregate stats on all input plugins
that are of the same input type. They are tagged with `input=<plugin_name>`
`version=<telegraf_version>` and `go_version=<go_build_version>`.

- internal_gather
  - errors            -- number of errors *logged* by the plugin
  - gather_errors     -- number of failing collection operations
                         (excluding startup-errors)
  - gather_time_ns    -- duration of the collection operation
  - gather_timeouts   -- number of times a collection took longer than the
                         defined interval
  - metrics_gathered  -- number of metrics produced by the plugin
  - startup_errors    -- number of errors while starting the plugin

internal_write stats collect aggregate stats on all output plugins
that are of the same input type. They are tagged with `output=<plugin_name>`
and `version=<telegraf_version>`.

- internal_write
  - buffer_limit      -- size of the metric buffer as configured by the user
  - buffer_size       -- number of metrics in the buffer
  - errors            -- number of errors *logged* by the plugin
  - metrics_added     -- number of metrics added to the plugin for writing
  - metrics_dropped   -- number of metrics dropped from buffer without sending
  - metrics_filtered  -- number of metrics not passing the metric-filter
  - metrics_rejected  -- number of metrics rejected by the service endpoint
  - metrics_written   -- number of metrics successfully written
  - startup_errors    -- number of errors while starting the plugin
  - write_errors      -- number of failing write operations
                         (excluding startup-errors)
  - write_time_ns     -- duration of the write operation

internal_<plugin_name> are metrics which are defined on a per-plugin basis, and
usually contain tags which differentiate each instance of a particular type of
plugin and `version=<telegraf_version>`.

- internal_<plugin_name>
  - individual plugin-specific fields, such as requests counts.

All measurements for specific plugins are tagged with information relevant
to each particular plugin and with `version=<telegraf_version>`.

[memstats]: https://golang.org/pkg/runtime/#MemStats

## Example Output

```text
internal_memstats,host=tyrion alloc_bytes=4457408i,sys_bytes=10590456i,pointer_lookups=7i,mallocs=17642i,frees=7473i,heap_sys_bytes=6848512i,heap_idle_bytes=1368064i,heap_in_use_bytes=5480448i,heap_released_bytes=0i,total_alloc_bytes=6875560i,heap_alloc_bytes=4457408i,heap_objects_bytes=10169i,num_gc=2i 1480682800000000000
internal_agent,host=tyrion,go_version=1.12.7,version=1.99.0 metrics_written=18i,metrics_dropped=0i,metrics_gathered=19i,metrics_rejected=0i,gather_errors=0i,gather_timeouts=0i,write_errors=0i 1480682800000000000
internal_write,output=file,host=tyrion,version=1.99.0 buffer_limit=10000i,buffer_size=0i,errors=0i,metrics_added=18i,metrics_dropped=0i,metrics_filtered=0i,metrics_rejected=0i,metrics_written=18i,startup_errors=1i,write_errors=0i,write_time_ns=636609i 1480682800000000000
internal_gather,input=internal,host=tyrion,version=1.99.0 errors=2i,gather_errors=1i,gather_time_ns=442114i,gather_timeouts=0i,metrics_gathered=19i,startup_errors=0i 1480682800000000000
internal_gather,input=http_listener,host=tyrion,version=1.99.0 errors=0i,gather_errors=0i,gather_time_ns=167285i,gather_timeouts=0i,metrics_gathered=0i,startup_errors=0i 1480682800000000000
internal_http_listener,address=:8186,host=tyrion,version=1.99.0 queries_received=0i,writes_received=0i,requests_received=0i,buffers_created=0i,requests_served=0i,pings_received=0i,bytes_received=0i,not_founds_served=0i,pings_served=0i,queries_served=0i,writes_served=0i 1480682800000000000
internal_mqtt_consumer,host=tyrion,version=1.99.0 messages_received=622i,payload_size=37942i 1657282270000000000
```
