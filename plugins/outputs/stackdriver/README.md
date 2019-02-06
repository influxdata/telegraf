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

### Restrictions

Stackdriver does not support string values in custom metrics, any string
fields will not be written.

The Stackdriver API does not allow writing points which are out of order,
older than 24 hours, or more with resolution greater than than one per point
minute.  Since Telegraf writes the newest points first and moves backwards
through the metric buffer, it may not be possible to write historical data
after an interruption.

Points collected with greater than 1 minute precision may need to be
aggregated before then can be written.  Consider using the [basicstats][]
aggregator to do this.

[basicstats]: /plugins/aggregators/basicstats/README.md
