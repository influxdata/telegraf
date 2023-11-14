# Noop Processor Plugin

The noop processor plugin does nothing to metrics. Instead it can be used to
apply the global configuration options after other processing. Global config
options like tagpass, fieldpass, and others are applied before a processor,
aggregator, or output. As such a user might want to apply these after doing
processing, but before an output or another processor.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Do nothing processor
[[processors.noop]]
```

## Examples

Consider a use-case where you have processed a metric based on a tag, but no
longer need that tag for additional processing:

```toml
[[processors.ifname]]
  order = 1
  ...

[[processors.noop]]
  order = 2
  tagdrop = ["useless_tag"]
```
