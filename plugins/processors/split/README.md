# Split Processor Plugin

This plugin splits a metric up into one or more metrics based on a template
the user provides.

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
  ## Drops the original metric by default
  # keep_original = false

  ## Users can define multiple templates to generate mutliple metrics from a
  ## single metric.
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
