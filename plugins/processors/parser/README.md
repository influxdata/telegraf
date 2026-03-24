# Parser Processor Plugin

This plugin parses defined fields or tags containing the specified
[data format][data_formats] and creates new metrics based on the resulting
fields and tags.

⭐ Telegraf v1.8.0
🏷️ transformation
💻 all

[data_formats]: /docs/DATA_FORMATS_INPUT.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Parse a value in a specified field(s)/tag(s) and add the result in a new metric
[[processors.parser]]
  ## The name of the fields whose value will be parsed.
  parse_fields = ["message"]

  ## Fields to base64 decode.
  ## These fields do not need to be specified in parse_fields.
  ## Fields specified here will have base64 decode applied to them.
  # parse_fields_base64 = []

  ## The name of the tags whose value will be parsed.
  # parse_tags = []

  ## If true, incoming metrics are not emitted.
  # drop_original = false

  ## Merge Behavior
  ## Possible options are:
  ##  - none: keep the newly parsed metrics as-is
  ##  - override: emit a single metric with all tags and fields of newly parsed
  ##    merged but retaining the first timestamp. If drop_original is
  ##    false, all metrics are merged into the original metric.
  ##    NOTE: Existing field or tag values will be overridden.
  ##  - override-with-timestamp: same as "override", but the timestamp is set
  ##    based on the new metrics if present.
  ##  - parent: emit one metric per newly parsed metric with each newly parsed
  ##    metric is merged individually into the parent metric keeping the parent
  ##    timestamp.
  ##  - parent-with-timestamp: same as "parent", but the timestamp is set
  ##    based on the new metric if present.
  # merge = "none"

  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Merge strategies

When parsing multiple metrics from a field or tag you can use the `merge`
strategy to combine the newly parsed metrics.

#### `override`

This strategy will merge all parsed metrics, i.e. the plugin will emit only one
metric containing the superset of all fields and tags of the parsed metric. If
`drop_original` is `false`, the parent metric is also merged in.

> [!IMPORTANT]
> In case identical field or tag names exist among the set of metrics those
> fields or tags will override each other and only the latest value will be
> emitted.

For example the parent metric

```text
test,source=foo message="...",additional=true 1773258782000000000
```

and parsed metrics

```text
metric,status=ok value1=1i 1773239679000000000
metric,status=warn value2=23i 1773239679100000000
metric,status=ok value3=19i 1773239679200000000
metric,status=fault value4=42i 1773239679300000000
```

will result in

```text
metric,status=fault value1=1i,value2=23i,value3=19i,value4=42i 1773258782000000000
```

with `drop_original = true`

and

```text
metric,source=foo,status=fault value1=1i,value2=23i,value3=19i,value4=42i,additional=true 1773258782000000000
```

with `drop_original = false`

#### `override-with-timestamp`

This strategy will behave the same way as `override` but will also override the
timestamp with the one of the latest parsed metric.

#### `parent`

This strategy will merge each parsed metric into its parent individually, i.e.
the plugin will emit one metric per parsed metric containing the superset of
all fields and tags of that parsed metric and its parent metric.

> [!IMPORTANT]
> In case identical field or tag names exist in a newly parsed metric and its
> parent those fields or tags will override each other and only the value of the
> parsed metric will be emitted.

For example the parent metric

```text
test,source=foo message="...",additional=true 1773258782000000000
```

and parsed metrics

```text
metric,status=ok value1=1i 1773239679000000000
metric,status=warn value2=23i 1773239679100000000
metric,status=ok value3=19i 1773239679200000000
metric,status=fault value4=42i 1773239679300000000
```

will result in

```text
metric,source=foo,status=ok value1=1i,additional=true 1773258782000000000
metric,source=foo,status=warn value2=23i,additional=true 1773258782000000000
metric,source=foo,status=ok value3=19i,additional=true 1773258782000000000
metric,source=foo,status=fault value4=42i,additional=true 1773258782000000000
```

#### `parent-with-timestamp`

This strategy will behave the same way as `parent` but will also override the
timestamp with the one of the parsed metric if it exists.

## Example

```toml
[[processors.parser]]
  parse_fields = ["message"]
  merge = "override"
  data_format = "logfmt"
```

### Input

```text
syslog,appname=influxd,facility=daemon,hostname=http://influxdb.example.org\ (influxdb.example.org),severity=info facility_code=3i,message=" ts=2018-08-09T21:01:48.137963Z lvl=info msg=\"Executing query\" log_id=09p7QbOG000 service=query query=\"SHOW DATABASES\"",procid="6629",severity_code=6i,timestamp=1533848508138040000i,version=1i
```

### Output

```text
syslog,appname=influxd,facility=daemon,hostname=http://influxdb.example.org\ (influxdb.example.org),severity=info facility_code=3i,log_id="09p7QbOG000",lvl="info",message=" ts=2018-08-09T21:01:48.137963Z lvl=info msg=\"Executing query\" log_id=09p7QbOG000 service=query query=\"SHOW DATABASES\"",msg="Executing query",procid="6629",query="SHOW DATABASES",service="query",severity_code=6i,timestamp=1533848508138040000i,ts="2018-08-09T21:01:48.137963Z",version=1i
```
