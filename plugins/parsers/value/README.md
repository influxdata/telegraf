# Value Parser Plugin

The "value" data format translates single values into Telegraf metrics. This
is done by assigning a measurement name and setting a single field ("value")
as the parsed metric.

## Configuration

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["cat /proc/sys/kernel/random/entropy_avail"]

  ## override the default metric name of "exec"
  name_override = "entropy_available"

  ## override the field name of "value"
  # value_field_name = "value"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "value"
  data_type = "integer" # required
```

### Metric name

It is recommended to set `name_override` to a measurement name that makes sense
for your metric, otherwise it will just be set to the name of the plugin.

### Datatype

You **must** tell Telegraf what type of metric to collect by using the
`data_type` configuration option. Available options are:

- `integer`: converts the received data to an integer value. This setting will
             produce an error on non-integer data.
- `float`:   converts the received data to a floating-point value. This setting
             will treat integers as floating-point values and produces an error
             on data that cannot be converted (e.g. strings).
- `string`:  outputs the data as a string.
- `boolean`: converts the received data to a boolean value. This setting will
             produce an error on any data except for `true` and `false` strings.
- `auto_integer`: converts the received data to an integer value if possible and
                  will return the data as string otherwise. This is helpful for
                  mixed-type data.
- `auto_float`: converts the received data to a floating-point value if possible
                and will return the data as string otherwise. This is helpful
                for mixed-type data. Integer data will be treated as
                floating-point values.

**NOTE**: The `auto` conversions might convert data to their prioritized type
by accident, for example if a string data-source provides `"55"` it will be
converted to integer/float. This might break outputs that require the same
datatype within a field or column. It is thus recommended to use *strict* typing
whenever possible.
