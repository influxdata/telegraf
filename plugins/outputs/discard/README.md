# Discard Output Plugin

This plugin discards all metrics written to it and is meant for testing
purposes.

⭐ Telegraf v1.2.0
🏷️ testing
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send metrics to nowhere at all
[[outputs.discard]]
  # no configuration
```
