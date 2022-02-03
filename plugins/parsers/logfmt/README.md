# Logfmt

The `logfmt` data format parses data in [logfmt] format.

[logfmt]: https://brandur.org/logfmt

## Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "logfmt"
```

## Metrics

Each key/value pair in the line is added to a new metric as a field.  The type
of the field is automatically determined based on the contents of the value.

## Examples

```text
- method=GET host=example.org ts=2018-07-24T19:43:40.275Z connect=4ms service=8ms status=200 bytes=1653
+ logfmt method="GET",host="example.org",ts="2018-07-24T19:43:40.275Z",connect="4ms",service="8ms",status=200i,bytes=1653i
```
