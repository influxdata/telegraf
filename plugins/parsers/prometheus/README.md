# Prometheus Text-Based Format Parser Plugin

Parser for [Prometheus Text-Based Format][]. The metrics are parsed directly
into Telegraf metrics. This parser (v1) is used internally in
[prometheus input](/plugins/inputs/prometheus) or can be used in
[http_listener_v2](/plugins/inputs/http_listener_v2) to simulate Pushgateway.

The parser comes in two versions. The version can be selected using
`prometheus_metric_version`. By default, version 2 is used.
In version 1, you will get one Telegraf metric with one field per Prometheus
metric for "simple" types like Gauge and Counter but a Telegraf metric with
multiple fields for "complex" types like Summary or Histogram.

Version 2 converts each Prometheus metric to a corresponding Telegraf metric
with one field each. The process will filter NaNs in values and skip
the corresponding metrics. Notably, the name of the generated telegraf metrics
will be set to "prometheus".

For example, consider this Prometheus metric:

```
# HELP my_awesome_gauge Some gauge
# TYPE my_awesome_gauge gauge
my_awesome_gauge{id="a8z1xw",} 0.01838084048871169 1755701460000
```

Here is how the two versions would parse it

```
# V1
{"fields":{"gauge":0.01838084048871169},"name":"my_awesome_gauge","tags":{"id":"a8z1xw"},"timestamp":1755701460}
# V2
{"fields":{"my_awesome_gauge":0.01838084048871169},"name":"prometheus","tags":{"id":"a8z1xw"},"timestamp":1755701460}
```


[Prometheus Text-Based Format]: https://prometheus.io/docs/instrumenting/exposition_formats/#text-based-format

## Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "prometheus"
  # V2 is already the default
  # prometheus_metric_version = 2 

```
