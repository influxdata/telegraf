# Logz.io Output Plugin

This plugin sends metrics to Logz.io over HTTPs.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

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

* `token`: Your Logz.io token, which can be found under "settings" in your account.

### Optional parameters

* `check_disk_space`: Set to true if Logz.io sender checks the disk space before adding metrics to the disk queue.
* `disk_threshold`: If the queue_dir space crosses this threshold (in % of disk usage), the plugin will start dropping logs.
* `drain_duration`: Time to sleep between sending attempts.
* `queue_dir`: Metrics disk path. All the unsent metrics are saved to the disk in this location.
* `url`: Logz.io listener URL.
