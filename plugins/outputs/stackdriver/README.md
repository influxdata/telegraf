# Stackdriver Google Cloud Monitoring Output Plugin

This plugin writes to the [Google Cloud Monitoring API][stackdriver] (formerly
Stackdriver) and requires [authentication][] with Google Cloud using either a
service account or user credentials

This plugin accesses APIs which are [chargeable][pricing]; you might incur
costs.

Requires `project` to specify where Stackdriver metrics will be delivered to.

By default, Metrics are grouped by the `namespace` variable and metric key -
eg: `custom.googleapis.com/telegraf/system/load5`. However, this is not the
best practice. Setting `metric_name_format = "official"` will produce a more
easily queried format of: `metric_type_prefix/[namespace_]name_key/kind`. If
the global namespace is not set, it is omitted as well.

[Resource type](https://cloud.google.com/monitoring/api/resources) is configured
by the `resource_type` variable (default `global`).

Additional resource labels can be configured by `resource_labels`. By default
the required `project_id` label is always set to the `project` variable.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Configuration for Google Cloud Stackdriver to send metrics to
[[outputs.stackdriver]]
  ## GCP Project
  project = "erudite-bloom-151019"

  ## The namespace for the metric descriptor
  ## This is optional and users are encouraged to set the namespace as a
  ## resource label instead. If omitted it is not included in the metric name.
  namespace = "telegraf"

  ## Metric Type Prefix
  ## The DNS name used with the metric type as a prefix.
  # metric_type_prefix = "custom.googleapis.com"

  ## Metric Name Format
  ## Specifies the layout of the metric name, choose from:
  ##  * path: 'metric_type_prefix_namespace_name_key'
  ##  * official: 'metric_type_prefix/namespace_name_key/kind'
  # metric_name_format = "path"

  ## Metric Data Type
  ## By default, telegraf will use whatever type the metric comes in as.
  ## However, for some use cases, forcing int64, may be preferred for values:
  ##   * source: use whatever was passed in
  ##   * double: preferred datatype to allow queries by PromQL.
  # metric_data_type = "source"

  ## Tags as resource labels
  ## Tags defined in this option, when they exist, are added as a resource
  ## label and not included as a metric label. The values from tags override
  ## the values defined under the resource_labels config options.
  # tags_as_resource_label = []

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
