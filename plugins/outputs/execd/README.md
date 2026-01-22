# Executable Daemon Output Plugin

This plugin writes metrics to an external daemon program via `stdin`. The
command will be executed once and metrics will be passed to it on every write
in one of the supported [data formats][data_formats].
The executable and the individual parameters must be defined as a list.

All outputs of the executable to `stderr` will be logged in the Telegraf log.
Telegraf minimum version: Telegraf 1.15.0

⭐ Telegraf v1.15.0
🏷️ system
💻 all

[data_formats]: /docs/DATA_FORMATS_OUTPUT.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Run executable as long-running output plugin
[[outputs.execd]]
  ## One program to run as daemon.
  ## NOTE: process and each argument should each be their own string
  command = ["my-telegraf-output", "--some-flag", "value"]

  ## Environment variables
  ## Array of "key=value" pairs to pass as environment variables
  ## e.g. "KEY=value", "USERNAME=John Doe",
  ## "LD_LIBRARY_PATH=/opt/custom/lib64:/usr/local/libs"
  # environment = []

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"

  ## Flag to determine whether execd should throw error when part of metrics is unserializable
  ## Setting this to true will skip the unserializable metrics and process the rest of metrics
  ## Setting this to false will throw error when encountering unserializable metrics and none will be processed
  ## This setting does not apply when use_batch_format is set.
  # ignore_serialization_error = false

  ## Use batch serialization instead of per metric. The batch format allows for the
  ## production of batch output formats and may more efficiently encode and write metrics.
  # use_batch_format = false

  ## Data format to export.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

## Error Handling

This plugin uses a fire-and-forget communication model. Metrics are considered
successfully written once they are written to the external process's `stdin`
pipe, not when the external plugin actually processes them. Due to OS pipe
buffering (typically ~64KB), writes to `stdin` are non-blocking until the
buffer fills.

If the external plugin encounters an error while processing metrics, it may
write error messages to `stderr`, which Telegraf will log. However, these
errors do not trigger Telegraf's retry mechanism or prevent metrics from being
removed from the buffer.

This means metrics can be lost if the external plugin fails to process them.
For use cases requiring guaranteed delivery, consider using a built-in output
plugin or implementing your own acknowledgment mechanism within the external
plugin.

## Example

see [examples][]

[examples]: examples/
