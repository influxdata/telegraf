# Amon Output Plugin

This plugin writes metrics to [Amon monitoring platform][amon]. It requires a
`serverkey` and `amoninstance` URL which can be obtained from the
[website][amon_monitoring] for your account.

> [!IMPORTANT]
> If point values being sent cannot be converted to a `float64`, the metric is
> skipped.

⭐ Telegraf v0.2.1
🚩 Telegraf v1.37.0
🔥 Telegraf v1.40.0
🏷️ datastore
💻 all

[amon]: https://www.amon.cx
[amon_monitoring]:https://www.amon.cx/docs/monitoring/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

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
