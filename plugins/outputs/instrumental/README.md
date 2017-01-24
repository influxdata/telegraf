# Instrumental Output Plugin

This plugin writes to the [Instrumental Collector API](https://instrumentalapp.com/docs/tcp-collector)
and requires a Project-specific API token.

Instrumental accepts stats in a format very close to Graphite, with the only difference being that
the type of stat (gauge, increment) is the first token, separated from the metric itself
by whitespace. The `increment` type is only used if the metric comes in as a counter through `[[input.statsd]]`.

## Configuration:

```toml
[[outputs.instrumental]]
  ## Project API Token (required)
  api_token = "API Token"  # required
  ## Timeout in seconds to connect
  timeout = "2s"
  ## Debug true - Print communcation to Instrumental
  debug = false
  
  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "graphite"
  # prefix each graphite bucket
  prefix = ""
  # Graphite output template
  template = "host.tags.measurement.field"
  # graphite protocol with plain/text or json.
  # If no value is set, plain/text is default.
  protocol = "plain/text"
```
