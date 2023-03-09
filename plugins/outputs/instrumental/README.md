# Instrumental Output Plugin

This plugin writes to the [Instrumental Collector
API](https://instrumentalapp.com/docs/tcp-collector) and requires a
Project-specific API token.

Instrumental accepts stats in a format very close to Graphite, with the only
difference being that the type of stat (gauge, increment) is the first token,
separated from the metric itself by whitespace. The `increment` type is only
used if the metric comes in as a counter through `[[input.statsd]]`.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `api_token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Configuration for sending metrics to an Instrumental project
[[outputs.instrumental]]
  ## Project API Token (required)
  api_token = "API Token"  # required
  ## Prefix the metrics with a given name
  prefix = ""
  ## Stats output template (Graphite formatting)
  ## see https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#graphite
  template = "host.tags.measurement.field"
  ## Timeout in seconds to connect
  timeout = "2s"
  ## Debug true - Print communication to Instrumental
  debug = false
```
