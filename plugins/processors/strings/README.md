# Strings Processor Plugin

The `strings` plugin maps certain go string functions onto tag and field values. If `result_key` parameter is present, it can produce new tags and fields from existing ones.

Implemented functions are: Lowercase, Uppercase, Trim, TrimPrefix, TrimSuffix, TrimRight, TrimLeft

Please note that in this implementation these are processed in the order that they appear above.

Specify the `tag` or `field` that you want processed in each section and optionally a `result_key` if you want the result stored in a new tag or field. You can specify lots of transformations on data with a single strings processor. Certain functions require an `argument` field to specify how they should process their throughput.

Functions that require an `argument` are: Trim, TrimPrefix, TrimSuffix, TrimRight, TrimLeft

### Configuration:

```toml
[[processors.strings]]
  namepass = ["uri_stem"]

  # Tag and field conversions defined in a separate sub-tables
  [[processors.strings.lowercase]]
    ## Tag to change
    tag = "uri_stem"

  [[processors.strings.lowercase]]
    ## Multiple tags or fields may be defined
    tag = "method"

  [[processors.strings.uppercase]]
    key = "cs-host"
    result_key = "cs-host_normalised"

  [[processors.strings.trimprefix]]
    tag = "uri_stem"
    argument = "/api/"
```

### Tags:

No tags are applied by this processor.

### Example Input:
```
iis_log,method=get,uri_stem=/API/HealthCheck cs-host="MIXEDCASE_host",referrer="-",ident="-",http_version=1.1,agent="UserAgent",resp_bytes=270i 1519652321000000000
```
### Example Output:
```
iis_log,method=get,uri_stem=healthcheck cs-host="MIXEDCASE_host",cs-host_normalised="MIXEDCASE_HOST",referrer="-",ident="-",http_version=1.1,agent="UserAgent",resp_bytes=270i 1519652321000000000
```
