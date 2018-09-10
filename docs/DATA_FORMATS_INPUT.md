# Input Data Formats

Telegraf is able to parse the following input data formats into metrics:

1. [InfluxDB Line Protocol](/plugins/parsers/influx)
1. [JSON](/plugins/parsers/json)
1. [Graphite](/plugins/parsers/graphite)
1. [Value](/plugins/parsers/value), ie: 45 or "booyah"
1. [Nagios](/plugins/parsers/nagios)
1. [Collectd](/plugins/parsers/collectd)
1. [Dropwizard](/plugins/parsers/dropwizard)
1. [Grok](/plugins/parsers/grok)
1. [Logfmt](/plugins/parsers/logfmt)
1. [Wavefront](/plugins/parsers/wavefront)
1. [CSV](/plugins/parsers/csv)

Telegraf metrics, similar to InfluxDB's [points][influxdb key concepts], are a
combination of four basic parts:

[influxdb key concepts]: https://docs.influxdata.com/influxdb/v1.6/concepts/key_concepts/

1. Measurement Name
1. Tags
1. Fields
1. Timestamp

These four parts are easily defined when using InfluxDB line-protocol as a
data format. But there are other data formats that users may want to use which
require more advanced configuration to create usable Telegraf metrics.

Plugins such as `exec` and `kafka_consumer` parse textual data. Up until now,
these plugins were statically configured to parse just a single
data format. `exec` mostly only supported parsing JSON, and `kafka_consumer` only
supported data in InfluxDB line-protocol.

But now we are normalizing the parsing of various data formats across all
plugins that can support it. You will be able to identify a plugin that supports
different data formats by the presence of a `data_format` config option, for
example, in the exec plugin:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/tmp/test.sh", "/usr/bin/mycollector --foo=bar"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"

  ## Additional configuration options go here
```

Each data_format has an additional set of configuration options available, which
I'll go over below.

