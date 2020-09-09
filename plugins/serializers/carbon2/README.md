# Carbon2

The `carbon2` serializer translates the Telegraf metric format to the [Carbon2 format](http://metrics20.org/implementations/).

## Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "carbon2"

  ## Optionally configure metrics format, whether to merge metric name and field name.
  ## Possible options:
  ## * "field_separate"
  ## * "metric_includes_field"
  ## * "" - defaults to "field_separate"
  # carbon2_format = "field_separate"
```

Standard form:

```
metric=name field=field_1 host=foo  30 1234567890
metric=name field=field_2 host=foo  4 1234567890
metric=name field=field_N host=foo  59 1234567890
```

### Metrics format

`Carbon2` serializer has a configuration option - `carbon2_format` - to change how
metrics names are being constructed.

By default `metric` will only inclue the metric name and a separate field `field`
will contain the field name.
This is the behavior of `carbon2_format = "field_separate"` which is the default
behavior (even if unspecified).

Optionally user can opt in to change this to make the metric inclue the field name
after the `_`.
This is the behavior of `carbon2_format = "metric_includes_field"` which would
make the above example look like:

```
metric=name_field_1 host=foo  30 1234567890
metric=name_field_2 host=foo  4 1234567890
metric=name_field_N host=foo  59 1234567890
```

## Metrics

The serializer converts the metrics by creating `intrinsic_tags` using the combination of metric name and fields.
So, if one Telegraf metric has 4 fields, the `carbon2` output will be 4 separate metrics.
There will be a `metric` tag that represents the name of the metric and a `field` tag to represent the field.

## Example

If we take the following InfluxDB Line Protocol:

```
weather,location=us-midwest,season=summer temperature=82,wind=100 1234567890
```

after serializing in Carbon2, the result would be:

```
metric=weather field=temperature location=us-midwest season=summer  82 1234567890
metric=weather field=wind location=us-midwest season=summer  100 1234567890
```

## Fields and Tags with spaces

When a field key or tag key/value have spaces, spaces will be replaced with `_`.

## Tags with empty values

When a tag's value is empty, it will be replaced with `null`
