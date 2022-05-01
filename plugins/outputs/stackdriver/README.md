# Stackdriver Google Cloud Monitoring Output Plugin

This plugin writes to the [Google Cloud Monitoring API][stackdriver] (formerly
Stackdriver) and requires [authentication][] with Google Cloud using either a
service account or user credentials

This plugin accesses APIs which are [chargeable][pricing]; you might incur
costs.

Requires `project` to specify where Stackdriver metrics will be delivered to.

Metrics are grouped by the `namespace` variable and metric key - eg:
`custom.googleapis.com/telegraf/system/load5`

[Resource type](https://cloud.google.com/monitoring/api/resources) is configured
by the `resource_type` variable (default `global`).

Additional resource labels can be configured by `resource_labels`. By default
the required `project_id` label is always set to the `project` variable.

## Configuration

```toml
# Configuration for Google Cloud Stackdriver to send metrics to
[[outputs.stackdriver]]
  ## GCP Project
  project = "erudite-bloom-151019"

  ## The namespace for the metric descriptor
  namespace = "telegraf"

  ## Custom resource type
  # resource_type = "generic_node"

  ## Additional resource labels
  # [outputs.stackdriver.resource_labels]
  #   node_id = "$HOSTNAME"
  #   namespace = "myapp"
  #   location = "eu-north0"
```

## Restrictions

Stackdriver does not support string values in custom metrics, any string fields
will not be written.

The Stackdriver API does not allow writing points which are out of order, older
than 24 hours, or more with resolution greater than than one per point minute.
Since Telegraf writes the newest points first and moves backwards through the
metric buffer, it may not be possible to write historical data after an
interruption.

Points collected with greater than 1 minute precision may need to be aggregated
before then can be written.  Consider using the [basicstats][] aggregator to do
this.

Histogram / distribution and delta metrics are not yet supported. These will be
dropped silently unless debugging is on.

Note that the plugin keeps an in-memory cache of the start times and last
observed values of all COUNTER metrics in order to comply with the requirements
of the stackdriver API.  This cache is not GCed: if you remove a large number of
counters from the input side, you may wish to restart telegraf to clear it.

[basicstats]: /plugins/aggregators/basicstats/README.md

[stackdriver]: https://cloud.google.com/monitoring/api/v3/

[authentication]: https://cloud.google.com/docs/authentication/getting-started

[pricing]: https://cloud.google.com/stackdriver/pricing#google-clouds-operations-suite-pricing
