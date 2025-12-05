# Clone Processor Plugin

This plugin creates a copy of each metric passing through it, preserving the
original metric and allowing modifications such as [metric modifiers][modifiers]
in the copied metric.

> [!NOTE]
> [Metric filtering][filtering] options apply to both the clone and the
> original metric.

⭐ Telegraf v1.13.0
🏷️ transformation
💻 all

[modifiers]: /docs/CONFIGURATION.md#modifiers
[filtering]: /docs/CONFIGURATION.md#metric-filtering

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Apply metric modifications using override semantics.
[[processors.clone]]
  ## All modifications on inputs and aggregators can be overridden:
  # name_override = "new_name"
  # name_prefix = "new_name_prefix"
  # name_suffix = "new_name_suffix"

  ## Tags to be added (all values must be strings)
  # [processors.clone.tags]
  #   additional_tag = "tag_value"
```
