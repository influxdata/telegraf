# Form Urlencoded

The `form-urlencoded` data format parses `application/x-www-form-urlencoded`
data, such as commonly used in the [query string][].

A common use case is to pair it with [http_listener_v2][] input plugin to parse
request body or query params.

## Configuration

```toml
[[inputs.http_listener_v2]]
  ## Address and port to host HTTP listener on
  service_address = ":8080"

  ## Part of the request to consume.  Available options are "body" and
  ## "query".
  data_source = "body"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "form_urlencoded"

  ## Array of key names which should be collected as tags.
  ## By default, keys with string value are ignored if not marked as tags.
  form_urlencoded_tag_keys = ["tag1"]
```

## Examples

### Basic parsing

Config:

```toml
[[inputs.http_listener_v2]]
  name_override = "mymetric"
  service_address = ":8080"
  data_source = "query"
  data_format = "form_urlencoded"
  form_urlencoded_tag_keys = ["tag1"]
```

Request:

```bash
curl -i -XGET 'http://localhost:8080/telegraf?tag1=foo&field1=0.42&field2=42'
```

Output:

```text
mymetric,tag1=foo field1=0.42,field2=42
```

[query string]: https://en.wikipedia.org/wiki/Query_string
[http_listener_v2]: /plugins/inputs/http_listener_v2
