# Exec Output Plugin

Please also see: [Telegraf Output Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md)

This plugin executes external programs with serialized metrics data passed on stdin, one metric per line.  
The external program is free to do what it wants with the data.

### Configuration:

```toml
[[outputs.exec]]
  # Shell/command
  # Full command line to executable with parameters.
  command = "/usr/local/bin/telegraf-output --config /etc/telegraf-output.conf"
  
  # Timeout for each command to complete.
  timeout = "30s"
  
  # Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "json"
```
