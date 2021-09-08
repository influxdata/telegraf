# Input Data Formats

Telegraf contains many general purpose plugins that support parsing input data
using a configurable parser into [metrics][].  This allows, for example, the
`kafka_consumer` input plugin to process messages in either InfluxDB Line
Protocol or in JSON format.

- [Collectd](/plugins/parsers/collectd)
- [CSV](/plugins/parsers/csv)
- [Dropwizard](/plugins/parsers/dropwizard)
- [Graphite](/plugins/parsers/graphite)
- [Grok](/plugins/parsers/grok)
- [InfluxDB Line Protocol](/plugins/parsers/influx)
- [JSON](/plugins/parsers/json)
- [JSON v2](/plugins/parsers/json_v2)
- [Logfmt](/plugins/parsers/logfmt)
- [Nagios](/plugins/parsers/nagios)
- [Prometheus](/plugins/parsers/prometheus)
- [PrometheusRemoteWrite](/plugins/parsers/prometheusremotewrite)
- [Value](/plugins/parsers/value), ie: 45 or "booyah"
- [Wavefront](/plugins/parsers/wavefront)
- [XPath](/plugins/parsers/xpath) (supports XML, JSON, MessagePack, Protocol Buffers)

Any input plugin containing the `data_format` option can use it to select the
desired parser:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/tmp/test.sh", "/usr/bin/mycollector --foo=bar"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume.
  data_format = "json"
```

[metrics]: /docs/METRICS.md
