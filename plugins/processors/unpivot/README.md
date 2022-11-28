# Unpivot Processor Plugin

You can use the `unpivot` processor to rotate a multi field series into single
valued metrics.  This transformation often results in data that is more easy to
aggregate across fields.

To perform the reverse operation use the [pivot] processor.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Rotate multi field metric into several single field metrics
[[processors.unpivot]]
  ## Tag to use for the name.
  tag_key = "name"
  ## Field to use for the name of the value.
  value_key = "value"
```

## Example

```diff
- cpu,cpu=cpu0 time_idle=42i,time_user=43i
+ cpu,cpu=cpu0,name=time_idle value=42i
+ cpu,cpu=cpu0,name=time_user value=43i
```

[pivot]: /plugins/processors/pivot/README.md
