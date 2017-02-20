# Global Fields Processor Plugin

The global fields processor plugin simplely adds fields to every metric passing through it.

### Configuration:

```toml
# Adds fields to all metrics
[[processors.global_fields]]
  [[processors.global_fields.field]]
    Name = "owner"
    Value = "Mr T."
  [[processors.global_fields.field]]
    Name = "age"
    Value = 67
```

### Tags:

No tags are applied by this processor.
