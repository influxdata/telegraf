# UDP Listener Input Plugin

**DEPRECATED: As of version 1.3 the UDP listener plugin has been deprecated in
favor of the [socket_listener plugin](../socket_listener/README.md)**

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Generic UDP listener
[[inputs.udp_listener]]
  # see https://github.com/influxdata/telegraf/tree/master/plugins/inputs/socket_listener
```
