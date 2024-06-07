# BasicStats Aggregator Plugin

The BasicStats aggregator plugin gives count, diff, max, min, mean,
non_negative_diff, sum, s2(variance), stdev for a set of values, emitting the
aggregate every `period` seconds.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Keep the aggregate basicstats of each metric passing through.
[[aggregators.basicstats]]
  ## The period on which to flush & clear the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Configures which basic stats to push as fields
  # stats = ["count","diff","rate","min","max","mean","non_negative_diff","non_negative_rate","percent_change","stdev","s2","sum","interval","last"]
```

- stats
  - If not specified, then `count`, `min`, `max`, `mean`, `stdev`, and `s2` are
  aggregated and pushed as fields. Other fields are not aggregated by default
  to maintain backwards compatibility.
  - If empty array, no stats are aggregated

## Measurements & Fields

- measurement1
  - field1_count
  - field1_diff (difference)
  - field1_rate (rate per second)
  - field1_max
  - field1_min
  - field1_mean
  - field1_non_negative_diff (non-negative difference)
  - field1_non_negative_rate (non-negative rate per second)
  - field1_percent_change
  - field1_sum
  - field1_s2 (variance)
  - field1_stdev (standard deviation)
  - field1_interval (interval in nanoseconds)
  - field1_last (last aggregated value)

## Tags

No tags are applied by this aggregator.

## Example Output

```text
system,host=tars load1=1 1475583980000000000
system,host=tars load1=1 1475583990000000000
system,host=tars load1_count=2,load1_diff=0,load1_rate=0,load1_max=1,load1_min=1,load1_mean=1,load1_sum=2,load1_s2=0,load1_stdev=0,load1_interval=10000000000i,load1_last=1 1475584010000000000
system,host=tars load1=1 1475584020000000000
system,host=tars load1=3 1475584030000000000
system,host=tars load1_count=2,load1_diff=2,load1_rate=0.2,load1_max=3,load1_min=1,load1_mean=2,load1_sum=4,load1_s2=2,load1_stdev=1.414162,load1_interval=10000000000i,load1_last=3 1475584010000000000
```
