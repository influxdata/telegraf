# Prometheus Client Service Output Plugin

This plugin starts a [Prometheus](https://prometheus.io/) Client, it exposes all metrics on `/metrics` to be polled by a Prometheus server.

## Configuration

```
# Publish all metrics to /metrics for Prometheus to scrape
[[outputs.prometheus_client]]
  # Address to listen on
  listen = ":9126"

  # Expiration interval for each metric. 0 == no expiration
  expiration_interval = "60s"
```
