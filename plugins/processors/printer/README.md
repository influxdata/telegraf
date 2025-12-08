# Printer Processor Plugin

This plugin prints every metric passing through it to the standard output.

⭐ Telegraf v1.1.0
🏷️ testing
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Print all metrics that pass through this filter.
[[processors.printer]]
  ## Maximum line length in bytes.  Useful only for debugging.
  # influx_max_line_bytes = 0

  ## When true, fields will be output in ascending lexical order.  Enabling
  ## this option will result in decreased performance and is only recommended
  ## when you need predictable ordering while debugging.
  # influx_sort_fields = false

  ## When true, Telegraf will output unsigned integers as unsigned values,
  ## for example `42u`.  You will need a version of InfluxDB supporting unsigned
  ## integer values.  Enabling this option will result in field type errors if
  ## existing data has been written.
  # influx_uint_support = false

  ## When true, Telegraf will omit the timestamp on data to allow InfluxDB
  ## to set the timestamp of the data during ingestion. This is generally NOT
  ## what you want as it can lead to data points captured at different times
  ## getting omitted due to similar data.
  # influx_omit_timestamp = false
```
