# T128 Transform Processor Plugin

The `t128_transform` transforms metrics based on the difference between two observed points.

### Configuration:

```toml
[[processors.t128_transform]]
  ## If more than this amount of time passes between data points, the
  ## previous value will be considered old and the value will be recalculated
  ## as if it hadn't been seen before. A zero expiration means never expire.
  # expiration = "0s"

  ## The operation that should be performed between two observed points.
  ## It can be 'diff' or 'rate'
  # transform = "rate"

  ## For the fields who's key/value pairs don't match, should the original
  ## field be removed?
  # remove-original = true

[processors.t128_transform.fields]
  ## Replace fields with their computed values, renaming them if indicated
  # "/rate/metric" = "/total/metric"
  # "/inline/replace" = "/inline/replace"
```

### Example Diff:

```toml
[[processors.t128_transform]]
  transform = "diff"
  remove-original = true
[processors.t128_transform.fields]
  diff = "total"
```

```diff
- measurement total=10i 1612214805000000000
- measurement total=15i 1612214810000000000
+ measurement diff=5i 1612214810000000000
```

### Example Rate:

```toml
[[processors.t128_transform]]
  transform = "rate"
  remove-original = true
[processors.t128_transform.fields]
  rate = "total"
```

```diff
- measurement total=10i 1612214805000000000
- measurement total=15i 1612214810000000000
+ measurement rate=1i 1612214810000000000
```
