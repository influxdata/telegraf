# Stackdriver Output Plugin

This plugin writes to the [Google Cloud Stackdriver API](https://cloud.google.com/monitoring/api/v3/)
and requires [authentication](https://cloud.google.com/docs/authentication/getting-started) with Google Cloud using either a service account or user credentials. See the [Stackdriver documentation](https://cloud.google.com/stackdriver/pricing#stackdriver_monitoring_services) for details on pricing.

Requires `project` to specify where Stackdriver metrics will be delivered to.

Metrics are grouped by the `namespace` variable and metric key - eg: `custom.googleapis.com/telegraf/system/load5`

### Configuration

```toml
[[outputs.stackdriver]]
  # GCP Project
  project = "erudite-bloom-151019"

  # The namespace for the metric descriptor
  namespace = "telegraf"
```
