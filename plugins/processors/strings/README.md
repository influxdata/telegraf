# Strings Processor Plugin

The `strings` plugin maps certain go string functions onto measurement, tag, and field values.  Values can be modified in place or stored in another key.

Implemented functions are:

- lowercase
- uppercase
- titlecase
- trim
- trim_left
- trim_right
- trim_prefix
- trim_suffix
- replace
- left
- base64decode
- valid_utf8

Please note that in this implementation these are processed in the order that they appear above.

Specify the `measurement`, `tag`, `tag_key`, `field`, or `field_key` that you want processed in each section and optionally a `dest` if you want the result stored in a new tag or field. You can specify lots of transformations on data with a single strings processor.

If you'd like to apply the change to every `tag`, `tag_key`, `field`, `field_key`, or `measurement`, use the value `"*"` for each respective field. Note that the `dest` field will be ignored if `"*"` is used.

If you'd like to apply multiple processings to the same `tag_key` or `field_key`, note the process order stated above. See the second example below for an example.

## Configuration

```toml
# Perform string processing on tags, fields, and measurements
[[processors.strings]]
  ## Convert a field value to lowercase and store in a new field
  # [[processors.strings.lowercase]]
  #   field = "uri_stem"
  #   dest = "uri_stem_normalised"

  ## Convert a tag value to uppercase
  # [[processors.strings.uppercase]]
  #   tag = "method"

  ## Convert a field value to titlecase
  # [[processors.strings.titlecase]]
  #   field = "status"

  ## Trim leading and trailing whitespace using the default cutset
  # [[processors.strings.trim]]
  #   field = "message"

  ## Trim leading characters in cutset
  # [[processors.strings.trim_left]]
  #   field = "message"
  #   cutset = "\t"

  ## Trim trailing characters in cutset
  # [[processors.strings.trim_right]]
  #   field = "message"
  #   cutset = "\r\n"

  ## Trim the given prefix from the field
  # [[processors.strings.trim_prefix]]
  #   field = "my_value"
  #   prefix = "my_"

  ## Trim the given suffix from the field
  # [[processors.strings.trim_suffix]]
  #   field = "read_count"
  #   suffix = "_count"

  ## Replace all non-overlapping instances of old with new
  # [[processors.strings.replace]]
  #   measurement = "*"
  #   old = ":"
  #   new = "_"

  ## Trims strings based on width
  # [[processors.strings.left]]
  #   field = "message"
  #   width = 10

  ## Decode a base64 encoded utf-8 string
  # [[processors.strings.base64decode]]
  #   field = "message"

  ## Sanitize a string to ensure it is a valid utf-8 string
  ## Each run of invalid UTF-8 byte sequences is replaced by the replacement string, which may be empty
  # [[processors.strings.valid_utf8]]
  #   field = "message"
  #   replacement = ""
```

### Trim, TrimLeft, TrimRight

The `trim`, `trim_left`, and `trim_right` functions take an optional parameter: `cutset`.  This value is a string containing the characters to remove from the value.

### TrimPrefix, TrimSuffix

The `trim_prefix` and `trim_suffix` functions remote the given `prefix` or `suffix`
respectively from the string.

### Replace

The `replace` function does a substring replacement across the entire
string to allow for different conventions between various input and output
plugins. Some example usages are eliminating disallowed characters in
field names or replacing separators between different separators.
Can also be used to eliminate unneeded chars that were in metrics.
If the entire name would be deleted, it will refuse to perform
the operation and keep the old name.

## Example

A sample configuration:

```toml
[[processors.strings]]
  [[processors.strings.lowercase]]
    tag = "uri_stem"

  [[processors.strings.trim_prefix]]
    tag = "uri_stem"
    prefix = "/api/"

  [[processors.strings.uppercase]]
    field = "cs-host"
    dest = "cs-host_normalised"
```

Sample input:

```text
iis_log,method=get,uri_stem=/API/HealthCheck cs-host="MIXEDCASE_host",http_version=1.1 1519652321000000000
```

Sample output:

```text
iis_log,method=get,uri_stem=healthcheck cs-host="MIXEDCASE_host",http_version=1.1,cs-host_normalised="MIXEDCASE_HOST" 1519652321000000000
```

### Second Example

A sample configuration:

```toml
[[processors.strings]]
  [[processors.strings.lowercase]]
    tag_key = "URI-Stem"

  [[processors.strings.replace]]
    tag_key = "uri-stem"
    old = "-"
    new = "_"
```

Sample input:

```text
iis_log,URI-Stem=/API/HealthCheck http_version=1.1 1519652321000000000
```

Sample output:

```text
iis_log,uri_stem=/API/HealthCheck http_version=1.1 1519652321000000000
```
