# Wavefront Output Plugin

This plugin writes to a [Wavefront](https://www.wavefront.com) instance or a
Wavefront Proxy instance over HTTP or HTTPS.

## Configuration

```toml @sample.conf
[[outputs.wavefront]]
  ## Url for Wavefront API or Wavefront Proxy instance.
  url = "https://metrics.wavefront.com"

  ## Authentication Token for Wavefront. Required if using Direct Ingestion. Not required if using a Wavefront Proxy.
  #token = "DUMMY_TOKEN"

  ## Maximum number of metrics to send per HTTP request. This value should be higher than the `metric_batch_size`. Default is 10,000. Values higher than 40,000 are not recommended.
  # http_maximum_batch_size = 10000

  ## Deprecated. DNS name of the Wavefront server or Wavefront Proxy. Use the `url` field instead.
  #host = "wavefront.example.com"

  ## Deprecated. Wavefront proxy port. Use the `url` field instead.
  #port = 2878

  ## prefix for metrics keys
  #prefix = "my.specific.prefix."

  ## whether to use "value" for name of simple fields. default is false
  #simple_fields = false

  ## character to use between metric and field name.  default is . (dot)
  #metric_separator = "."

  ## Convert metric name paths to use metricSeparator character
  ## When true will convert all _ (underscore) characters in final metric name. default is true
  #convert_paths = true

  ## Use Strict rules to sanitize metric and tag names from invalid characters
  ## When enabled forward slash (/) and comma (,) will be accepted
  #use_strict = false

  ## Use Regex to sanitize metric and tag names from invalid characters
  ## Regex is more thorough, but significantly slower. default is false
  #use_regex = false

  ## point tags to use as the source name for Wavefront (if none found, host will be used)
  #source_override = ["hostname", "address", "agent_host", "node_host"]

  ## whether to convert boolean values to numeric values, with false -> 0.0 and true -> 1.0. default is true
  #convert_bool = true

  ## Truncate metric tags to a total of 254 characters for the tag name value. Wavefront will reject any
  ## data point exceeding this limit if not truncated. Defaults to 'false' to provide backwards compatibility.
  #truncate_tags = false

  ## Flush the internal buffers after each batch. This effectively bypasses the background sending of metrics
  ## normally done by the Wavefront SDK. This can be used if you are experiencing buffer overruns. The sending
  ## of metrics will block for a longer time, but this will be handled gracefully by the internal buffering in
  ## Telegraf.
  #immediate_flush = true
```

### Convert Path & Metric Separator

If the `convert_path` option is true any `_` in metric and field names will be
converted to the `metric_separator` value.  By default, to ease metrics browsing
in the Wavefront UI, the `convert_path` option is true, and `metric_separator`
is `.` (dot).  Default integrations within Wavefront expect these values to be
set to their defaults, however if converting from another platform it may be
desirable to change these defaults.

### Use Regex

Most illegal characters in the metric name are automatically converted to `-`.
The `use_regex` setting can be used to ensure all illegal characters are
properly handled, but can lead to performance degradation.

### Source Override

Often when collecting metrics from another system, you want to use the target
system as the source, not the one running Telegraf.  Many Telegraf plugins will
identify the target source with a tag. The tag name can vary for different
plugins. The `source_override` option will use the value specified in any of the
listed tags if found. The tag names are checked in the same order as listed, and
if found, the other tags will not be checked. If no tags specified are found,
the default host tag will be used to identify the source of the metric.

### Wavefront Data format

The expected input for Wavefront is specified in the following way:

```text
<metric> <value> [<timestamp>] <source|host>=<sourceTagValue> [tagk1=tagv1 ...tagkN=tagvN]
```

More information about the Wavefront data format is available
[here](https://community.wavefront.com/docs/DOC-1031)

### Allowed values for metrics

Wavefront allows `integers` and `floats` as input values.  By default it also
maps `bool` values to numeric, false -> 0.0, true -> 1.0.  To map `strings` use
the [enum](../../processors/enum) processor plugin.
