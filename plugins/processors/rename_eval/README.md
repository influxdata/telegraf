# Rename Processor Plugin

The `rename` processor renames measurements, fields, and tags.

### Configuration:

```toml
[[processors.rename_eval]]
  ## Specify one sub-table per rename operation.
  # Tag to evaluate from points
  tag = "host"
  # Should original point be dropped and replaced by updated name
  dropOriginal = false
  # only takes effect if position is 0 , this is used as a replace and replaces measurement name
  dest = "cpu_host"
  # prefix=1 , postfix=2, 0=replace uses existing measurement , evaluates tag and replaces the measurement name
  position = 2
```

### Tags:

No tags are applied by this processor, though it can alter them by renaming.

### Example processing:

```diff
- http.response,service=mercury,hostname=backend.example.com lower=10i,upper=1000i,mean=500i 1502489900000000000
+ http.response_mercury,service=mercury,hostname=backend.example.com lower=10i,upper=1000i,mean=500i 1502489900000000000
```
