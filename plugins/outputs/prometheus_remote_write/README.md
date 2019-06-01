# Prometheus Remote Write Output Plugin

This plugin sends metrics to services which speak the [Prometheus Remote Write](https://prometheus.io/docs/operating/integrations/#remote-endpoints-and-storage) format, such as [Cortex](https://github.com/cortexproject/cortex).  ***Note*** Prometheus does not accept writes in this format; it only sends them.

## Configuration

```toml
# Send metrics on Prometheus
[[outputs.prometheus_remote_write]]
  ## URL to send Prometheus remote write requests to.
  url = "http://localhost/push"

  ## HTTP asic auth credentials (optional).
  # basic_username = "username"
  # basic_password = "pa55w0rd"
```

## TODO
- Handle summaries and histograms.
