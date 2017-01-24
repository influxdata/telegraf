# Graphite Output Plugin

This plugin writes to [Graphite](http://graphite.readthedocs.org/en/latest/index.html)
via raw TCP.

## Configuration:

```toml
# Configuration for Graphite server to send metrics to
[[outputs.graphite]]
  ## TCP endpoint for your graphite instance.
  ## If multiple endpoints are configured, the output will be load balanced.
  ## Only one of the endpoints will be written to with each iteration.
  servers = ["127.0.0.1:2003"]
  ## timeout in seconds for the write connection to graphite
  timeout = 2
  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "graphite"
  # prefix each graphite bucket, use to prefix all sent metrics.
  prefix = ""
  # Graphite output template. See
  # https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # for more details.
  template = "host.tags.measurement.field"
  # graphite protocol with plain/text or json.
  # If no value is set, plain/text is default.
  protocol = "plain/text"
```

Parameters:

    Servers  []string
    Timeout  int

* `servers`: List of strings, ["mygraphiteserver:2003"].
* `timeout`: Connection timeout in seconds.