# GrayLog Input Plugin

This plugin collects data from [Graylog servers][graylog], currently supporting
two type of end points `multiple`
(e.g. `http://<host>:9000/api/system/metrics/multiple`) and `namespace`
(e.g. `http://<host>:9000/api/system/metrics/namespace/{namespace}`).

Multiple endpoint can be queried and mixing `multiple` and serveral `namespace`
end points is possible. Check `http://<host>:9000/api/api-browser` for the full
list of available endpoints.

> [!NOTE]
> When specifying a `namespace` endpoint without an actual namespace, the
> metrics array will be ignored.

⭐ Telegraf v1.0.0
🏷️ logging
💻 all

[graylog]: https://graylog.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read flattened metrics from one or more GrayLog HTTP endpoints
[[inputs.graylog]]
  ## API endpoint, currently supported API:
  ##
  ##   - multiple  (e.g. http://<host>:9000/api/system/metrics/multiple)
  ##   - namespace (e.g. http://<host>:9000/api/system/metrics/namespace/{namespace})
  ##
  ## For namespace endpoint, the metrics array will be ignored for that call.
  ## Endpoint can contain namespace and multiple type calls.
  ##
  ## Please check http://[graylog-server-ip]:9000/api/api-browser for full list
  ## of endpoints
  servers = [
    "http://[graylog-server-ip]:9000/api/system/metrics/multiple",
  ]

  ## Set timeout (default 5 seconds)
  # timeout = "5s"

  ## Metrics list
  ## List of metrics can be found on Graylog webservice documentation.
  ## Or by hitting the web service api at:
  ##   http://[graylog-host]:9000/api/system/metrics
  metrics = [
    "jvm.cl.loaded",
    "jvm.memory.pools.Metaspace.committed"
  ]

  ## Username and password
  username = ""
  password = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

Please refer to GrayLog metrics API browser for full metric end points:
`http://host:9000/api/api-browser`

## Metrics

## Example Output
