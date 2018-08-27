# Enum Processor Plugin

The Enum Processor allows the configuration of value mappings for metric fields.
The main use-case for this is to rewrite status codes such as _red_, _amber_ and
_green_ by numeric values such as 0, 1, 2. The plugin supports string and bool
types for the field values. Multiple Fields can be configured with separate
value mappings for each field. Default mapping values can be configured to be
used for all values, which are not contained in the value_mappings. The
processor supports explicit configuration of a destination field. By default the
source field is overwritten.

### Configuration:

```toml
[[processors.enum]]
  [[processors.enum.fields]]
    ## Name of the field to map
    source = "name"

    ## Destination field to be used for the mapped value.  By default the source
    ## field is used, overwriting the original value.
    # destination = "mapped"

    ## Default value to be used for all values not contained in the mapping
    ## table.  When unset, the unmodified value for the field will be used if no
    ## match is found.
    # default = 0

    ## Table of mappings
    [processors.enum.fields.value_mappings]
      value1 = 1
      value2 = 2
```
