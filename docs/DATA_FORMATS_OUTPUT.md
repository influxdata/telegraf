# Output Data Formats

In addition to output specific data formats, Telegraf supports a set of
standard data formats that may be selected from when configuring many output
plugins.

1. [InfluxDB Line Protocol](#influx)
1. [JSON](#json)
1. [Graphite](#graphite)

You will be able to identify the plugins with support by the presence of a
`data_format` config option, for example, in the `file` output plugin:
```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

## Influx

The `influx` data format outputs metrics using
[InfluxDB Line Protocol](https://docs.influxdata.com/influxdb/latest/write_protocols/line_protocol_tutorial/).
This is the recommended format unless another format is required for
interoperability.

### Influx Configuration
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
```

## Graphite

The Graphite data format translates Telegraf metrics into _dot_ buckets. A
template can be specified for the output of Telegraf metrics into Graphite
buckets. The default template is:

```
template = "host.tags.measurement.field"
```

In the above template, we have four parts:

1. _host_ is a tag key. This can be any tag key that is in the Telegraf
metric(s). If the key doesn't exist, it will be ignored. If it does exist, the
tag value will be filled in.
1. _tags_ is a special keyword that outputs all remaining tag values, separated
by dots and in alphabetical order (by tag key). These will be filled after all
tag keys are filled.
1. _measurement_ is a special keyword that outputs the measurement name.
1. _field_ is a special keyword that outputs the field name.

Which means the following influx metric -> graphite conversion would happen:

```
cpu,cpu=cpu-total,dc=us-east-1,host=tars usage_idle=98.09,usage_user=0.89 1455320660004257758
=>
tars.cpu-total.us-east-1.cpu.usage_user 0.89 1455320690
tars.cpu-total.us-east-1.cpu.usage_idle 98.09 1455320690
```

Fields with string values will be skipped.  Boolean fields will be converted
to 1 (true) or 0 (false).

### Graphite Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "graphite"

  # prefix each graphite bucket
  prefix = "telegraf"
  # graphite template
  template = "host.tags.measurement.field"
```

## JSON

The JSON output data format output for a single metric is in the
form:
```json
{
    "fields": {
        "field_1": 30,
        "field_2": 4,
        "field_N": 59,
        "n_images": 660
    },
    "name": "docker",
    "tags": {
        "host": "raynor"
    },
    "timestamp": 1458229140
}
```

When an output plugin needs to emit multiple metrics at one time, it may use
the batch format.  The use of batch format is determined by the plugin,
reference the documentation for the specific plugin.
```json
{
    "metrics": [
        {
            "fields": {
                "field_1": 30,
                "field_2": 4,
                "field_N": 59,
                "n_images": 660
            },
            "name": "docker",
            "tags": {
                "host": "raynor"
            },
            "timestamp": 1458229140
        },
        {
            "fields": {
                "field_1": 30,
                "field_2": 4,
                "field_N": 59,
                "n_images": 660
            },
            "name": "docker",
            "tags": {
                "host": "raynor"
            },
            "timestamp": 1458229140
        }
    ]
}
```

### JSON Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "json"

  ## The resolution to use for the metric timestamp.  Must be a duration string
  ## such as "1ns", "1us", "1ms", "10ms", "1s".  Durations are truncated to
  ## the power of 10 less than the specified units.
  json_timestamp_units = "1s"
```
