# Telegraf Internal Input Plugin

This plugin collects metrics about the telegraf agent and its plugins.

> [!NOTE]
> Some metrics are aggregates across all instances of a plugin type.

⭐ Telegraf v1.2.0
🏷️ applications
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

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
  - gather_errors
  - gather_timeouts
  - metrics_dropped
  - metrics_gathered
  - metrics_written

internal_gather stats collect aggregate stats on all input plugins
that are of the same input type. They are tagged with `input=<plugin_name>`
`version=<telegraf_version>` and `go_version=<go_build_version>`.

- internal_gather
  - gather_time_ns
  - metrics_gathered
  - gather_timeouts

internal_write stats collect aggregate stats on all output plugins
that are of the same input type. They are tagged with `output=<plugin_name>`
and `version=<telegraf_version>`.

- internal_write
  - buffer_limit
  - buffer_size
  - metrics_added
  - metrics_written
  - metrics_dropped
  - metrics_filtered
  - write_time_ns

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
internal_agent,host=tyrion,go_version=1.12.7,version=1.99.0 metrics_written=18i,metrics_dropped=0i,metrics_gathered=19i,gather_errors=0i,gather_timeouts=0i 1480682800000000000
internal_write,output=file,host=tyrion,version=1.99.0 buffer_limit=10000i,write_time_ns=636609i,metrics_added=18i,metrics_written=18i,buffer_size=0i 1480682800000000000
internal_gather,input=internal,host=tyrion,version=1.99.0 metrics_gathered=19i,gather_time_ns=442114i,gather_timeouts=0i 1480682800000000000
internal_gather,input=http_listener,host=tyrion,version=1.99.0 metrics_gathered=0i,gather_time_ns=167285i,gather_timeouts=0i 1480682800000000000
internal_http_listener,address=:8186,host=tyrion,version=1.99.0 queries_received=0i,writes_received=0i,requests_received=0i,buffers_created=0i,requests_served=0i,pings_received=0i,bytes_received=0i,not_founds_served=0i,pings_served=0i,queries_served=0i,writes_served=0i 1480682800000000000
internal_mqtt_consumer,host=tyrion,version=1.99.0 messages_received=622i,payload_size=37942i 1657282270000000000
```
