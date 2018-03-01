# Override Processor Plugin

The override processor plugin allows overriding all modifications that are
supported by input plugins and aggregators:

* name_override
* name_prefix
* name_suffix
* tags

All metrics passing through this processor will be modified accordingly. Values
of *name_override*, *name_prefix*, *name_suffix* and already present *tags* with
conflicting keys will be overwritten. Absent *tags* will be created.

Use-case of this plugin encompass ensuring certain tags or naming conventions
are adhered to irrespective of input plugin configurations, e.g. by
`taginclude`.

### Configuration:

```toml
# Add a global tag to all metrics
[[processors.override]]
  name_override = "new name_override"
  name_prefix = "new name_prefix"
  name_suffix = ":new name_suffix"
  [processors.tags.add]
    additional_tag = "tag_value"
    existing_tag = "new tag_value"
```
