# OpenTelemetry Output Plugin

This plugin writes to any backend that support the [OpenTelemetry Protocol (OTLP)](https://github.com/open-telemetry/opentelemetry-specification/tree/main/specification/protocol). Metrics are named by combining the metric name and field key - eg: `cpu.usage_user`. Additional [resource attributes](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/sdk.md#resource-sdk) can be configured by setting `attributes` in your configuration. If the endpoint you're connecting to requires additional gRPC headers, this data can be configured via the `headers` option.

### Configuration

```toml
[[outputs.opentelemetry]]
  ## OpenTelemetry endpoint
  # endpoint = "http://localhost:4317"

  ## Timeout when sending data over grpc
  # timeout = "10s"

  ## Compression used to send data, supports: "gzip", "none"
  # compression = "gzip"

  ## Optional TLS Config for use on gRPC connections.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  # Additional resource attributes
  [outputs.opentelemetry.attributes]
    "service.name" = "demo"

  # Additional grpc metadata
  [outputs.opentelemetry.headers]
    key1 = "value1"

```

### Restrictions

* OpenTelemetry protocol does not support string values in custom metrics, any string fields will be omitted and not written to the endpoint.
* The plugin implements the protocol using protobufs over gRPC, so the backend must support this protocol.