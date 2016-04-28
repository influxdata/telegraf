# TCP listener service input plugin

The TCP listener is a service input plugin that listens for messages on a TCP
socket and adds those messages to InfluxDB.
The plugin expects messages in the
[Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

### Configuration:

This is a sample configuration for the plugin.

```toml
# Generic TCP listener
[[inputs.tcp_listener]]
  ## Address and port to host TCP listener on
  service_address = ":8094"

  ## Number of TCP messages allowed to queue up. Once filled, the
  ## TCP listener will start dropping packets.
  allowed_pending_messages = 10000

  ## Maximum number of concurrent TCP connections to allow
  max_tcp_connections = 250

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```
