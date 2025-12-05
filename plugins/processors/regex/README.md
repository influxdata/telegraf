# Regex Processor Plugin

This plugin transforms tag and field _values_ as well as renaming tags, fields
and metrics using regular expression patterns. Tag and field _values_ can be
transformed using named-groups in a batch fashion.

The regex processor **only operates on string fields**. It will not work on
any other data types, like an integer or float.

⭐ Telegraf v1.7.0
🏷️ transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Transforms tag and field values as well as measurement, tag and field names with regex pattern
[[processors.regex]]
  namepass = ["nginx_requests"]

  ## Tag value conversion(s). Multiple instances are allowed.
  [[processors.regex.tags]]
    ## Tag(s) to process with optional glob expressions such as '*'.
    key = "resp_code"
    ## Regular expression to match the tag value. If the value doesn't
    ## match the tag is ignored.
    pattern = "^(\\d)\\d\\d$"
    ## Replacement expression defining the value of the target tag. You can
    ## use regexp groups or named groups e.g. ${1} references the first group.
    replacement = "${1}xx"
    ## Name of the target tag defaulting to 'key' if not specified.
    ## In case of wildcards being used in `key` the currently processed
    ## tag-name is used as target.
    # result_key = "method"
    ## Appends the replacement to the target tag instead of overwriting it when
    ## set to true.
    # append = false

  ## Field value conversion(s). Multiple instances are allowed.
  [[processors.regex.fields]]
    ## Field(s) to process with optional glob expressions such as '*'.
    key = "request"
    ## Regular expression to match the field value. If the value doesn't
    ## match or the field doesn't contain a string the field is ignored.
    pattern = "^/api(?P<method>/[\\w/]+)\\S*"
    ## Replacement expression defining the value of the target field. You can
    ## use regexp groups or named groups e.g. ${method} references the group
    ## named "method".
    replacement = "${method}"
    ## Name of the target field defaulting to 'key' if not specified.
    ## In case of wildcards being used in `key` the currently processed
    ## field-name is used as target.
    # result_key = "method"

  ## Rename metric fields
  [[processors.regex.field_rename]]
    ## Regular expression to match on the field name
    pattern = "^search_(\\w+)d$"
    ## Replacement expression defining the name of the new field
    replacement = "${1}"
    ## If the new field name already exists, you can either "overwrite" the
    ## existing one with the value of the renamed field OR you can "keep"
    ## both the existing and source field.
    # result_key = "keep"

  ## Rename metric tags
  [[processors.regex.tag_rename]]
    ## Regular expression to match on a tag name
    pattern = "^search_(\\w+)d$"
    ## Replacement expression defining the name of the new tag
    replacement = "${1}"
    ## If the new tag name already exists, you can either "overwrite" the
    ## existing one with the value of the renamed tag OR you can "keep"
    ## both the existing and source tag.
    # result_key = "keep"

  ## Rename metrics
  [[processors.regex.metric_rename]]
    ## Regular expression to match on an metric name
    pattern = "^search_(\\w+)d$"
    ## Replacement expression defining the new name of the metric
    replacement = "${1}"
