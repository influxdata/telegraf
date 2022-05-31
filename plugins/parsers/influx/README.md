# Influx Line Protocol

Parses metrics using the [Influx Line Protocol][].

[Influx Line Protocol]: https://docs.influxdata.com/influxdb/latest/reference/syntax/line-protocol/

## Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Influx line protocol parser
  ## 'internal' is the default. 'upstream' is a newer parser that is faster
  ## and more memory efficient.
  ## influx_parser_type = "internal"
```
