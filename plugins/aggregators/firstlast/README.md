# FirstLast Aggregator Plugin

The firstlast aggregator emits the first and last metrics of a series.

The first metric is resent with the `first_suffix` appended to the measurement name
for all new series. The `warmup` parameter defines how long after Telegraf startup
we wait until starting to send the first metrics. This can be used to avoid a flood
of `first_` metrics if Telegraf is restarted.

The last metric is resent with the `last_suffix` appended. The `timeout` parameter defines
how long a series is idle until sending the last metric.

Assumptions for data to be aggregated properly:
 - The data should arrive in order
 - The data that arrives should not be older than "current time - `timeout`"
 - The metric should be contained in a single measurement line

### Configuration:

```toml
# Keep the first and last series
[[aggregators.firstlast]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The amount of time until a series is considered ended
  timeout = "30s"
  ## The amount of time before we start considering first series
  warmup = "20s"
  ## Emit first entry of a series
  first = true
  ## Suffix for the measurement names of first entries
  first_suffix = "_first"
  ## Emit last entry of a series
  last = true
  ## Suffix for the measurement names of last entries
  last_suffix = "_last"
```

### Measurements & Fields:

- measurement1_first
- measurement1_last

### Tags:

No tags are applied by this aggregator.

### Example Output:

```
counter_first,host=bar i=1 1554281633101153300
counter_first,host=foo i=1 1554281633099323601
counter_last,host=bar i=3 1554281635115090133
counter_last,host=foo i=3 1554281635112992012
```

Original input:
```
counter,host=bar i=1 1554281633101153300
counter,host=foo i=1 1554281633099323601
counter,host=bar i=2 1554281634107980073
counter,host=foo i=2 1554281634105931116
counter,host=bar i=3 1554281635115090133
counter,host=foo i=3 1554281635112992012
```
