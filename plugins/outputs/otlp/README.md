# OpenTelemetry Protocol Output Plugin

This plugin writes to any backend that support the OpenTelemetry Protocol (OTLP)

Metrics are named by combining the metric name and field key - eg: `cpu.usage_user`

Additional resource attributes can be configured by `attributes`.

Additional gRPC metadata can be configured by `headers`.

### Configuration

```toml
[[outputs.otlp]]
  ## OpenTelemetry endpoint
  endpoint = "localhost:4317"

  ## Timeout used when sending data over grpc
  timeout = 10s

  # Additional resource attributes
  [outputs.otlp.attributes]
  	service.name = "demo"

  # Additional grpc metadata
  [outputs.otlp.headers]
    key1 = "value1"

```

### Restrictions

OTLP does not support string values in custom metrics, any string
fields will not be written.
