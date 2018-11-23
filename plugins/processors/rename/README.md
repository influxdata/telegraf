# Rename Processor Plugin

The `rename` processor renames measurements, fields, and tags.

### Configuration:

```toml
[[processors.rename]]
  ## Specify one sub-table per rename operation.
  [[processors.rename.replace]]
    measurement = "network_interface_throughput"
    dest = "throughput"

  [[processors.rename.replace]]
    tag = "hostname"
    dest = "host"

  [[processors.rename.replace]]
    field = "lower"
    dest = "min"

  [[processors.rename.replace]]
    field = "upper"
    dest = "max"
```

### Tags:

No tags are applied by this processor, though it can alter them by renaming.

### Example processing:

```diff
- network_interface_throughput,hostname=backend.example.com lower=10i,upper=1000i,mean=500i 1502489900000000000
+ throughput,host=backend.example.com min=10i,max=1000i,mean=500i 1502489900000000000
```
