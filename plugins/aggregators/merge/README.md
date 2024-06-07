# Merge Aggregator Plugin

Merge metrics together into a metric with multiple fields into the most memory
and network transfer efficient form.

Use this plugin when fields are split over multiple metrics, with the same
measurement, tag set and timestamp.  By merging into a single metric they can
be handled more efficiently by the output.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Merge metrics into multifield metrics by series key
[[aggregators.merge]]
  ## Precision to round the metric timestamp to
  ## This is useful for cases where metrics to merge arrive within a small
  ## interval and thus vary in timestamp. The timestamp of the resulting metric
  ## is also rounded.
  # round_timestamp_to = "1ns"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = true
```

## Example

```diff
- cpu,host=localhost usage_time=42 1567562620000000000
- cpu,host=localhost idle_time=42 1567562620000000000
+ cpu,host=localhost idle_time=42,usage_time=42 1567562620000000000
```
