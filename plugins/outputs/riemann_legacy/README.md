# Riemann Legacy Output Plugin

This is a deprecated plugin. Please use the [Riemann Output Plugin][new]
instead.

[new]: ../riemann/README.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Configuration for the Riemann server to send metrics to
[[outputs.riemann_legacy]]
  ## URL of server
  url = "localhost:5555"
  ## transport protocol to use either tcp or udp
  transport = "tcp"
  ## separator to use between input name and field name in Riemann service name
  separator = " "
```
