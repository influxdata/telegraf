# Lookup Processor Plugin

The Lookup Processor allows to use one or more files containing a lookup-table
for annotating incoming metrics. The main use-case for this is to annotate
metrics with additional tags e.g. dependent on their source. Multiple tags can
be added depending on the lookup-table _files_.

The lookup key can be generated using a Golang template with the ability to
access the metric name via `{{.Name}}`, the tag values via `{{.Tag "mytag"}}`,
with `mytag` being the tag-name and field-values via `{{.Field "myfield"}}`,
with `myfield` being the field-name. Non-existing tags and field will result
in an empty string or `nil` respectively. In case the key cannot be found, the
metric is passed-trough unchanged. By default all matching tags are added and
existing tag-values are overwritten.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Lookup a key derived from metrics in a static file
[[processors.lookup]]
  ## List of files containing the lookup-table
  files = ["path/to/lut.json", "path/to/another_lut.json"]

  ## Template for generating the lookup-key from the metric.
  ## This is a Golang template (see https://pkg.go.dev/text/template) to
  ## access the metric name (`{{.Name}}`), a tag value (`{{.Tag "name"}}`) or
  ## a field value (`{{.Field "name"}}`).
  key = '{{.Tag "host"}}'
```

## Example

With a lookup table of

```json
{
  "xyzzy-green": {
    "location": "eu-central",
    "rack": "C12-01"
  },
  "xyzzy-red": {
    "location": "us-west",
    "rack": "C01-42"
  },
}
```

and a `key` of `key = '{{.Name}}-{{.Tag "host"}}'` you get

```diff
- xyzzy,host=green value=3.14 1502489900000000000
- xyzzy,host=red  value=2.71 1502499100000000000
+ xyzzy,host=green,location=eu-central,rack=C12-01 value=3.14 1502489900000000000
+ xyzzy,host=red,location=us-west,rack=C01-42 value=2.71 1502499100000000000
xyzzy,host=blue  value=6.62 1502499700000000000
```
