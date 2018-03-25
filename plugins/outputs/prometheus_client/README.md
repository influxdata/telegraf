# Prometheus Client Service Output Plugin

This plugin starts a [Prometheus](https://prometheus.io/) Client, it exposes all metrics on `/metrics` (default) to be polled by a Prometheus server.

## Configuration

```
# Publish all metrics to /metrics for Prometheus to scrape
[[outputs.prometheus_client]]
  # Address to listen on
  listen = ":9273"

  # Use TLS
  tls_cert = "/etc/ssl/telegraf.crt"
  tls_key = "/etc/ssl/telegraf.key"

  # Use http basic authentication
  basic_username = "Foo"
  basic_password = "Bar"

  # Path to publish the metrics on, defaults to /metrics
  path = "/metrics"   

  # Expiration interval for each metric. 0 == no expiration
  expiration_interval = "60s"

  # Send string metrics as Prometheus labels.
  # Unless set to false all string metrics will be sent as labels.
  string_as_label = true
```
