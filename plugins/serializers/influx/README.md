# Influx

The `influx` data format outputs metrics into [InfluxDB Line Protocol][line
protocol].  This is the recommended format unless another format is required
for interoperability.

## Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## Maximum line length in bytes.  Useful only for debugging.
  influx_max_line_bytes = 0

  ## When true, fields will be output in ascending lexical order.  Enabling
  ## this option will result in decreased performance and is only recommended
  ## when you need predictable ordering while debugging.
  influx_sort_fields = false

  ## When true, Telegraf will output unsigned integers as unsigned values,
  ## i.e.: `42u`.  You will need a version of InfluxDB supporting unsigned
  ## integer values.  Enabling this option will result in field type errors if
  ## existing data has been written.
  influx_uint_support = false

  ## By default, the line format timestamp is at nanosecond precision. The
  ## precision can be adjusted here. This parameter can be used to set the
  ## timestamp units to nanoseconds (`ns`), microseconds (`us` or `µs`),
  ## milliseconds (`ms`), or seconds (`s`). Note that this parameter will be
  ## truncated to the nearest power of 10, so if the `influx_timestamp_units`
  ## are set to `15ms` the timestamps for the serialized line will be output in
  ## hundredths of a second (`10ms`).
  influx_timestamp_units = "1ns"
```

## Metrics

Conversion is direct taking into account some limitations of the Line Protocol
format:

- Float fields that are `NaN` or `Inf` are skipped.
- Trailing backslash `\` characters are removed from tag keys and values.
- Tags with a key or value that is the empty string are skipped.
- When not using `influx_uint_support`, unsigned integers are capped at the max int64.

[line protocol]: https://docs.influxdata.com/influxdb/latest/write_protocols/line_protocol_tutorial/
