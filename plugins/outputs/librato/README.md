# Librato Output Plugin

This plugin writes metrics to the [Librato][librato] service. It requires an
`api_user` and `api_token` which can be obtained on the [website][tokens] for
your account.

The `source_tag` option in the Configuration file is used to send contextual
information from Point Tags to the API. Besides from this, the plugin currently
does not send any additional associated Point Tags.

> [!IMPOTANT]
> If the point value being sent cannot be converted to a `float64`, the metric
> is skipped.

⭐ Telegraf v0.2.0
🏷️ cloud, datastore
💻 all

[librato]: https://www.librato.com/
[tokens]: https://metrics.librato.com/account/api_tokens

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `api_user` and
`api_token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Configuration for Librato API to send metrics to.
[[outputs.librato]]
  ## Librato API Docs
  ## http://dev.librato.com/v1/metrics-authentication
  ## Librato API user
  api_user = "telegraf@influxdb.com" # required.
  ## Librato API token
  api_token = "my-secret-token" # required.
  ## Debug
  # debug = false
  ## Connection timeout.
  # timeout = "5s"
  ## Output source Template (same as graphite buckets)
  ## see https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#graphite
  ## This template is used in librato's source (not metric's name)
  template = "host"
```
