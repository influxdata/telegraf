# Exec Output Plugin

This plugin sends telegraf metrics to an external application over stdin.

The command should be defined similar to docker's `exec` form:

    ["executable", "param1", "param2"]

On non-zero exit stderr will be logged at error level.

### Configuration

```toml
[[outputs.exec]]
  ## Command to ingest metrics via stdin.
  command = ["tee", "-a", "/dev/null"]

  ## Timeout for command to complete.
  # timeout = "5s"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
```
