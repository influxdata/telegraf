# Exec Output Plugin

This plugin sends telegraf metrics to an external application over stdin.

The command should be defined similar to docker's `exec` form:

```text
["executable", "param1", "param2"]
```

On non-zero exit stderr will be logged at error level.

For better performance, consider execd, which runs continuously.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Send metrics to command as input over stdin
[[outputs.exec]]
  ## Command to ingest metrics via stdin.
  command = ["tee", "-a", "/dev/null"]

  ## Environment variables
  ## Array of "key=value" pairs to pass as environment variables
  ## e.g. "KEY=value", "USERNAME=John Doe",
  ## "LD_LIBRARY_PATH=/opt/custom/lib64:/usr/local/libs"
  # environment = []

  ## Timeout for command to complete.
  # timeout = "5s"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
```
