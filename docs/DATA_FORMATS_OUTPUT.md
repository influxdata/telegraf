# Output Data Formats

In addition to output specific data formats, Telegraf supports a set of
standard data formats that may be selected from when configuring many output
plugins.

1. [InfluxDB Line Protocol](/plugins/serializers/influx)
1. [Binary](/plugins/serializers/binary)
1. [Carbon2](/plugins/serializers/carbon2)
1. [CloudEvents](/plugins/serializers/cloudevents)
1. [CSV](/plugins/serializers/csv)
1. [Graphite](/plugins/serializers/graphite)
1. [JSON](/plugins/serializers/json)
1. [MessagePack](/plugins/serializers/msgpack)
1. [Prometheus](/plugins/serializers/prometheus)
1. [Prometheus Remote Write](/plugins/serializers/prometheusremotewrite)
1. [ServiceNow Metrics](/plugins/serializers/nowmetric)
1. [SplunkMetric](/plugins/serializers/splunkmetric)
1. [Template](/plugins/serializers/template)
1. [Wavefront](/plugins/serializers/wavefront)

You will be able to identify the plugins with support by the presence of a
`data_format` config option, for example, in the `file` output plugin:

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout"]

  ## Data format to output.
  data_format = "influx"
```
