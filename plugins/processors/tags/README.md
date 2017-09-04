# Add Global Tags Processor Plugin

The tags processor plugin adds all configured tags to every metric passing
through it. Values of already present tags with conflicting keys will be
overwritten.

While global tags can be configured in the respective section of the
configuration file those tags may be ignored when using the `taginclude`
property.
This plugin provides global tags, that are unaffected by this filtering.

### Configuration:

```toml
# Add a global tag to all metrics
[[processors.tags]]
[processors.tags.add]
  additional_tag = "tag_value"
```
