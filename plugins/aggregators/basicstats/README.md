# BasicStats Aggregator Plugin

The BasicStats aggregator plugin give us count, diff, max, min, mean, sum, s2(variance), stdev and percentiles for a set of values,
emitting the aggregate every `period` seconds.

### Configuration:

```toml
# Keep the aggregate basicstats of each metric passing through.
[[aggregators.basicstats]]
  ## The period on which to flush & clear the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Configures which basic stats to push as fields. This option
  ## is deprecated and only kept for backward compatibility. If any
  ## fields is configured, this option will be ignored.
  # stats = ["count", "min", "max", "mean", "stdev", "s2", "sum"]

  ## Configures which basic stats to push as fields. "*" is the default configuration for all fields.
  ## Use strings like "p95" to add 95th percentile. Supported percentile range is [0, 100].
  # [aggregators.basicstats.fields]
  #   "*" = ["count", "min", "max", "mean", "stdev", "s2", "sum"]
  #   "some_field" = ["count", "p90", "p95"]
  ## If "*" is not specified, unmatched fields will be dropped.
  # [aggregators.basicstats.fields]
  #   "only_field" = ["count", "sum"]
```

- stats
    - Deprecated, use `fields` instead.

- fields
    - If not specified, then `count`, `min`, `max`, `mean`, `stdev`, and `s2` are aggregated and pushed as fields.  `sum` and percentiles are not aggregated by default to maintain backwards compatibility.
    - If empty array, no stats are aggregated.
    - If `"*"` not specified, unmatched fields will be dropped.

### Measurements & Fields:

- measurement1
    - field1_count
    - field1_diff (difference)
    - field1_max
    - field1_min
    - field1_mean
    - field1_non_negative_diff (non-negative difference)
    - field1_sum
    - field1_s2 (variance)
    - field1_stdev (standard deviation)
    - field1_pX (Xth percentile)
    - field1_pY (Yth percentile)

### Tags:

No tags are applied by this aggregator.

### Example Output:

```
$ telegraf --config telegraf.conf --quiet
system,host=tars load1=1 1475583980000000000
system,host=tars load1=1 1475583990000000000
system,host=tars load1_count=2,load1_diff=0,load1_max=1,load1_min=1,load1_mean=1,load1_sum=2,load1_s2=0,load1_stdev=0 1475584010000000000
system,host=tars load1=1 1475584020000000000
system,host=tars load1=3 1475584030000000000
system,host=tars load1_count=2,load1_diff=2,load1_max=3,load1_min=1,load1_mean=2,load1_sum=4,load1_s2=1,load1_stdev=1 1475584010000000000
```
