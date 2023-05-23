# Loki Output Plugin

This plugin sends logs to Loki, using metric name and tags as labels, log line
will content all fields in `key="value"` format which is easily parsable with
`logfmt` parser in Loki.

Logs within each stream are sorted by timestamp before being sent to Loki.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

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

  ## Metric Name Label
  ## Label to use for the metric name to when sending metrics. If set to an
  ## empty string, this will not add the label. This is NOT suggested as there
  ## is no way to differentiate between multiple metrics.
  # metric_name_label = "__name"
```
