# Split Processor Plugin

This plugin splits a metric up into one or more metrics based on a template
the user provides. The timestamp of the new metric is based on the source
metric. Templates can overlap, where a field or tag, is used across templates
and as a result end up in multiple metrics.

**NOTE**: If drop original is changed to true, then the plugin can result in
dropping all metrics when no match is found! Please ensure to test
templates before putting into production *and* use metric filtering to
avoid data loss.

Some outputs are sensitive to the number of metric series that are produced.
Multiple metrics of the same series (i.e. identical name, tag key-values and
field name) with the same timestamp might result in squashing those points
to the latest metric produced.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Split a metric into one or more metrics with the specified field(s)/tag(s)
[[processors.split]]
  ## Keeps the original metric by default
  # drop_original = false

  ## Template for an output metric
  ## Users can define multiple templates to split the original metric into
  ## multiple, potentially overlapping, metrics.
  [[processors.split.template]]
    ## New metric name
    name = ""

    ## List of tag keys for this metric template, accepts globs, e.g. "*"
    tags = []

    ## List of field keys for this metric template, accepts globs, e.g. "*"
    fields = []
```

## Example

The following takes a single metric with data from two sensors and splits out
each sensor into its own metric. It also copies all tags from the original
metric to the new metric.

```toml
[[processors.split]]
  [[processors.split.metric]]
    name = "sensor1"
    tags = [ "*" ]
    fields = [ "sensor1*" ]
  [[processors.split.metric]]
    name = "sensor2"
    tags = [ "*" ]
    fields = [ "sensor2*" ]
```

```diff
-metric,status=active sensor1_channel1=4i,sensor1_channel2=2i,sensor2_channel1=1i,sensor2_channel2=2i 1684784689000000000
+sensor1,status=active sensor1_channel1=4i,sensor1_channel2=2i 1684784689000000000
+sensor2,status=active sensor2_channel1=1i,sensor2_channel2=2i 1684784689000000000
```
