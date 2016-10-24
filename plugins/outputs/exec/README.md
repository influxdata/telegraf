# Exec Output Plugin

Please also see: [Telegraf Output Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md)

This plugin executes external programs with serialized metrics data passed on stdin, one metric per line.  
The external program is free to do what it wants with the data.

### Configuration:

```toml
[[outputs.exec]]
  # Shell/commands array
  # Full command line to executable with parameters, or a glob pattern to run all matching files.
  commands = ["/usr/local/bin/telegraf-output --config /etc/telegraf-output.conf"]
  
  # Timeout for each command to complete.
  timeout = "30s"
  
  # Data format to consume.
  data_format = "json"
```
