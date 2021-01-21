# Regex Processor Plugin

The `regex` plugin transforms tag and field values with regex pattern. If `result_key` parameter is present, it can produce new tags and fields from existing ones.

For tags transforms, if `append` is set to `true`, it will append the transformation to the existing tag value, instead of overwriting it.

### Configuration:

```toml
[[processors.regex]]
  namepass = ["nginx_requests"]

  # Tag and field conversions defined in a separate sub-tables
  [[processors.regex.tags]]
    ## Tag to change
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
```

### Tags:

No tags are applied by this processor.

### Example Output:
```
nginx_requests,verb=GET,resp_code=2xx request="/api/search/?category=plugins&q=regex&sort=asc",method="/search/",search_category="plugins",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
```
