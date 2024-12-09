# Riemann Listener Input Plugin

The Riemann Listener is a simple input plugin that listens for messages from
client that use riemann clients using riemann-protobuff format.

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Riemann protobuff listener
[[inputs.riemann_listener]]
  ## URL to listen on
  ## Default is "tcp://:5555"
  #  service_address = "tcp://:8094"
  #  service_address = "tcp://127.0.0.1:http"
  #  service_address = "tcp4://:8094"
  #  service_address = "tcp6://:8094"
  #  service_address = "tcp6://[2001:db8::1]:8094"

  ## Maximum number of concurrent connections.
  ## 0 (default) is unlimited.
  #  max_connections = 1024
  ## Read timeout.
  ## 0 (default) is unlimited.
  #  read_timeout = "30s"
  ## Optional TLS configuration.
  #  tls_cert = "/etc/telegraf/cert.pem"
  #  tls_key  = "/etc/telegraf/key.pem"
  ## Enables client authentication if set.
  #  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]
  ## Maximum socket buffer size (in bytes when no unit specified).
  #  read_buffer_size = "64KiB"
  ## Period between keep alive probes.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  #  keep_alive_period = "5m"
```

Just like Riemann the default port is 5555. This can be configured, refer
configuration above.

Riemann `Service` is mapped as `measurement`. `metric` and `TTL` are converted
into field values.  As Riemann tags as simply an array, they are converted into
the `influx_line` format key-value, where both key and value are the tags.

## Metrics

## Example Output
