# Dedup Processor Plugin

Filter metrics whose field values are exact repetitions of the previous values.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Filter metrics with repeating field values
[[processors.dedup]]
  ## Maximum time to suppress output
  dedup_interval = "600s"
```

## Example

```diff
- cpu,cpu=cpu0 time_idle=42i,time_guest=1i
- cpu,cpu=cpu0 time_idle=42i,time_guest=2i
- cpu,cpu=cpu0 time_idle=42i,time_guest=2i
- cpu,cpu=cpu0 time_idle=44i,time_guest=2i
- cpu,cpu=cpu0 time_idle=44i,time_guest=2i
+ cpu,cpu=cpu0 time_idle=42i,time_guest=1i
+ cpu,cpu=cpu0 time_idle=42i,time_guest=2i
+ cpu,cpu=cpu0 time_idle=44i,time_guest=2i
```
