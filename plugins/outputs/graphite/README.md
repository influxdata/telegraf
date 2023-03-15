# Graphite Output Plugin

This plugin writes to [Graphite][1] via raw TCP.

For details on the translation between Telegraf Metrics and Graphite output,
see the [Graphite Data Format][2].

[1]: http://graphite.readthedocs.org/en/latest/index.html
[2]: ../../../docs/DATA_FORMATS_OUTPUT.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
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

  ## Strict sanitization regex
  ## This is the default sanitization regex that is used on data passed to the
  ## graphite serializer. Users can add additional characters here if required.
  ## Be aware that the characters, '/' '@' '*' are always replaced with '_',
  ## '..' is replaced with '.', and '\' is removed even if added to the
  ## following regex.
  # graphite_strict_sanitize_regex = '[^a-zA-Z0-9-:._=\p{L}]'

  ## Enable Graphite tags support
  # graphite_tag_support = false

  ## Applied sanitization mode when graphite tag support is enabled.
  ## * strict - uses the regex specified above
  ## * compatible - allows for greater number of characters
  # graphite_tag_sanitize_mode = "strict"

  ## Character for separating metric name and field for Graphite tags
  # graphite_separator = "."

  ## Graphite templates patterns
  ## 1. Template for cpu
  ## 2. Template for disk*
  ## 3. Default template
  # templates = [
  #  "cpu tags.measurement.host.field",
  #  "disk* measurement.field",
  #  "host.measurement.tags.field"
  #]

  ## timeout in seconds for the write connection to graphite
  timeout = 2

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```
