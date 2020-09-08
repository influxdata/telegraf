# Execd Output Plugin

The `execd` plugin runs an external program as a daemon.

### Configuration:

```toml
[[outputs.execd]]
  ## Program to run as daemon
  command = ["my-telegraf-output", "--some-flag", "value"]

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"

  ## Data format to export.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

### Example

see [examples][]

[examples]: https://github.com/influxdata/telegraf/blob/master/plugins/outputs/execd/examples/