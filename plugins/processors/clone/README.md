# Clone Processor Plugin

The clone processor plugin create a copy of each metric passing through it,
preserving untouched the original metric and allowing modifications in the
copied one.

The modifications allowed are the ones supported by input plugins and
aggregators:

* name_override
* name_prefix
* name_suffix
* tags

Select the metrics to modify using the standard [metric
filtering](../../../docs/CONFIGURATION.md#metric-filtering) options.

Values of *name_override*, *name_prefix*, *name_suffix* and already present
*tags* with conflicting keys will be overwritten. Absent *tags* will be
created.

A typical use-case is gathering metrics once and cloning them to simulate
having several hosts (modifying ``host`` tag).

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Apply metric modifications using override semantics.
[[processors.clone]]
  ## All modifications on inputs and aggregators can be overridden:
  # name_override = "new_name"
  # name_prefix = "new_name_prefix"
  # name_suffix = "new_name_suffix"

  ## Tags to be added (all values must be strings)
  # [processors.clone.tags]
  #   additional_tag = "tag_value"
```
