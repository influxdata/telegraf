# Yandex Cloud Monitoring Output Plugin

This plugin will send custom metrics to [Yandex Cloud
Monitoring](https://cloud.yandex.com/services/monitoring).

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Send aggregated metrics to Yandex.Cloud Monitoring
[[outputs.yandex_cloud_monitoring]]
  ## Timeout for HTTP writes.
  # timeout = "20s"

  ## Yandex.Cloud monitoring API endpoint. Normally should not be changed
  # endpoint_url = "https://monitoring.api.cloud.yandex.net/monitoring/v2/data/write"

  ## All user metrics should be sent with "custom" service specified. Normally should not be changed
  # service = "custom"
```

### Authentication

This plugin currently support only YC.Compute metadata based authentication.

When plugin is working inside a YC.Compute instance it will take IAM token and
Folder ID from instance metadata.

Other authentication methods will be added later.
