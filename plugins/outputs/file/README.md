# file, stdout, & discard Output Plugins

## file Configuration

The `file` output plugin will write metrics to file(s).

```toml
[[outputs.file]]
  ## Files to write to.
  files = ["/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

## stdout Configuration

The `stdout` output plugin will write metrics to stdout.

```toml
[[outputs.stdout]]
  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

## discard Configuration

The `discard` output plugin will write metrics to nowhere at all.

```toml
[[outputs.discard]]
  # no configuration
```
