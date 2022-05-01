# Regex Processor Plugin

The `regex` plugin transforms tag and field values with regex pattern. If `result_key` parameter is present, it can produce new tags and fields from existing ones.

The regex processor **only operates on string fields**. It will not work on
any other data types, like an integer or float.

For tags transforms, if `append` is set to `true`, it will append the transformation to the existing tag value, instead of overwriting it.

For metrics transforms, `key` denotes the element that should be transformed. Furthermore, `result_key` allows control over the behavior applied in case the resulting `tag` or `field` name already exists.

## Configuration

```toml
# Transforms tag and field values as well as measurement, tag and field names with regex pattern
[[processors.regex]]
  namepass = ["nginx_requests"]

  # Tag and field conversions defined in a separate sub-tables
  [[processors.regex.tags]]
    ## Tag to change, "*" will change every tag
    key = "resp_code"
    ## Regular expression to match on a tag value
    pattern = "^(\\d)\\d\\d$"
    ## Matches of the pattern will be replaced with this string.  Use ${1}
    ## notation to use the text of the first submatch.
    replacement = "${1}xx"

  [[processors.regex.fields]]
    ## Field to change
    key = "request"
    ## All the power of the Go regular expressions available here
    ## For example, named subgroups
    pattern = "^/api(?P<method>/[\\w/]+)\\S*"
    replacement = "${method}"
    ## If result_key is present, a new field will be created
    ## instead of changing existing field
    result_key = "method"

  # Multiple conversions may be applied for one field sequentially
  # Let's extract one more value
  [[processors.regex.fields]]
    key = "request"
    pattern = ".*category=(\\w+).*"
    replacement = "${1}"
    result_key = "search_category"

  # Rename metric fields
  [[processors.regex.field_rename]]
    ## Regular expression to match on a field name
    pattern = "^search_(\\w+)d$"
    ## Matches of the pattern will be replaced with this string.  Use ${1}
    ## notation to use the text of the first submatch.
    replacement = "${1}"
    ## If the new field name already exists, you can either "overwrite" the
    ## existing one with the value of the renamed field OR you can "keep"
    ## both the existing and source field.
    # result_key = "keep"

  # Rename metric tags
  # [[processors.regex.tag_rename]]
  #   ## Regular expression to match on a tag name
  #   pattern = "^search_(\\w+)d$"
  #   ## Matches of the pattern will be replaced with this string.  Use ${1}
  #   ## notation to use the text of the first submatch.
  #   replacement = "${1}"
  #   ## If the new tag name already exists, you can either "overwrite" the
  #   ## existing one with the value of the renamed tag OR you can "keep"
  #   ## both the existing and source tag.
  #   # result_key = "keep"

  # Rename metrics
  # [[processors.regex.metric_rename]]
  #   ## Regular expression to match on an metric name
  #   pattern = "^search_(\\w+)d$"
  #   ## Matches of the pattern will be replaced with this string.  Use ${1}
  #   ## notation to use the text of the first submatch.
  #   replacement = "${1}"
```

## Tags

No tags are applied by this processor.

## Example

```text
nginx_requests,verb=GET,resp_code=2xx request="/api/search/?category=plugins&q=regex&sort=asc",method="/search/",category="plugins",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
```
