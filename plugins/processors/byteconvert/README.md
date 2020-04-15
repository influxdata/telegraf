# Byte Convert Processor Plugin

The byte convert processor plugin takes a source metric, in bytes, and converts
it to a configured unit. The original metric is left untouched and the converted
metric is output to a new field.

### Configuration:
```toml
[[processors.byteconvert]]
  ## Name of the field to source data from.
  ##
  ## The given field should contain a measurement represented in bytes.
  field_src = "total_net_usage_bytes"

  ## Name of the new field to contain the converted value
  field_name = "total_net_usage_mb"

  ## Unit to convert the source value into.
  ##
  ## Allowed values: KiB, MiB, GiB 
  convert_unit = "MiB"
```
