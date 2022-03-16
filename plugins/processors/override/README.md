# Override Processor Plugin

The override processor plugin allows overriding all modifications that are
supported by input plugins and aggregators:

* name_override
* name_prefix
* name_suffix
* tags

All metrics passing through this processor will be modified accordingly.
Select the metrics to modify using the standard
[measurement filtering](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#measurement-filtering)
options.

Values of *name_override*, *name_prefix*, *name_suffix* and already present
*tags* with conflicting keys will be overwritten. Absent *tags* will be
created.

Use-case of this plugin encompass ensuring certain tags or naming conventions
are adhered to irrespective of input plugin configurations, e.g. by
`taginclude`.

## Configuration

```toml
# Apply metric modifications using override semantics.
[[processors.override]]
  ## All modifications on inputs and aggregators can be overridden:
  # name_override = "new_name"
  # name_prefix = "new_name_prefix"
  # name_suffix = "new_name_suffix"

  ## Tags to be added (all values must be strings)
  # [processors.override.tags]
  #   additional_tag = "tag_value"
```
