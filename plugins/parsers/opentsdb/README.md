# OpenTSDB Telnet Style Put Format Parser Plugin

The `OpenTSDB` data format parses data in OpenTSDB's Telnet style put API
format. There are no additional configuration options for OpenTSDB. The metrics
are parsed directly into Telegraf metrics.

For more detail on the format, see:

- [OpenTSDB Telnet "PUT" API guide](http://opentsdb.net/docs/build/html/api_telnet/put.html)
- [OpenTSDB data specification](http://opentsdb.net/docs/build/html/user_guide/writing/index.html#data-specification)

## Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "opentsdb"
```

## Example

```opentsdb
put sys.cpu.user 1356998400 42.5 host=webserver01 cpu=0
```
