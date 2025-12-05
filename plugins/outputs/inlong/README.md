# Inlong Output Plugin

This plugin publishes metrics to an [Apache InLong][inlong] instance.

⭐ Telegraf v1.35.0
🏷️ messaging
💻 all

[inlong]: https://inlong.apache.org

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send telegraf metrics to Apache Inlong
[[outputs.inlong]]
  ## Manager URL to obtain the Inlong data-proxy IP list for sending the data
  url = "http://127.0.0.1:8083"

  ## Unique identifier for the data-stream group
  group_id = "telegraf"  

  ## Unique identifier for the data stream within its group
  stream_id = "telegraf"  

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
```
