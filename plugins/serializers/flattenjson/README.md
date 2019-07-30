# Customjson serializer

The customjson serializer allow you to have unitary metric and to customize output format for common parameter (metric_family, metric_name, metric_value,timestamp) based on default unitary metric format. To specify custom output format for common parameters, you have to specify the parameter "jmespath_expression" with JMESPath grammar expression on the Telegraf configuration file. Tags parameter can be prefixed by the parameter "tags_prefix" on the Telegraf configuration file.
jmespath_expression and tags_prefix parameter are mandatory on the Telegraf configuration file for customjson data format.

As an exemple of default unitary metric format (with jmespath_expression="" and tags_prefix="":
```javascript
{
  "metric_family": "diskio",
  "metric_name": "read_time",
  "metric_value": 14270,
  "host": "myhost",
  "name": "sdb",
  "timestamp": 1564482860000
}
```

As an exemple of default unitary metric format (with jmespath_expression="" and tags_prefix="tags":
```javascript
{
  "metric_family": "diskio",
  "metric_name": "read_time",
  "metric_value": 14270,
  "tags_host": "myhost",
  "tags_name": "sdb",
  "timestamp": 1564482860000
}
```

As an exemple of default unitary metric format (with jmespath_expression="{timestamp:timestamp,event:'metric',family_name:join('_',[metric_family,metric_name]),fields:{_value:metric_value,name:metric_name}}" and tags_prefix="tags":
```javascript
{
  "event": "metric",
  "family_name": "diskio_read_time",
  "fields": {
    "_value": 14270,
    "name": "read_time"
  },
  "tags_host": "myhost",
  "tags_name": "sdb",
  "timestamp": 1564482860000
}
```

## Using with the File output

An example configuration of a file based output with default unitary metric format:

```toml
 # Send telegraf metrics to file(s)
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["/tmp/metrics.out"]

  ## Data format to output.
  data_format = "customjson"
  jmespath_expression=""
  tags_prefix=""
```

An example configuration of a file based output with default unitary metric format and prefixed tags:

```toml
 # Send telegraf metrics to file(s)
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["/tmp/metrics.out"]

  ## Data format to output.
  data_format = "customjson"
  jmespath_expression=""
  tags_prefix="tags"
```

An example configuration of a file based output with custom unitary metric format and prefixed tags:

```toml
 # Send telegraf metrics to file(s)
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["/tmp/metrics.out"]

  ## Data format to output.
  data_format = "customjson"
  jmespath_expression="{timestamp:timestamp,event:'metric',family_name:join('_',[metric_family,metric_name]),fields:{_value:metric_value,name:metric_name}}"
  tags_prefix="tags"
```
