# TCP Listener Input Plugin

**DEPRECATED: As of version 1.3 the TCP listener plugin has been deprecated in
favor of the [socket_listener plugin](../socket_listener/README.md)**

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Generic TCP listener
[[inputs.tcp_listener]]
  # socket_listener plugin
  # see https://github.com/influxdata/telegraf/tree/master/plugins/inputs/socket_listener
```

## Metrics

## Example Output