```

Please note, you can use multiple `tags`, `fields`, `tag_rename`, `field_rename`
and `metric_rename` sections in one processor. All of those are applied.

### Tag and field _value_ conversions

Conversions are only applied if a tag/field _name_ matches the `key` which can
contain glob statements such as `*` (asterix) _and_ the `pattern` matches the
tag/field _value_. For fields the field values has to be of type `string` to
apply the conversion. If any of the given criteria does not apply the conversion
is not applied to the metric.

The `replacement` option specifies the value of the resulting tag or field. It
can reference capturing groups by index (e.g. `${1}` being the first group) or
by name (e.g. `${mygroup}` being the group named `mygroup`).

By default, the currently processed tag or field is overwritten by the
`replacement`. To create a new tag or field you can additionally specify the
`result_key` option containing the new target tag or field name. In case the
given tag or field already exists, its value is overwritten. For `tags` you
might use the `append` flag to append the `replacement` value to an existing
tag.

### Batch processing using named groups

In `tags` and `fields` sections it is possible to use named groups to create
multiple new tags or fields respectively. To do so, _all_ capture groups have
to be named in the `pattern`. Additional non-capturing ones or other
expressions are allowed. Furthermore, neither `replacement` nor `result_key`
can be set as the resulting tag/field name is the name of the group and the
value corresponds to the group's content.

### Tag and field _name_ conversions

You can batch-rename tags and fields using the `tag_rename` and `field_rename`
sections. Contrary to the `tags` and `fields` sections, the rename operates on
the tag or field _name_, not its _value_.

A tag or field is renamed if the given `pattern` matches the name. The new name
is specified via the `replacement` option. Optionally, the `result_key` can be
set to either `overwrite` or `keep` (default) to control the behavior in case
the target tag/field already exists. For `overwrite` the target tag/field is
replaced by the source key. With this setting, the source tag/field
is removed in any case. When using the `keep` setting (default), the target
tag/field as well as the source is left unchanged and no renaming takes place.

### Metric _name_ conversions

Similar to the tag and field renaming, `metric_rename` section(s) can be used
to rename metrics matching the given `pattern`. The resulting metric name is
given via `replacement` option. If matching `pattern` the conversion is always
applied. The `result_key` option has no effect on metric renaming and shall
not be specified.

## Example

In the following examples we are using this metric

```text
nginx_requests,verb=GET,resp_code=200 request="/api/search/?category=plugins&q=regex&sort=asc",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
```

### Explicit specification

```toml
[[processors.regex]]
  namepass = ["nginx_requests"]

  [[processors.regex.tags]]
    key = "resp_code"
    pattern = "^(\\d)\\d\\d$"
    replacement = "${1}xx"

  [[processors.regex.fields]]
    key = "request"
    pattern = "^/api(?P<method>/[\\w/]+)\\S*"
    replacement = "${method}"
    result_key = "method"

  [[processors.regex.fields]]
    key = "request"
    pattern = ".*category=(\\w+).*"
    replacement = "${1}"
    result_key = "search_category"

  [[processors.regex.field_rename]]
    pattern = "^client_(\\w+)$"
    replacement = "${1}"
```

will result in

```diff
-nginx_requests,verb=GET,resp_code=200 request="/api/search/?category=plugins&q=regex&sort=asc",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
+nginx_requests,verb=GET,resp_code=2xx request="/api/search/?category=plugins&q=regex&sort=asc",method="/search/",category="plugins",referrer="-",ident="-",http_version=1.1,agent="UserAgent",ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
```

### Appending

```toml
[[processors.regex]]
  namepass = ["nginx_requests"]

  [[processors.regex.tags]]
    key = "resp_code"
    pattern = '^2\d\d$'
    replacement = " OK"
    result_key = "verb"
    append = true
```

will result in

```diff
-nginx_requests,verb=GET,resp_code=200 request="/api/search/?category=plugins&q=regex&sort=asc",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
+nginx_requests,verb=GET\ OK,resp_code=200 request="/api/search/?category=plugins&q=regex&sort=asc",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
```

### Named groups

```toml
[[processors.regex]]
  namepass = ["nginx_requests"]

  [[processors.regex.fields]]
    key = "request"
    pattern = '^/api/(?P<method>\w+)[/?].*category=(?P<category>\w+)&(?:.*)'
```

will result in

```diff
-nginx_requests,verb=GET,resp_code=200 request="/api/search/?category=plugins&q=regex&sort=asc",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
+nginx_requests,verb=GET,resp_code=200 request="/api/search/?category=plugins&q=regex&sort=asc",method="search",category="plugins",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
```

### Metric renaming

```toml
[[processors.regex]]
  [[processors.regex.metric_rename]]
    pattern = '^(\w+)_.*$'
    replacement = "${1}"
```

will result in

```diff
-nginx_requests,verb=GET,resp_code=200 request="/api/search/?category=plugins&q=regex&sort=asc",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
+nginx,verb=GET,resp_code=200 request="/api/search/?category=plugins&q=regex&sort=asc",referrer="-",ident="-",http_version=1.1,agent="UserAgent",client_ip="127.0.0.1",auth="-",resp_bytes=270i 1519652321000000000
```
