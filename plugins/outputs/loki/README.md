# Grafana Loki Output Plugin

This plugin writes logs to a [Grafana Loki][loki] instance, using the metric
name and tags as labels. The log line will contain all fields in
`key="value"` format easily parsable with the `logfmt` parser in Loki.

Logs within each stream are sorted by timestamp before being sent to Loki.

⭐ Telegraf v1.18.0
🏷️ logging
💻 all

[loki]: https://grafana.com/loki

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
# A plugin that can transmit logs to Loki
[[outputs.loki]]
  ## The domain of Loki
  domain = "https://loki.domain.tld"

  ## Endpoint to write api
  # endpoint = "/loki/api/v1/push"

  ## Connection timeout, defaults to "5s" if not set.
  # timeout = "5s"

  ## Basic auth credential
  # username = "loki"
  # password = "pass"

  ## Additional HTTP headers
  # http_headers = {"X-Scope-OrgID" = "1"}

  ## If the request must be gzip encoded
  # gzip_request = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Sanitize Tag Names
  ## If true, all tag names will have invalid characters replaced with
  ## underscores that do not match the regex: ^[a-zA-Z_][a-zA-Z0-9_]*.
  # sanitize_label_names = false

  ## Metric Name Label
  ## Label to use for the metric name to when sending metrics. If set to an
  ## empty string, this will not add the label. This is NOT suggested as there
  ## is no way to differentiate between multiple metrics.
  # metric_name_label = "__name"
```
