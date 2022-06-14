# Wavefront

The `wavefront` serializer translates the Telegraf metric format to the [Wavefront Data Format](https://docs.wavefront.com/wavefront_data_format.html).

## Configuration

```toml
[[outputs.file]]
  files = ["stdout"]

  ## Use Strict rules to sanitize metric and tag names from invalid characters
  ## When enabled forward slash (/) and comma (,) will be accepted
  # wavefront_use_strict = false

  ## point tags to use as the source name for Wavefront (if none found, host will be used)
  # wavefront_source_override = ["hostname", "address", "agent_host", "node_host"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "wavefront"
  ## Users who wish their prefix paths to not be converted may set the following:
  ## default behavior (enabled prefix/path conversion):       prod.prefix.name.metric.name
  ## configurable behavior (disabled prefix/path conversion): prod.prefix_name.metric_name
  # wavefront_disable_prefix_conversion = true
```

## Metrics

A Wavefront metric is equivalent to a single field value of a Telegraf measurement.
The Wavefront metric name will be: `<measurement_name>.<field_name>`
If a prefix is specified it will be honored.
Only boolean and numeric metrics will be serialized, all other types will generate
an error.

## Example

The following Telegraf metric

```text
cpu,cpu=cpu0,host=testHost user=12,idle=88,system=0 1234567890
```

will serialize into the following Wavefront metrics

```text
"cpu.user" 12.000000 1234567890 source="testHost" "cpu"="cpu0"
"cpu.idle" 88.000000 1234567890 source="testHost" "cpu"="cpu0"
"cpu.system" 0.000000 1234567890 source="testHost" "cpu"="cpu0"
```
