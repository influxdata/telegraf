# Enum Processor Plugin

The Enum Processor allows the configuration of value mappings for metric fields.
The main use-case for this is to rewrite status codes such as _red_, _amber_ and
_green_ by numeric values such as 0, 1, 2. The plugin supports string and bool
types for the field values. Multiple Fields can be configured with separate
value mappings for each field. Default mapping values can be configured to be
used for all values, which are not contained in the value_mappings. The
processor supports explicit configuration of a destination field. By default the
source field is overwritten.

### Configuration
Configuration using table syntax:
`toml
# Configure a status mapping for field 'status'
[[processors.enum.fields]]
  source = "status"
  destination = "code"
  default = -1
  [processors.enum.fields.value_mappings]
    green = 0
    yellow = 1
    red = 2
`

Configuration using inline syntax:
`toml
# Configure a status mapping for field 'status'
[[processors.enum.fields]]
  source = "status"
  destination = "code"
  default = -1
  value_mappings = {green = 0, yellow = 1, red = 2 }
`
