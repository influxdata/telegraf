# Value

The "value" data format translates single values into Telegraf metrics. This
is done by assigning a measurement name and setting a single field ("value")
as the parsed metric.

## Configuration

You **must** tell Telegraf what type of metric to collect by using the
`data_type` configuration option. Available options are:

1. integer
2. float or long
3. string
4. boolean

**Note:** It is also recommended that you set `name_override` to a measurement
name that makes sense for your metric, otherwise it will just be set to the
name of the plugin.

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
