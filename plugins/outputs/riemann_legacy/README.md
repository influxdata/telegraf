# Riemann Legacy

This is a deprecated plugin

## Configuration

```toml
# Configuration for the Riemann server to send metrics to
[[outputs.riemann_legacy]]
  ## DEPRECATED: The 'riemann_legacy' plugin is deprecated in version 1.3.0, use 'outputs.riemann' instead (see https://github.com/influxdata/telegraf/issues/1878).
  ## URL of server
  url = "localhost:5555"
  ## transport protocol to use either tcp or udp
  transport = "tcp"
  ## separator to use between input name and field name in Riemann service name
  separator = " "
```
