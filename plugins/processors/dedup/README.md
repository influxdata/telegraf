# Dedup Processor Plugin

If a metric sends the same value over successive intervals, suppress sending
the same value to the TSD until this many seconds have elapsed.  This helps
graphs over narrow time ranges still see timeseries with suppressed datapoints.

This feature can be used to reduce traffic when metric's value does not change over
time while maintain proper precision when value gets changed rapidly

### Configuration

```toml
[[processors.dedup]]
  ## Maximum time to suppress output
  dedup_interval = "600s"
```

