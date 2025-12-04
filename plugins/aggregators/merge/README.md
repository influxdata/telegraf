# Merge Aggregator Plugin

This plugin merges metrics of the same series and timestamp into new metrics
with the super-set of fields. A series here is defined by the metric name and
the tag key-value set.

Use this plugin when fields are split over multiple metrics, with the same
measurement, tag set and timestamp.

⭐ Telegraf v1.13.0
🏷️ transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Merge metrics into multifield metrics by series key
[[aggregators.merge]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  # period = "30s"

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
