# discard Output Plugin

This output plugin simply drops all metrics that are sent to it. It is only
meant to be used for testing purposes.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Send metrics to nowhere at all
[[outputs.discard]]
  # no configuration
```
