# Logz.io Output Plugin

This plugin writes metrics to the [Logz.io][logzio] service via HTTP.

⭐ Telegraf v1.17.0
🏷️ cloud, datastore
💻 all

[logzio]: https://logz.io

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
# A plugin that can send metrics over HTTPs to Logz.io
[[outputs.logzio]]
  ## Connection timeout, defaults to "5s" if not set.
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Logz.io account token
  token = "your logz.io token" # required

  ## Use your listener URL for your Logz.io account region.
  # url = "https://listener.logz.io:8071"
```

### Required parameters

Your Logz.io `token`, which can be found under "settings" in your account, is
required.

### Optional parameters

- `check_disk_space`: Set to true if Logz.io sender checks the disk space before
                      adding metrics to the disk queue.
- `disk_threshold`: If the queue_dir space crosses this threshold
                    (in % of disk usage), the plugin will start dropping logs.
- `drain_duration`: Time to sleep between sending attempts.
- `queue_dir`: Metrics disk path. All the unsent metrics are saved to the disk
               in this location.
- `url`: Logz.io listener URL.
