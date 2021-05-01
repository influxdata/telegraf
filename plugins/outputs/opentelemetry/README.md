# OpenTelemetry Output Plugin

This plugin sends metrics to [OpenTelemetry](https://opentelemetry.io) servers and agents via gRPC.

### Configuration

```toml
[[outputs.opentelemetry]]
  ## Override the OpenTelemetry gRPC service address:port 
  # service_address = "localhost:4317"
  
  ## Override the default request timeout
  # timeout = "5s"
  
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
Metrics received with `metrics_schema=prometheus-v1` are assigned OTel `Metric.name` from the measurement.
Metrics received with `metrics_schema=prometheus-v2` are expected to have measurement `prometheus`.

Also see the OpenTelemetry input plugin for Telegraf.
