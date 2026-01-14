# InfluxDB v3.x Output Plugin

This plugin writes metrics to a [InfluxDB v3.x][influxdb_v3] Core or Enterprise
instance via the HTTP API.

⭐ Telegraf v1.38.0
🏷️ datastore
💻 all

[influxdb_v3]: https://docs.influxdata.com

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Configuration for sending metrics to InfluxDB 3.x Core and Enterprise
[[outputs.influxdb_v3]]
  ## Multiple URLs can be specified but only ONE of them will be selected
  ## randomly in each interval for writing. If endpoints are unavailable another
  ## one will be used until all are exhausted or the write succeeds.
  urls = ["http://127.0.0.1:8181"]

  ## Local address to bind when connecting to the server
  ## If empty or not set, the local address is automatically chosen
  # local_address = ""

  ## Token for authentication
  token = ""

  ## Destination database to write into
  database = ""

  ## The value of this tag will be used to determine the database. If this
  ## tag is not set the 'database' option is used as the default.
  # database_tag = ""

  ## If true, the database tag will not be added to the metric
  # exclude_database_tag = false

  ## Wait for WAL persistence to complete synchronization
  ## Setting this to false reduces latency but increases the risk of data loss.
  ## See https://docs.influxdata.com/influxdb3/enterprise/write-data/http-api/v3-write-lp/#use-no_sync-for-immediate-write-responses
  # sync = true

  ## Timeout for HTTP messages
  # timeout = "5s"

  ## Enable or disable support for unsigned integer fields
  # influx_uint_support = false

  ## Omit the timestamp of the metrics when sinding to allow InfluxDB to set the
  ## timestamp of the data during ingestion. You likely want this to be false
  ## to submit the metric timestamp
  # influx_omit_timestamp = false

  ## HTTP User-Agent
  # user_agent = "telegraf"

  ## Content-Encoding for write request body, available values are "gzip",
  ## "none" and "identity"
  # content_encoding = "gzip"

  ## Additional HTTP headers
  # http_headers = {"X-Special-Header" = "Special-Value"}

  ## HTTP Proxy override, if unset values the standard proxy environment
  ## variables are consulted to determine which proxy, if any, should be used.
  # http_proxy = "http://corporate.proxy:3128"

  ## HTTP/2 Timeouts
  ## The following values control the HTTP/2 client's timeouts. These settings
  ## are generally not required unless a user is seeing issues with client
  ## disconnects. If a user does see issues, then it is suggested to set these
  ## values to "15s" for ping timeout and "30s" for read idle timeout and
  ## retry.
  ##
  ## Note that the timer for read_idle_timeout begins at the end of the last
  ## successful write and not at the beginning of the next write.
  # ping_timeout = "0s"
  # read_idle_timeout = "0s"

  ## Optional TLS Config for use on HTTP connections.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Rate limits for sending data (disabled by default)
  ## Available, uncompressed payload size e.g. "5Mb"
  # rate_limit = "unlimited"
  ## Fixed time-window for the available payload size e.g. "5m"
  # rate_limit_period = "0s"
```
