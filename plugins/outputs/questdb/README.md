# QuestDB Output Plugin

This plugin sends metrics to [QuestDB](https://questdb.io) and automatically create the target tables when they
don't already exist. The plugin supports authentication and TLS.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# QuestDB writer, supports authentication
[[outputs.questdb]]
  ## URL to connect to
  # address = "tcp://127.0.0.1:9009"

  ## Optional authentication
  # user = "admin"
  # token = "M34uWQbu_eKO3S5NnFOBXq3u43nXHwHfU34hykhDdEY"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Period between keep alive probes.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"
```
