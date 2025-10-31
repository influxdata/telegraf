# Prometheus Remote Write Parser Plugin

Converts prometheus remote write samples directly into Telegraf metrics. It can
be used with [http_listener_v2][http_listener_v2]. There are no additional
configuration options for Prometheus Remote Write Samples.

[http_listener_v2]: /plugins/inputs/http_listener_v2

## Configuration

```toml
[[inputs.http_listener_v2]]
  ## Address and port to host HTTP listener on
  service_address = ":1234"

  ## Paths to listen to.
  paths = ["/receive"]

  ## Data format to consume.
  data_format = "prometheusremotewrite"

  ## Metric version to use, either 1 or 2
  # metric_version = 2
```

## Example Input

```json
prompb.WriteRequest{
        Timeseries: []*prompb.TimeSeries{
            {
                Labels: []*prompb.Label{
                    {Name: "__name__", Value: "go_gc_duration_seconds"},
                    {Name: "instance", Value: "localhost:9090"},
                    {Name: "job", Value: "prometheus"},
                    {Name: "quantile", Value: "0.99"},
                },
                Samples: []prompb.Sample{
                    {Value: 4.63, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
                },
            },
        },
    }

```

## Example Output (v1)

```text
go_gc_duration_seconds,instance=localhost:9090,job=prometheus,quantile=0.99 value=4.63 1614889298859000000
```

## Example Output (v2)

```text
prometheus_remote_write,instance=localhost:9090,job=prometheus,quantile=0.99 go_gc_duration_seconds=4.63 1614889298859000000
```

## Alignment with Prometheus Remote Write Specification

To align the output with the [InfluxDB v1.x Prometheus Remote Write Specification][spec]
use

- V1: already aligned, it parses metrics according to the spec
- V2: use the [Starlark processor][starlark] with the
      [rename prometheus remote write script][rename_script] to rename the
      measurement to the field-name and rename the field-name to value

[spec]: https://docs.influxdata.com/influxdb/v1.8/supported_protocols/prometheus/#how-prometheus-metrics-are-parsed-in-influxdb
[starlark]: /plugins/processors/starlark/README.md
[rename_script]: https://github.com/influxdata/telegraf/blob/master/plugins/processors/starlark/testdata/rename_prometheus_remote_write.star
