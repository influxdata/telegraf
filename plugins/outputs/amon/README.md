# Amon Output Plugin

This plugin writes metrics to [Amon monitoring platform][amon]. It requires a
`serverkey` and `amoninstance` URL which can be obtained [here][amon_monitoring]
for your account.

> [!IMPORTANT]
> If point values being sent cannot be converted to a `float64`, the metric is
> skipped.

⭐ Telegraf v0.2.1
🏷️ datastore
💻 all

[amon]: https://www.amon.cx
[amon_monitoring]:https://www.amon.cx/docs/monitoring/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Configuration for Amon Server to send metrics to.
[[outputs.amon]]
  ## Amon Server Key
  server_key = "my-server-key" # required.

  ## Amon Instance URL
  amon_instance = "https://youramoninstance" # required

  ## Connection timeout.
  # timeout = "5s"
```

## Conversions

Metrics are grouped by converting any `_` characters to `.` in the point name
