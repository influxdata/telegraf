# FormData

The FormData data format parses a [query string/x-www-form-urlencoded][query_string] data into metric fields.

Common use case is to pair it with http listener input plugin to parse request body or query params.

### Configuration

```toml
[[inputs.http_listener_v2]]
  ## Address and port to host HTTP listener on
  service_address = ":8080"  

  ## Part of the request to consume.
  ## Available options are "body" and "query".
  ## To consume standard query params or application/x-www-form-urlencoded body,
  ## set the data_format option to "formdata".
  data_source = "body"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "formdata"

  ## Array of key names which should be collected as tags.
  ## By default, keys with string value are ignored if not marked as tags.
  form_data_tag_keys = ["tag1"]
```

### Examples

#### Basic parsing
Config:
```toml
[[inputs.http_listener_v2]]
  service_address = ":8080"  
  data_source = "query"
  data_format = "formdata"
  name_override = "mymetric"
```

Request:
```bash
curl -i -XGET 'http://localhost:8080/telegraf?field=0.42'
```

Output:
```
mymetric field=0.42
```

#### Tags and key filter

Config:
```toml
[[inputs.http_listener_v2]]
  service_address = ":8080"  
  data_source = "query"
  data_format = "formdata"
  name_override = "mymetric"
  fielddrop = ["tag2", "field2"]
  form_data_tag_keys = ["tag1"]
```

Request:
```bash
curl -i -XGET 'http://localhost:8080/telegraf?tag1=foo&tag2=bar&field1=42&field2=69'
```

Output:
```
mymetric,tag1=foo field1=42
```

[query_string]: https://en.wikipedia.org/wiki/Query_string
