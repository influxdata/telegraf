# Yandex Cloud Monitoring

This plugin will send custom metrics to Yandex Cloud Monitoring. 
https://cloud.yandex.com/services/monitoring

### Configuration:

```toml
[[outputs.yandex_cloud_monitoring]]
  ## Timeout for HTTP writes.
  # timeout = "20s"

  ## Normally should not be changed
  # endpoint_url = "https://monitoring.api.cloud.yandex.net/monitoring/v2/data/write"

  ## Normally folder ID is taken from Compute instance metadata
  # folder_id = "..."

  ## Can be set explicitly for authentification debugging purposes 
  # iam_token = "..."  
```

### Authentication

This plugin currently support only YC.Compute metadata based authentication.

When plugin is working inside a YC.Compute instance it will take IAM token and Folder ID from instance metadata.

Other authentication methods will be added later.
