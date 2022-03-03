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

  ## Influx parser type to use. Users can choose between 'internal' and
  ## 'upstream'. The internal parser is what Telegraf has historically used.
  ## While the upstream parser involved a large re-write to make it more
  ## memory efficient and performant.
  ## influx_parser_version = "internal"
```
