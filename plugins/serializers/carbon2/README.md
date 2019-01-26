# Carbon2

The `carbon2` serializer translates the Telegraf metric format to the [Carbon2 format](http://metrics20.org/implementations/).

### Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "carbon2"
```

Standard form:
```
metric=name field=field_1 host=foo  30 1234567890
metric=name field=field_2 host=foo  4 1234567890
metric=name field=field_N host=foo  59 1234567890
```

### Metrics

The serializer converts the metrics by creating `intrinsic_tags` using the combination of metric name and fields.  So, if one Telegraf metric has 4 fields, the `carbon2` output will be 4 separate metrics. There will be a `metric` tag that represents the name of the metric and a `field` tag to represent the field.

### Example

If we take the following InfluxDB Line Protocol:

```
weather,location=us-midwest,season=summer temperature=82,wind=100 1234567890
```

after serializing in Carbon2, the result would be:

```
metric=weather field=temperature location=us-midwest season=summer  82 1234567890
metric=weather field=wind location=us-midwest season=summer  100 1234567890
```

### Fields and Tags with spaces
When a field key or tag key/value have spaces, spaces will be replaced with `_`.

### Tags with empty values
When a tag's value is empty, it will be replaced with `null`
