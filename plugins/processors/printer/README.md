# Printer Processor Plugin

The printer processor plugin simple prints every metric passing through it.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

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
  ## i.e.: `42u`.  You will need a version of InfluxDB supporting unsigned
  ## integer values.  Enabling this option will result in field type errors if
  ## existing data has been written.
  # influx_uint_support = false

  ## When true, Telegraf will omit the timestamp on data to allow InfluxDB
  ## to set the timestamp of the data during ingestion. This is generally NOT
  ## what you want as it can lead to data points captured at different times
  ## getting omitted due to similar data.
  # influx_omit_timestamp = false
```

## Tags

No tags are applied by this processor.
