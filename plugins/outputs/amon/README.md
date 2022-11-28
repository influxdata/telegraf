# Amon Output Plugin

This plugin writes to [Amon](https://www.amon.cx) and requires an `serverkey`
and `amoninstance` URL which can be obtained
[here](https://www.amon.cx/docs/monitoring/) for the account.

If the point value being sent cannot be converted to a float64, the metric is
skipped.

Metrics are grouped by converting any `_` characters to `.` in the Point Name.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

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
