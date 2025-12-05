# Rename Processor Plugin

This plugin allows to rename measurements, fields and tags.

⭐ Telegraf v1.8.0
🏷️ transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Rename measurements, tags, and fields that pass through this filter.
[[processors.rename]]
  ## Specify one sub-table per rename operation.
  [[processors.rename.replace]]
    measurement = "network_interface_throughput"
    dest = "throughput"

  [[processors.rename.replace]]
    tag = "hostname"
    dest = "host"

  [[processors.rename.replace]]
    field = "lower"
    dest = "min"

  [[processors.rename.replace]]
    field = "upper"
    dest = "max"
```

## Example

```diff
- network_interface_throughput,hostname=backend.example.com lower=10i,upper=1000i,mean=500i 1502489900000000000
+ throughput,host=backend.example.com min=10i,max=1000i,mean=500i 1502489900000000000
```
