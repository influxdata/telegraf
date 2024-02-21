# Nebius Cloud Monitoring Output Plugin

This plugin will send custom metrics to
[Nebuis Cloud Monitoring](https://nebius.com/il/services/monitoring).

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send aggregated metrics to Nebius.Cloud Monitoring
[[outputs.nebius_cloud_monitoring]]
  ## Timeout for HTTP writes.
  # timeout = "20s"

  ## Nebius.Cloud monitoring API endpoint. Normally should not be changed
  # endpoint = "https://monitoring.api.il.nebius.cloud/monitoring/v2/data/write"
```

### Authentication

This plugin currently only supports Compute metadata based authentication
in Nebius Cloud Platform.

When plugin is working inside a Compute instance it will take IAM token and
Folder ID from instance metadata. In this plugin we use [Google Cloud notation]
This internal metadata endpoint is only accessible for VMs from the cloud.

[Google Cloud notation]: https://nebius.com/il/docs/compute/operations/vm-info/get-info#gce-metadata

### Reserved Labels

Nebius Monitoring backend using json format to receive the metrics:

```json
{
  "name": "metric_name",
  "labels": {
    "key": "value",
    "foo": "bar"
  },
  "ts": "2023-06-06T11:10:50Z",
  "value": 0
}
```

But key of label cannot be `name` because it's reserved for `metric_name`.

So this payload:

```json
{
  "name": "systemd_units_load_code",
  "labels": {
    "active": "active",
    "host": "vm",
    "load": "loaded",
    "name": "accounts-daemon.service",
    "sub": "running"
  },
  "ts": "2023-06-06T11:10:50Z",
  "value": 0
}
```

will be replaced with:

```json
{
  "name": "systemd_units_load_code",
  "labels": {
    "active": "active",
    "host": "vm",
    "load": "loaded",
    "_name": "accounts-daemon.service",
    "sub": "running"
  },
  "ts": "2023-06-06T11:10:50Z",
  "value": 0
}
```
