# Lookup Processor Plugin

The Lookup Processor allows to use one or more files containing a lookup-table
for annotating incoming metrics. The lookup is _static_ as the files are only
used on startup. The main use-case for this is to annotate metrics with
additional tags e.g. dependent on their source. Multiple tags can be added
depending on the lookup-table _files_.

The lookup key can be generated using a Golang template with the ability to
access the metric name via `{{.Name}}`, the tag values via `{{.Tag "mytag"}}`,
with `mytag` being the tag-name and field-values via `{{.Field "myfield"}}`,
with `myfield` being the field-name. Non-existing tags and field will result
in an empty string or `nil` respectively. In case the key cannot be found, the
metric is passed-trough unchanged. By default all matching tags are added and
existing tag-values are overwritten.

Please note: The plugin only supports the addition of tags and thus all mapped
tag-values need to be strings!

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

  ## Format of the lookup file(s)
  ## Available formats are:
  ##    json               -- JSON file with 'key: {tag-key: tag-value, ...}' mapping
  ##    csv_key_name_value -- CSV file with 'key,tag-key,tag-value,...,tag-key,tag-value' mapping
  ##    csv_key_values     -- CSV file with a header containing tag-names and
  ##                          rows with 'key,tag-value,...,tag-value' mappings
  # format = "json"

  ## Template for generating the lookup-key from the metric.
  ## This is a Golang template (see https://pkg.go.dev/text/template) to
  ## access the metric name (`{{.Name}}`), a tag value (`{{.Tag "name"}}`) or
  ## a field value (`{{.Field "name"}}`).
  key = '{{.Tag "host"}}'
```

## File formats

The following descriptions assume `key`s to be unique identifiers used for
matching the configured `key`. The `tag-name`/`tag-value` pairs are the tags
added to a metric if the key matches.

### `json` format

In the `json` format, the input `files` must have the following format

```json
{
  "keyA": {
    "tag-name1": "tag-value1",
    ...
    "tag-nameN": "tag-valueN",
  },
  ...
  "keyZ": {
    "tag-name1": "tag-value1",
    ...
    "tag-nameM": "tag-valueM",
  }
}
```

Please note that only _strings_ are supported for all elements.

### `csv_key_name_value` format

The `csv_key_name_value` format specifies comma-separated-value files with
the following format

```csv
# Optional comments
keyA,tag-name1,tag-value1,...,tag-nameN,tag-valueN
keyB,tag-name1,tag-value1
...
keyZ,tag-name1,tag-value1,...,tag-nameM,tag-valueM
```

The formatting uses colons (`,`) as separators and allows for comments defined
as lines starting with a hash (`#`). All lines can have different numbers but
must at least contain three columns and follow the name/value pair format, i.e.
there cannot be a name without value.

### `csv_key_values` format

This setting specifies comma-separated-value files with the following format

```csv
# Optional comments
ignored,tag-name1,...,tag-valueN
keyA,tag-value1,...,,,,
keyB,tag-value1,,,,...,
...
keyZ,tag-value1,...,tag-valueM,...,
```

The formatting uses colons (`,`) as separators and allows for comments defined
as lines starting with a hash (`#`). All lines __must__ contain the same number
of columns. The first non-comment line __must__ contain a header specifying the
tag-names. As the first column contains the key to match the first header value
is ignored. There have to be at least two columns.

Please note that empty tag-values will be ignored and the tag will not be added.

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

in `format = "json"` and a `key` of `key = '{{.Name}}-{{.Tag "host"}}'` you get

```diff
- xyzzy,host=green value=3.14 1502489900000000000
- xyzzy,host=red  value=2.71 1502499100000000000
+ xyzzy,host=green,location=eu-central,rack=C12-01 value=3.14 1502489900000000000
+ xyzzy,host=red,location=us-west,rack=C01-42 value=2.71 1502499100000000000
xyzzy,host=blue  value=6.62 1502499700000000000
```

The same results can be achieved with `format = "csv_key_name_value"` and

```csv
xyzzy-green,location,eu-central,rack,C12-01
xyzzy-red,location,us-west,rack,C01-42
```

or `format = "csv_key_values"` and

```csv
-,location,rack
xyzzy-green,eu-central,C12-01
xyzzy-red,us-west,C01-42
```
