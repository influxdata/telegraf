# Rename Processor Plugin

The `rename` processor renames measurements, fields, and tags.

### Configuration:

```toml
## Measurement, tag, and field renamings are stored in separate sub-tables.
## Specify one sub-table per rename operation.
[[processors.rename]]
[[processors.rename.measurement]]
  ## measurement to change
  from = "network_interface_throughput"
  to = "throughput"

[[processors.rename.tag]]
  ## tag to change
  from = "hostname"
  to = "host"

[[processors.rename.field]]
  ## field to change
  from = "lower"
  to = "min"

[[processors.rename.field]]
  ## field to change
  from = "upper"
  to = "max"
```

### Tags:

No tags are applied by this processor, though it can alter them by renaming.

### Example processing:

```diff
- network_interface_throughput,hostname=backend.example.com,units=kbps lower=10i,upper=1000i,mean=500i 1502489900000000000
+ throughput,host=backend.example.com,units=kbps min=10i,max=1000i,mean=500i 1502489900000000000
```
