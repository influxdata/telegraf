# OpenTelemetry Output Plugin

This plugin sends metrics to [OpenTelemetry](https://opentelemetry.io) servers and agents via gRPC.

### Configuration

```toml
[[outputs.opentelemetry]]
  ## Override the default (localhost:4317) OpenTelemetry gRPC service
  ## address:port
  # service_address = "localhost:4317"

  ## Override the default (5s) request timeout
  # timeout = "5s"

  ## Override the default (prometheus-v1) metrics schema.
  ## Supports: "prometheus-v1", "prometheus-v2"
  ## For more information about the alternatives, read the Prometheus input
  ## plugin notes.
  # metrics_schema = "prometheus-v1"

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

#### Schema

The InfluxDB->OpenTelemetry conversion [schema](https://github.com/influxdata/influxdb-observability/blob/main/docs/index.md)
and [implementation](https://github.com/influxdata/influxdb-observability/tree/main/influx2otel)
are hosted at https://github.com/influxdata/influxdb-observability .

For metrics, two output schemata exist.
When this plugin is configured with `metrics_schema=prometheus-v1`,
measurement name is used for OTel `Metric.name`.
When this plugin is configured with `metrics_schema=prometheus-v2`,
input points are expected to have measurement `prometheus`,
and OTel `Metric.name` is inferred from field keys.

Also see the OpenTelemetry input plugin for Telegraf.
