# Prometheus remote write

The `prometheusremotewrite` data format converts metrics into the Prometheus protobuf
exposition format.

**Warning**: When generating histogram and summary types, output may
not be correct if the metric spans multiple batches.  This issue can be
somewhat, but not fully, mitigated by using outputs that support writing in
"batch format".  When using histogram and summary types, it is recommended to
use only the `prometheus_client` output.

## Configuration

```toml
[[outputs.http]]
  ## URL is the address to send metrics to
  url = "https://cortex/api/prom/push"

  ## Optional TLS Config
  tls_ca = "/etc/telegraf/ca.pem"
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

  ## Data format to output.
  data_format = "prometheusremotewrite"

  [outputs.http.headers]
     Content-Type = "application/x-protobuf"
     Content-Encoding = "snappy"
     X-Prometheus-Remote-Write-Version = "0.1.0"
```

### Metrics

A Prometheus metric is created for each integer, float, boolean or unsigned
field.  Boolean values are converted to *1.0* for true and *0.0* for false.

The Prometheus metric names are produced by joining the measurement name with
the field key.  In the special case where the measurement name is `prometheus`
it is not included in the final metric name.

Prometheus labels are produced for each tag.

**Note:** String fields are ignored and do not produce Prometheus metrics.
