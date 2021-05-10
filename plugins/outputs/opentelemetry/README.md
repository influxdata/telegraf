# OpenTelemetry Output Plugin

This plugin sends metrics to [OpenTelemetry](https://opentelemetry.io) servers and agents via gRPC.

### Configuration

```toml
[[outputs.opentelemetry]]
  ## Override the OpenTelemetry gRPC service address:port 
  # service_address = "localhost:4317"
  
  ## Select a schema for metrics: "prometheus-v1" or "prometheus-v2"
  ## For more information about the alternatives, read the Prometheus input
  ## plugin notes.
  # metrics_schema = "prometheus-v1"
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
