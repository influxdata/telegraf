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
  servers = ["localhost:2003"]
  ## Prefix metrics name
  prefix = ""
  ## Graphite output template
  ## see https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  template = "host.tags.measurement.field"
  ## timeout in seconds for the write connection to graphite
  timeout = 2
```

Parameters:

    Servers  []string
    Prefix   string
    Timeout  int
    Template string

* `servers`: List of strings, ["mygraphiteserver:2003"].
* `prefix`: String use to prefix all sent metrics.
* `timeout`: Connection timeout in seconds.
* `template`: Template for graphite output format, see
https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
for more details.
