# Yandex Cloud Monitoring Output Plugin

This plugin will send custom metrics to [Yandex Cloud
Monitoring](https://cloud.yandex.com/services/monitoring).

## Configuration

```toml
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
