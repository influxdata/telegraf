# Clone Processor Plugin

The clone processor plugin create a copy of each metric passing through it,
preserving untouched the original metric and allowing modifications in the
copied one.

The modifications allowed are the ones supported by input plugins and aggregators:

* name_override
* name_prefix
* name_suffix
* tags

Select the metrics to modify using the standard
[measurement filtering](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#measurement-filtering)
options.

Values of *name_override*, *name_prefix*, *name_suffix* and already present
*tags* with conflicting keys will be overwritten. Absent *tags* will be
created.

A typical use-case is gathering metrics once and cloning them to simulate
having several hosts (modifying ``host`` tag).

## Configuration

```toml
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
