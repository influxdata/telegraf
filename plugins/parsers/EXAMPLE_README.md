# Example

This description explains at a high level what the parser does and provides
links to where additional information about the format can be found.

### Configuration

This section contains the sample configuration for the parser.  Since the
configuration for a parser is not have a standalone plugin, use the `file` or
`exec` input as the base config.

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "example"

  ## Describe variables using the standard SampleConfig style.
  ##   https://github.com/influxdata/telegraf/wiki/SampleConfig
  example_option = "example_value"
```

#### example_option

If an option requires a more expansive explanation than can be included inline
in the sample configuration, it may be described here.

### Metrics

The optional Metrics section contains details about how the parser converts
input data into Telegraf metrics.

### Examples

The optional Examples section can show an example conversion from the input
format using InfluxDB Line Protocol as the reference format.

For line delimited text formats a diff may be appropriate:
```diff
- cpu|host=localhost|source=example.org|value=42
+ cpu,host=localhost,source=example.org value=42
```
