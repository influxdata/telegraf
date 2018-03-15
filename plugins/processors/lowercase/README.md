# Lowercase Processor Plugin

The `lowercase` plugin transforms tag and field values to lower case. If `result_key` parameter is present, it can produce new tags and fields from existing ones.

### Configuration:

```toml
[[processors.lowercase]]
  namepass = ["uri_stem"]

  # Tag and field conversions defined in a separate sub-tables
  [[processors.lowercase.tags]]
    ## Tag to change
    key = "uri_stem"

  [[processors.lowercase.tags]]
    ## Multiple tags or fields may be defined
    key = "method"

  [[processors.lowercase.fields]]
    key = "cs-host"
    result_key = "cs-host_normalised"
```

### Tags:

No tags are applied by this processor.

### Example Output:
```
iis_log,method=get,uri_stem=/api/healthcheck cs-host="MIXEDCASE_host",cs-host_normalised="mixedcase_host",referrer="-",ident="-",http_version=1.1,agent="UserAgent",resp_bytes=270i 1519652321000000000
```
