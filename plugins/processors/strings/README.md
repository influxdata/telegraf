# Strings Processor Plugin

The `strings` plugin maps certain go string functions onto measurement, tag, and field values.  Values can be modified in place or stored in another key.

Implemented functions are:
- lowercase
- uppercase
- trim
- trim_left
- trim_right
- trim_prefix
- trim_suffix

Please note that in this implementation these are processed in the order that they appear above.

Specify the `measurement`, `tag` or `field` that you want processed in each section and optionally a `dest` if you want the result stored in a new tag or field. You can specify lots of transformations on data with a single strings processor.

### Configuration:

```toml
[[processors.strings]]
  # [[processors.strings.uppercase]]
  #   tag = "method"

  # [[processors.strings.lowercase]]
  #   field = "uri_stem"
  #   dest = "uri_stem_normalised"

  ## Convert a tag value to lowercase
  # [[processors.strings.trim]]
  #   field = "message"

  # [[processors.strings.trim_left]]
  #   field = "message"
  #   cutset = "\t"

  # [[processors.strings.trim_right]]
  #   field = "message"
  #   cutset = "\r\n"

  # [[processors.strings.trim_prefix]]
  #   field = "my_value"
  #   prefix = "my_"

  # [[processors.strings.trim_suffix]]
  #   field = "read_count"
  #   suffix = "_count"
```

#### Trim, TrimLeft, TrimRight

The `trim`, `trim_left`, and `trim_right` functions take an optional parameter: `cutset`.  This value is a string containing the characters to remove from the value.

#### TrimPrefix, TrimSuffix

The `trim_prefix` and `trim_suffix` functions remote the given `prefix` or `suffix`
respectively from the string.

### Example
**Config**
```toml
[[processors.strings]]
  [[processors.strings.lowercase]]
    field = "uri-stem"

  [[processors.strings.trim_prefix]]
    field = "uri_stem"
    prefix = "/api/"

  [[processors.strings.uppercase]]
    field = "cs-host"
    dest = "cs-host_normalised"
```

**Input**
```
iis_log,method=get,uri_stem=/API/HealthCheck cs-host="MIXEDCASE_host",referrer="-",ident="-",http_version=1.1,agent="UserAgent",resp_bytes=270i 1519652321000000000
```

**Output**
```
iis_log,method=get,uri_stem=healthcheck cs-host="MIXEDCASE_host",cs-host_normalised="MIXEDCASE_HOST",referrer="-",ident="-",http_version=1.1,agent="UserAgent",resp_bytes=270i 1519652321000000000
```
