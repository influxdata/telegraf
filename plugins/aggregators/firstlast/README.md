# FirstLast Aggregator Plugin

The firstlast aggregator emits the first and last metrics of a contiguous series.
A contiguous series is defined as a series which receives updates within the time
period in `series_timeout`. The contiguous series may be longer than the time interval
defined by `period`.

This is useful for getting the first and/or final values for data sources that produce
discrete time series such as procstat, cgroup, kubernetes etc.

The first metric is emitted with the `_first` appended to field names for all new series
that have appeared within the time interval defined by `period`

The `warmup` parameter defines how long after Telegraf startup we wait until starting to
send the first metrics. This can be used to avoid a flood of `_first` metrics if Telegraf
is restarted.

If a series has not been updated within the time defined in `series_timeout`, the last metric 
is emttied with the `_last` appended.

Assumptions for data to be aggregated properly:

- The data should arrive in order
- The data that arrives should not be older than "current time - `series_timeout`"
- The metric should be contained in a single measurement line

### Configuration

```toml
# Keep the first and/or last metrics of a contiguous series
[[aggregators.firstlast]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The time that a series is not updated until considering it ended
  series_timeout = "30s"
  ## The amount of time to wait after Telegraf startup until evaluating new series
  warmup = "10s"
  ## Emit first entry of a series
  first = true
  ## Emit last entry of a series
  last = true
```

### Measurements & Fields

- measurement1
  - field1_first
  - field1_last

### Tags

No tags are applied by this aggregator.

### Example Output

```
counter,host=bar i_first=1 1554281633101153300
counter,host=foo i_first=1 1554281633099323601
counter,host=bar i_last=3 1554281635115090133
counter,host=foo i_last=3 1554281635112992012
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
