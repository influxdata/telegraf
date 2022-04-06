# Final Aggregator Plugin

The final aggregator emits the last metric of a contiguous series.  A
contiguous series is defined as a series which receives updates within the
time period in `series_timeout`. The contiguous series may be longer than the
time interval defined by `period`.

This is useful for getting the final value for data sources that produce
discrete time series such as procstat, cgroup, kubernetes etc.

When a series has not been updated within the time defined in
`series_timeout`, the last metric is emitted with the `_final` appended.

## Configuration

```toml
# Report the final metric of a series
[[aggregators.final]]
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## The time that a series is not updated until considering it final.
  series_timeout = "5m"
```

## Metrics

Measurement and tags are unchanged, fields are emitted with the suffix
`_final`.

## Example Output

```text
counter,host=bar i_final=3,j_final=6 1554281635115090133
counter,host=foo i_final=3,j_final=6 1554281635112992012
```

Original input:

```text
counter,host=bar i=1,j=4 1554281633101153300
counter,host=foo i=1,j=4 1554281633099323601
counter,host=bar i=2,j=5 1554281634107980073
counter,host=foo i=2,j=5 1554281634105931116
counter,host=bar i=3,j=6 1554281635115090133
counter,host=foo i=3,j=6 1554281635112992012
```
