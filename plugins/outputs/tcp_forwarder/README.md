
# Tcp Forwarder Output Plugin

This plugin will send all metrics through TCP in the chosen format, this can be 
use by example with tcp listener input plugin

```toml
[[outputs.tcp_forwarder]]
  ## TCP server/endpoint to send metrics to.
  servers = ["localhost:8089"]
  ## timeout in seconds for the write connection
  timeout = 2
  ## reconnect before every push
  reconnect = false
  ## Data format to _output_.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```
