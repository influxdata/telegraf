# Output Data Formats

In addition to output specific data formats, Telegraf supports a set of
standard data formats that may be selected from when configuring many output
plugins.

1. [InfluxDB Line Protocol](#influx)
1. [JSON](#json)
1. [Graphite](#graphite)
1. [SplunkMetric](#splunkmetric)

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

The Graphite data format is translated from Telegraf Metrics using either the
template pattern or tag support method.  You can select between the two
methods using the [`graphite_tag_support`](#graphite-tag-support) option.  When set, the tag support
method is used, otherwise the [`template` pattern](#template-pattern) is used.

#### Template Pattern

The `template` option describes how Telegraf traslates metrics into _dot_
buckets.  The default template is:

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

**Example Conversion**:

```
cpu,cpu=cpu-total,dc=us-east-1,host=tars usage_idle=98.09,usage_user=0.89 1455320660004257758
=>
tars.cpu-total.us-east-1.cpu.usage_user 0.89 1455320690
tars.cpu-total.us-east-1.cpu.usage_idle 98.09 1455320690
```

Fields with string values will be skipped.  Boolean fields will be converted
to 1 (true) or 0 (false).

#### Graphite Tag Support

When the `graphite_tag_support` option is enabled, the template pattern is not
used.  Instead, tags are encoded using
[Graphite tag support](http://graphite.readthedocs.io/en/latest/tags.html)
added in Graphite 1.1.  The `metric_path` is a combination of the optional
`prefix` option, measurement name, and field name.

The tag `name` is reserved by Graphite, any conflicting tags and will be encoded as `_name`.

**Example Conversion**:
```
cpu,cpu=cpu-total,dc=us-east-1,host=tars usage_idle=98.09,usage_user=0.89 1455320660004257758
=>
cpu.usage_user;cpu=cpu-total;dc=us-east-1;host=tars 0.89 1455320690
cpu.usage_idle;cpu=cpu-total;dc=us-east-1;host=tars 98.09 1455320690
```

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

  ## Prefix added to each graphite bucket
  prefix = "telegraf"
  ## Graphite template pattern
  template = "host.tags.measurement.field"

  ## Support Graphite tags, recommended to enable when using Graphite 1.1 or later.
  # graphite_tag_support = false
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

## SplunkMetric

The SplunkMetric format translates the telegraf metrics into a JSON format compatible
with a Splunk Metrics index. Normally you would use this to send the metrics to a
HTTP Event Collector (HEC), it follows the format specified in the document [here.](http://dev.splunk.com/view/SP-CAAAFDN#json)

```toml
[[outputs.http]]
#   ## URL is the address to send metrics to
   url = "https://localhost:8088/services/collector"
#
#   ## Timeout for HTTP message
#   # timeout = "5s"
#
#   ## HTTP method, one of: "POST" or "PUT"
#   # method = "POST"
#
#   ## HTTP Basic Auth credentials
#   # username = "username"
#   # password = "pa$$word"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
#
#   ## Data format to output.
#   ## Each data format has it's own unique set of configuration options, read
#   ## more about them here:
#   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
   data_format = "splunkmetric"
   # Enable hec_routing 
   hec_routing = true
#
#   ## Additional HTTP headers
    [outputs.http.headers]
#   # Should be set manually to "application/json" for json data_format
      Content-Type = "application/json"
      Authorization = "Splunk xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      X-Splunk-Request-Channel = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
```

It also has the option to output a more basic JSON structure that can be used by a
Splunk Universal Forwarder or modular input.

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "json"
  ## hec_routing defaults to false
  # hec_routing = false
```

Please see the [Splunk Metrics README](../plugins/serializers/splunkmetric/README.md) for additional
information.
