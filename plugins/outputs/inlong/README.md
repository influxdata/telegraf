# Inlong Output Plugin

This plugin publishes metrics to an [Apache InLong][inlong] instance.

⭐ Telegraf v1.35.0
🏷️ messaging
💻 all

[inlong]: https://inlong.apache.org

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send telegraf metrics to Apache Inlong
[[outputs.inlong]]
  ## Unique identifier for the data-stream group
  group_id = "test_group"  

  ## Unique identifier for the data stream within its group
  stream_id = "test_stream"  

  ## URL to obtain the Inlong data-proxy IP list for sending the data
  manager_url = "http://127.0.0.1:8083"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "csv"
```
