# OpenTelemetry Output Plugin

This plugin sends metrics to [OpenTelemetry](https://opentelemetry.io) servers
and agents via gRPC.

## Configuration

```toml
# Send OpenTelemetry metrics over gRPC
[[outputs.opentelemetry]]
  ## Override the default (localhost:4317) OpenTelemetry gRPC service
  ## address:port
  # service_address = "localhost:4317"

  ## Override the default (5s) request timeout
  # timeout = "5s"

  ## Optional TLS Config.
  ##
  ## Root certificates for verifying server certificates encoded in PEM format.
  # tls_ca = "/etc/telegraf/ca.pem"
  ## The public and private keypairs for the client encoded in PEM format.
  ## May contain intermediate certificates.
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS, but skip TLS chain and host verification.
  # insecure_skip_verify = false
  ## Send the specified TLS server name via SNI.
  # tls_server_name = "foo.example.com"

  ## Override the default (gzip) compression used to send data.
  ## Supports: "gzip", "none"
  # compression = "gzip"

  ## Additional OpenTelemetry resource attributes
  # [outputs.opentelemetry.attributes]
  # "service.name" = "demo"

  ## Additional gRPC request metadata
  # [outputs.opentelemetry.headers]
  # key1 = "value1"
```

### Schema

The InfluxDB->OpenTelemetry conversion [schema][] and [implementation][] are
hosted on [GitHub][repo].

For metrics, two input schemata exist.  Line protocol with measurement name
`prometheus` is assumed to have a schema matching [Prometheus input
plugin](../../inputs/prometheus/README.md) when `metric_version = 2`.  Line
protocol with other measurement names is assumed to have schema matching
[Prometheus input plugin](../../inputs/prometheus/README.md) when
`metric_version = 1`.  If both schema assumptions fail, then the line protocol
data is interpreted as:

- Metric type = gauge (or counter, if indicated by the input plugin)
- Metric name = `[measurement]_[field key]`
- Metric value = line protocol field value, cast to float
- Metric labels = line protocol tags

Also see the [OpenTelemetry input plugin](../../inputs/opentelemetry/README.md).

[schema]: https://github.com/influxdata/influxdb-observability/blob/main/docs/index.md

[implementation]: https://github.com/influxdata/influxdb-observability/tree/main/influx2otel

[repo]: https://github.com/influxdata/influxdb-observability
