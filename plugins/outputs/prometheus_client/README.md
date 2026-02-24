# Prometheus Output Plugin

This plugin starts a [Prometheus][prometheus] client and exposes the written
metrics on a `/metrics` endpoint by default. This endpoint can then be polled
by a Prometheus server.

⭐ Telegraf v0.2.1
🏷️ applications
💻 all

[prometheus]: https://prometheus.io

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `basic_password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Configuration for the Prometheus client to spawn
[[outputs.prometheus_client]]
  ## Address to listen on.
  ##   ex:
  ##     listen = ":9273"
  ##     listen = "vsock://:9273"
  listen = ":9273"

  ## Maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## Maximum duration before timing out write of the response
  # write_timeout = "10s"

  ## Metric version controls the mapping from Prometheus metrics into Telegraf metrics.
  ## See "Metric Format Configuration" in plugins/inputs/prometheus/README.md for details.
  ## Valid options: 1, 2
  # metric_version = 1

  ## Use HTTP Basic Authentication.
  # basic_username = "Foo"
  # basic_password = "Bar"

  ## If set, the IP Ranges which are allowed to access metrics.
  ##   ex: ip_range = ["192.168.0.0/24", "192.168.1.0/30"]
  # ip_range = []

  ## Path to publish the metrics on.
  # path = "/metrics"

  ## Expiration interval for each metric. 0 == no expiration
  # expiration_interval = "60s"

  ## Collectors to enable, valid entries are "gocollector" and "process".
  ## If unset, both are enabled.
  # collectors_exclude = ["gocollector", "process"]

  ## Send string metrics as Prometheus labels.
  ## Unless set to false all string metrics will be sent as labels.
  # string_as_label = true

  ## Control how metric names and label names are sanitized.
  ## The default "legacy" keeps ASCII-only Prometheus name rules.
  ## Set to "utf8" to allow UTF-8 metric and label names.
  ## Valid options: "legacy", "utf8"
  # name_sanitization = "legacy"

  ## If set, enable TLS with the given certificate.
  # tls_cert = "/etc/ssl/telegraf.crt"
  # tls_key = "/etc/ssl/telegraf.key"

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Export metric collection time.
  # export_timestamp = false

  ## Set custom headers for HTTP responses.
  # http_headers = {"X-Special-Header" = "Special-Value"}

  ## Specify the metric type explicitly.
  ## This overrides the metric-type of the Telegraf metric. Globbing is allowed.
  # [outputs.prometheus_client.metric_types]
  #   counter = []
  #   gauge = []
```

## Metrics

Prometheus metrics are produced in the same manner as the [prometheus
serializer][].

[prometheus serializer]: /plugins/serializers/prometheus/README.md#Metrics
