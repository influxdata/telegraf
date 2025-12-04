# Pivot Processor Plugin

This plugin rotates single-valued metrics into a multi-field metric. The result
is a more compact representation for applying mathematical operators to or do
comparisons between metrics or flatten fields.

> [!TIP]
> To perform the reverse operation use the [unpivot][unpivot] processor.

⭐ Telegraf v1.12.0
🏷️ transformation
💻 all

[unpivot]: /plugins/processors/unpivot/README.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Rotate a single valued metric into a multi field metric
[[processors.pivot]]
  ## Tag to use for naming the new field.
  tag_key = "name"
  ## Field to use as the value of the new field.
  value_key = "value"
```

## Example

```diff
- cpu,cpu=cpu0,name=time_idle value=42i
- cpu,cpu=cpu0,name=time_user value=43i
+ cpu,cpu=cpu0 time_idle=42i
+ cpu,cpu=cpu0 time_user=43i
```
