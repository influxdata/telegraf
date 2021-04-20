# OpenTelemetry Output Plugin

This plugin writes to any backend that support the OpenTelemetry Protocol (OTLP)

Metrics are named by combining the metric name and field key - eg: `cpu.usage_user`

Additional resource attributes can be configured by `attributes`.

Additional gRPC metadata can be configured by `headers`.

### Configuration

```toml
[[outputs.opentelemetry]]
  ## OpenTelemetry endpoint
  endpoint = "http://localhost:4317"

  ## Timeout when sending data over grpc
  timeout = "10s"

  ## Compression used to send data, supports: "gzip", "none"
  compression = "gzip"

  ## Optional TLS Config for use on gRPC connections.
  tls_ca = "/etc/telegraf/ca.pem"
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false

  # Additional resource attributes
  [outputs.opentelemetry.attributes]
      "service.name" = "demo"

  # Additional grpc metadata
  [outputs.opentelemetry.headers]
      key1 = "value1"

```

### Restrictions

OTLP does not support string values in custom metrics, any string
fields will be omitted and not written to the endpoint.
