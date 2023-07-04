# VNGCLOUD vMonitor Output Plugin

This plugin writes to the [VNGCLOUD vMonitor Metrics API][metrics] and requires
`client_id` and `client_secret` which can be obtained from [IAM service][iam] .

[metrics]: https://hcm-3.console.vngcloud.vn/vmonitor/metric/info
[iam]: https://hcm-3.console.vngcloud.vn/iam/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Write metrics to Warp 10
[[outputs.vngcloud_vmonitor]]
  # Require data_format
  data_format = "vngcloud_vmonitor"
  # Require URL to VNGCLOUD vMonitor endpoint
  url = ""

  # From IAM service
  client_id = ""
  client_secret = ""

  # vMonitor query timeout
  timeout = "10s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false
```
