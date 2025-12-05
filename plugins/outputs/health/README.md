# Health Output Plugin

This plugin provides a HTTP health check endpoint that can be configured to
return failure status codes based on the value of a metric.

When the plugin is healthy it will return a 200 response; when unhealthy it
will return a 503 response. The default state is healthy, one or more checks
must fail in order for the resource to enter the failed state.

⭐ Telegraf v1.11.0
🏷️ applications
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Configurable HTTP health check resource based on metrics
[[outputs.health]]
  ## Address and port to listen on.
  ##   ex: service_address = "http://localhost:8080"
  ##       service_address = "unix:///var/run/telegraf-health.sock"
  # service_address = "http://:8080"

  ## The maximum duration for reading the entire request.
  # read_timeout = "5s"
  ## The maximum duration for writing the entire response.
  # write_timeout = "5s"

  ## Username and password to accept for HTTP basic authentication.
  # basic_username = "user1"
  # basic_password = "secret"

  ## Allowed CA certificates for client certificates.
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## TLS server certificate and private key.
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Maximum expected time between metrics being written
  ## Enforces an unhealthy state if there was no new metric seen for at least
  ## the specified time. The check is disabled by default and only used if a
  ## positive time is specified.
  # max_time_between_metrics = "0s"

  ## NOTE: Due to the way TOML is parsed, tables must be at the END of the
  ## plugin definition, otherwise additional config options are read as part of
  ## the table

  ## One or more check sub-tables should be defined, it is also recommended to
  ## use metric filtering to limit the metrics that flow into this output.
  ##
  ## When using the default buffer sizes, this example will fail when the
  ## metric buffer is half full.
  ##
  ## namepass = ["internal_write"]
  ## tagpass = { output = ["influxdb"] }
  ##
  ## [[outputs.health.compares]]
  ##   field = "buffer_size"
  ##   lt = 5000.0
  ##
  ## [[outputs.health.contains]]
  ##   field = "buffer_size"
```

### Maximum time between metrics

The health plugin can assert that metrics are being delivered to it at an
expected rate when setting `max_time_between_metrics` to a positive number.
The check measures the time between consecutive writes to the plugin and
compares it to the defined `max_time_between_metrics`. When the time
elapsed between writes is greater than the configured maximum time, the plugin
will report an unhealthy status. As soon as metrics are written again to the
plugin, the health status will reset to healthy.

Note that the metric timestamps are not taken into account, rather the time they
are written to the plugin.

### compares

The `compares` check is used to assert basic mathematical relationships.  Use
it by choosing a field key and one or more comparisons that must hold true.  If
the field is not found on a metric no comparison will be made.

Comparisons must be hold true on all metrics for the check to pass.

The available comparison operators are:

- `gt` greater than
- `ge` greater than or equal to
- `lt` less than
- `le` less than or equal to
- `eq` equal to
- `ne` not equal to

### contains

The `contains` check can be used to require a field key to exist on at least
one metric.

If the field is found on any metric the check passes.
