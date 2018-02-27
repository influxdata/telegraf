# Telegraf Output Data Formats

Telegraf is able to serialize metrics into the following output data formats:

1. [InfluxDB Line Protocol](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#influx)
1. [JSON](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#json)
1. [Graphite](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#graphite)

Telegraf metrics, like InfluxDB
[points](https://docs.influxdata.com/influxdb/v0.10/write_protocols/line/),
are a combination of four basic parts:

1. Measurement Name
1. Tags
1. Fields
1. Timestamp

In InfluxDB line protocol, these 4 parts are easily defined in textual form:

```
measurement_name[,tag1=val1,...]  field1=val1[,field2=val2,...]  [timestamp]
```

For Telegraf outputs that write textual data (such as `kafka`, `mqtt`, and `file`),
InfluxDB line protocol was originally the only available output format. But now
we are normalizing telegraf metric "serializers" into a
[plugin-like interface](https://github.com/influxdata/telegraf/tree/master/plugins/serializers)
across all output plugins that can support it.
You will be able to identify a plugin that supports different data formats
by the presence of a `data_format`
config option, for example, in the `file` output plugin:

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## Additional configuration options go here
```

Each data_format has an additional set of configuration options available, which
I'll go over below.

# Influx:

There are no additional configuration options for InfluxDB line-protocol. The
metrics are serialized directly into InfluxDB line-protocol.

### Influx Configuration:

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

# Graphite:

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

### Graphite Configuration:

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

# JSON:

The JSON data format serialized Telegraf metrics in json format. The format is:

```json
{
   "fields":{
      "field_1":30,
      "field_2":4,
      "field_N":59,
      "n_images":660
   },
   "name":"docker",
   "tags":{
      "host":"raynor"
   },
   "timestamp":1458229140
}
```

### JSON Configuration:

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "json"
  json_timestamp_units = "1ns"
```

By default, the timestamp that is output in JSON data format serialized Telegraf
metrics is in seconds. The precision of this timestamp can be adjusted for any output
by adding the optional `json_timestamp_units` parameter to the configuration for
that output. This parameter can be used to set the timestamp units to  nanoseconds (`ns`),
microseconds (`us` or `µs`), milliseconds (`ms`), or seconds (`s`). Note that this
parameter will be truncated to the nearest power of 10 that, so if the `json_timestamp_units`
are set to `15ms` the timestamps for the JSON format serialized Telegraf metrics will be
output in hundredths of a second (`10ms`).
