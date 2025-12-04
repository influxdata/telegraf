# Execd Input Plugin

This plugin runs the given external program as a long-running daemon and collects
the metrics in one of the supported [data formats][data_formats] on the
process's `stdout`. The program is expected to stay running and output data
when receiving the configured `signal`.

The `stderr` output of the process will be relayed to Telegraf's logging
facilities and will be logged as _error_ by default. However, you can log to
other levels by prefixing your message with `E!` for error, `W!` for warning,
`I!` for info, `D!` for debugging and `T!` for trace levels followed by a space
and the actual message. For example outputting `I! A log message` will create a
`info` log line in your Telegraf logging output.

⭐ Telegraf v1.14.0
🏷️ system
💻 all

[data_formats]: /docs/DATA_FORMATS_INPUT.md

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listen and wait for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Run executable as long-running input plugin
[[inputs.execd]]
  ## One program to run as daemon.
  ## NOTE: process and each argument should each be their own string
  command = ["telegraf-smartctl", "-d", "/dev/sda"]

  ## Environment variables
  ## Array of "key=value" pairs to pass as environment variables
  ## e.g. "KEY=value", "USERNAME=John Doe",
  ## "LD_LIBRARY_PATH=/opt/custom/lib64:/usr/local/libs"
  # environment = []

  ## Define how the process is signaled on each collection interval.
  ## Valid values are:
  ##   "none"    : Do not signal anything. (Recommended for service inputs)
  ##               The process must output metrics by itself.
  ##   "STDIN"   : Send a newline on STDIN. (Recommended for gather inputs)
  ##   "SIGHUP"  : Send a HUP signal. Not available on Windows. (not recommended)
  ##   "SIGUSR1" : Send a USR1 signal. Not available on Windows.
  ##   "SIGUSR2" : Send a USR2 signal. Not available on Windows.
  # signal = "none"

  ## Delay before the process is restarted after an unexpected termination
  # restart_delay = "10s"

  ## Buffer size used to read from the command output stream
  ## Optional parameter. Default is 64 Kib, minimum is 16 bytes
  # buffer_size = "64Kib"

  ## Disable automatic restart of the program and stop if the program exits
  ## with an error (non-zero error code)
  # stop_on_error = false

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"
```

## Example

See the examples directory for basic examples in different languages expecting
various signals from Telegraf:

- [Go](./examples/count.go): Example expects `signal = "SIGHUP"`
- [Python](./examples/count.py): Example expects `signal = "none"`
- [Ruby](./examples/count.rb): Example expects `signal = "none"`
- [shell](./examples/count.sh): Example expects `signal = "STDIN"`

## Metrics

Varies depending on the users data.

## Example Output

Varies depending on the users data.
