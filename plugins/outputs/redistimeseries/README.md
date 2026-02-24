# Redis Time Series Output Plugin

This plugin writes metrics to a [Redis time-series][redists] server.

⭐ Telegraf v1.0.0
🏷️ datastore
💻 all

[redists]: https://redis.io/timeseries

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username` and
`password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Publishes metrics to a redis timeseries server
[[outputs.redistimeseries]]
  ## The address of the RedisTimeSeries server.
  address = "127.0.0.1:6379"

  ## Redis ACL credentials
  # username = ""
  # password = ""
  # database = 0

  ## Timeout for operations such as ping or sending metrics
  # timeout = "10s"

  ## Set a time-to-live (TTL) on each Redis key
  ## If set, Redis will expire the key after the specified duration
  ## The TTL is refreshed on every write, so the key only expires
  ## if no new data arrives within the configured period
  ## Disabled by default (no expiry)
  # expire = ""

  ## Enable attempt to convert string fields to numeric values
  ## If "false" or in case the string value cannot be converted the string
  ## field will be dropped.
  # convert_string_fields = true

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  # insecure_skip_verify = false
```
