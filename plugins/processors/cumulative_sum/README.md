# Cumulative Sum Processor Plugin

This plugin accumulates field values per-metric over time and emit metrics with 
cumulative sums whenever a metric is updated. This is useful when using outputs 
relying on monotonically increasing values

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Accumulate field metrics if it's a numerical and add new field with summed values
[[processors.cumulative_sum]]
  ## If true, the original field will be dropped by the
  ## processor and will be removed from original metric.
  ## Defaults to true.
  # drop_original_field = true

  ## Fields to be processed (all if empty)
  # fields = []

  ## Maximum time to save accumulated value without update.original
  ## Default = 10m
  # clean_up_interval = "600s"
```

## Example

```diff
- net,host=server01 bytes_sent=1000,bytes_received=500
- net,host=server01 bytes_sent=2500,bytes_received=1500
- net,host=server01 bytes_sent=3000,bytes_received=2500
+ net,host=server01 bytes_sent_sum=6500,bytes_received_sum=4500
```
