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

  ## It comes from the Inlong system, a `DataStream` (StreamID) defines a specific data pipeline with a unique source, 
  ## format, and destination. It is part of a DataStreamGroup and operates within its business context.
  stream_id = "test_stream"  # (string) Unique identifier for the data stream within its group

  ## The URL used to obtain the Inlong DataProxy IP list to which the data will be sent
  manager_url = "http://127.0.0.1:8083"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "csv"

  ## The delimiter used when serializing data in CSV format needs to be consistent with the delimiter
  ## configured for Inlong, so that the data can be parsed properly after it reaches Inlong.
  ## It can be a space, vertical bar (|), comma (,), semicolon (;), asterisk (*), double quotes ("), etc.
  csv_separator = "|"

  ## The final output field order here needs to be consistent with the field order defined by the data
  ## stream in Inlong
  csv_columns = ["field.key","file.value"]
```
