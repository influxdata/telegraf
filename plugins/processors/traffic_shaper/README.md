# Traffic Shaper Processor Plugin

An in-memory traffic shaper processor which evens out traffic so that
output traffic is uniform

Example of uneven traffic distribution

12:00:01 - 1000 samples received

12:00:02 - 0 samples received

...

12:00:10 - 0 samples received

by using this processor and setting samples as 100 per second
you can shape the outgoing traffic at 100 samples per second

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

  ## No of samples to be emitted in the specified rate
  # samples = 20000

  ## Rate at which the samples will be emitted
  # rate = 1s
   
  ## Buffer Size
  ## If buffer is full the incoming metrics will be dropped
  # buffer_size = 1000000
  
  ## Wait for queue to be drained before stopping
  # wait_for_drain = true
```
