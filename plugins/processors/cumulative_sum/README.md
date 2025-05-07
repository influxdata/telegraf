# Cumulative Sum Processor Plugin

This plugin accumulates field values per-metric over time and emit metrics with 
cumulative sums whenever a metric is updated. This is useful when using outputs 
relying on monotonically increasing values

> [!NOTE]
> Metrics within a series are accumulated in the **order of arrival** and not in 
> order of their timestamps!

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Compute the cumulative sum of the given fields
[[processors.cumulative_sum]]
  ## Numerical fields to be processed (accepting wildcards)
  # fields = ["*"]

  ## Interval after which metrics are evicted from the cache and the
  ## sum values are reset to zero. A zero or unset value will keep the
  ## metric forever.
  ## It is strongly recommended to set an expiry interval to avoid
  ## growing memory usage when varying metric series are processed.
  # expiry_interval = "0s"
```

## Example

```diff
- net,host=server01 bytes_sent=1000,bytes_received=500
- net,host=server01 bytes_sent=2500,bytes_received=1500
- net,host=server01 bytes_sent=3000,bytes_received=2500
+ net,host=server01 bytes_sent=1000,bytes_sent_sum=1000,bytes_received=500,bytes_received_sum=500
+ net,host=server01 bytes_sent=2500,bytes_sent_sum=3500,bytes_received=1500,bytes_received_sum=2000
+ net,host=server01 bytes_sent=3000,bytes_sent_sum=6500,bytes_received=2500,bytes_received_sum=4500
```
