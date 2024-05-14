# Traffic Shaper Processor Plugin

An in-memory traffic shaper processor which evens out traffic so that 
output traffic is uniform

Example of uneven traffic distribution
![traffic_distribution](./docs/traffic_distribution.png)
After applying traffic shaper the output traffic distribution is uniform

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Traffic Shaper outputs metrics at an uniform rate
[[processors.traffic_shaper]]

  ## No of samples to be emitted per time unit, default is seconds
  ## This should be used in conjunction with number of telegraf instances.
  samples = 20000

  ## Buffer Size
  ## If buffer is full the incoming metrics will be dropped
  buffer_size = 1000000
```
