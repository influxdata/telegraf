# Rename Processor Plugin

The `rename` processor renames measurements, fields, and tags.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

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

## Tags

No tags are applied by this processor, though it can alter them by renaming.

## Example

```diff
- network_interface_throughput,hostname=backend.example.com lower=10i,upper=1000i,mean=500i 1502489900000000000
+ throughput,host=backend.example.com min=10i,max=1000i,mean=500i 1502489900000000000
```
