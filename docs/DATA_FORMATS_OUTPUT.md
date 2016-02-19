# Telegraf Output Data Formats

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

  ## Data format to output. This can be "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## Additional configuration options go here
```

Each data_format has an additional set of configuration options available, which
I'll go over below.

## Influx:

There are no additional configuration options for InfluxDB line-protocol. The
metrics are serialized directly into InfluxDB line-protocol.

#### Influx Configuration:

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output. This can be "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

## Graphite:

The Graphite data format translates Telegraf metrics into _dot_ buckets.
The format is:

```
[prefix].[host tag].[all tags (alphabetical)].[measurement name].[field name] value timestamp
```

Which means the following influx metric -> graphite conversion would happen:

```
cpu,cpu=cpu-total,dc=us-east-1,host=tars usage_idle=98.09,usage_user=0.89 1455320660004257758
=>
tars.cpu-total.us-east-1.cpu.usage_user 0.89 1455320690
tars.cpu-total.us-east-1.cpu.usage_idle 98.09 1455320690
```

`prefix` is a configuration option when using the graphite output data format.

#### Graphite Configuration:

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output. This can be "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  prefix = "telegraf"
```
