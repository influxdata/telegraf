# Telegraf Input Data Formats

Telegraf metrics, like InfluxDB
[points](https://docs.influxdata.com/influxdb/v0.10/write_protocols/line/),
are a combination of four basic parts:

1. Measurement Name
1. Tags
1. Fields
1. Timestamp

These four parts are easily defined when using InfluxDB line-protocol as a
data format. But there are other data formats that users may want to use which
require more advanced configuration to create usable Telegraf metrics.

Plugins such as `exec` and `kafka_consumer` parse textual data. Up until now,
these plugins were statically configured to parse just a single
data format. `exec` mostly only supported parsing JSON, and `kafka_consumer` only
supported data in InfluxDB line-protocol.

But now we are normalizing the parsing of various data formats across all
plugins that can support it. You will be able to identify a plugin that supports
different data formats by the presence of a `data_format` config option, for
example, in the exec plugin:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/tmp/test.sh", "/usr/bin/mycollector --foo=bar"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume. This can be "json", "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"

  ## Additional configuration options go here
```

Each data_format has an additional set of configuration options available, which
I'll go over below.

## Influx:

There are no additional configuration options for InfluxDB line-protocol. The
metrics are parsed directly into Telegraf metrics.

#### Influx Configuration:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/tmp/test.sh", "/usr/bin/mycollector --foo=bar"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume. This can be "json", "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## JSON:

The JSON data format flattens JSON into metric _fields_. For example, this JSON:

```json
{
    "a": 5,
    "b": {
        "c": 6
    }
}
```

Would get translated into _fields_ of a measurement:

```
myjsonmetric a=5,b_c=6
```

The _measurement_ _name_ is usually the name of the plugin,
but can be overridden using the `name_override` config option.

#### JSON Configuration:

The JSON data format supports specifying "tag keys". If specified, keys
will be searched for in the root-level of the JSON blob. If the key(s) exist,
they will be applied as tags to the Telegraf metrics.

For example, if you had this configuration:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/tmp/test.sh", "/usr/bin/mycollector --foo=bar"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume. This can be "json", "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"

  ## List of tag names to extract from top-level of JSON server response
  tag_keys = [
    "my_tag_1",
    "my_tag_2"
  ]
```

with this JSON output from a command:

```json
{
    "a": 5,
    "b": {
        "c": 6
    },
    "my_tag_1": "foo"
}
```

Your Telegraf metrics would get tagged with "my_tag_1"

```
exec_mycollector,my_tag_1=foo a=5,b_c=6
```

## Graphite:

The Graphite data format translates graphite _dot_ buckets directly into
telegraf measurement names, with a single value field, and without any tags. For
more advanced options, Telegraf supports specifying "templates" to translate
graphite buckets into Telegraf metrics.

#### Separator:

You can specify a separator to use for the parsed metrics.
By default, it will leave the metrics with a "." separator.
Setting `separator = "_"` will translate:

```
cpu.usage.idle 99
=> cpu_usage_idle value=99
```

#### Measurement/Tag Templates:

The most basic template is to specify a single transformation to apply to all
incoming metrics. _measurement_ is a special keyword that tells Telegraf which
parts of the graphite bucket to combine into the measurement name. It can have a
trailing `*` to indicate that the remainder of the metric should be used.
Other words are considered tag keys. So the following template:

```toml
templates = [
    "region.measurement*"
]
```

would result in the following Graphite -> Telegraf transformation.

```
us-west.cpu.load 100
=> cpu.load,region=us-west value=100
```

#### Field Templates:

There is also a _field_ keyword, which can only be specified once.
The field keyword tells Telegraf to give the metric that field name.
So the following template:

```toml
templates = [
    "measurement.measurement.field.region"
]
```

would result in the following Graphite -> Telegraf transformation.

```
cpu.usage.idle.us-west 100
=> cpu_usage,region=us-west idle=100
```

#### Filter Templates:

Users can also filter the template(s) to use based on the name of the bucket,
using glob matching, like so:

```toml
templates = [
    "cpu.* measurement.measurement.region",
    "mem.* measurement.measurement.host"
]
```

which would result in the following transformation:

```
cpu.load.us-west 100
=> cpu_load,region=us-west value=100

mem.cached.localhost 256
=> mem_cached,host=localhost value=256
```

#### Adding Tags:

Additional tags can be added to a metric that don't exist on the received metric.
You can add additional tags by specifying them after the pattern.
Tags have the same format as the line protocol.
Multiple tags are separated by commas.

```toml
templates = [
    "measurement.measurement.field.region datacenter=1a"
]
```

would result in the following Graphite -> Telegraf transformation.

```
cpu.usage.idle.us-west 100
=> cpu_usage,region=us-west,datacenter=1a idle=100
```

There are many more options available,
[More details can be found here](https://github.com/influxdata/influxdb/tree/master/services/graphite#templates)

#### Graphite Configuration:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/tmp/test.sh", "/usr/bin/mycollector --foo=bar"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume. This can be "json", "influx" or "graphite" (line-protocol)
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "graphite"

  ## This string will be used to join the matched values.
  separator = "_"

  ## Each template line requires a template pattern. It can have an optional
  ## filter before the template and separated by spaces. It can also have optional extra
  ## tags following the template. Multiple tags should be separated by commas and no spaces
  ## similar to the line protocol format. There can be only one default template.
  ## Templates support below format:
  ## 1. filter + template
  ## 2. filter + template + extra tag
  ## 3. filter + template with field key
  ## 4. default template
  templates = [
    "*.app env.service.resource.measurement",
    "stats.* .host.measurement* region=us-west,agent=sensu",
    "stats2.* .host.measurement.field",
    "measurement*"
  ]
```

## LTSV:

The [Labeled Tab-separated Values (LTSV)](http://ltsv.org/) data format translate a LTSV line into a measurement with _timestamp_, _fields_ and _tags_. For example, this line:

```
time:2016-03-06T09:24:12Z\tstr1:value1\tint1:23\tint2:34\tfloat1:1.23\tbool1:true\tbool2:false\tignore_field1:foo\ttag1:tval1\tignore_tag1:bar\ttag2:tval2
```

Would get translate into _timestamp_, _fields_ and _tags_ of a measurement using the example configuration in the following section:

```
ltsv_example str1=value1,int1=23i,int2=34i,float1=1.23,bool1=true,bool2=false tag1=tval1,tag2=tval2,log_host=log.example.com 2016-03-06T09:24:12Z
```

### LTSV Configuration:

The LTSV data format specifying the following configurations.

- metric_name
- time_label
- time_format
- str_field_labels
- int_field_labels
- float_field_labels
- bool_field_labels
- tag_labels
- duplicate_points_modifier_method
- duplicate_points_modifier_uniq_tag

For details, please see the comments in the following configuration example.

```toml
[[inputs.tail]]
  ## The measurement name
  override_name = "nginx_access"

  ## A LTSV formatted log file path.
  ## See http://ltsv.org/ for Labeled Tab-separated Values (LTSV)
  ## Here is an example config for nginx (http://nginx.org/en/).
  ##
  ##  log_format  ltsv  'time:$time_iso8601\t'
  ##                    'host:$host\t'
  ##                    'http_host:$http_host\t'
  ##                    'scheme:$scheme\t'
  ##                    'remote_addr:$remote_addr\t'
  ##                    'remote_user:$remote_user\t'
  ##                    'request:$request\t'
  ##                    'status:$status\t'
  ##                    'body_bytes_sent:$body_bytes_sent\t'
  ##                    'http_referer:$http_referer\t'
  ##                    'http_user_agent:$http_user_agent\t'
  ##                    'http_x_forwarded_for:$http_x_forwarded_for\t'
  ##                    'request_time:$request_time';
  ##  access_log  /var/log/nginx/access.ltsv.log  ltsv;
  ##
  filename = "/var/log/nginx/access.ltsv.log"

  ## Seek to this location before tailing
  seek_offset = 0

  ## Seek from whence. See https://golang.org/pkg/os/#File.Seek
  seek_whence = 0

  ## Reopen recreated files (tail -F)
  re_open = true

  ## Fail early if the file does not exist
  must_exist = false

  ## Poll for file changes instead of using inotify
  poll = false

  ## Set this to true if the file is a named pipe (mkfifo)
  pipe = false

  ## Continue looking for new lines (tail -f)
  follow = true

  ## If non-zero, split longer lines into multiple lines
  max_line_size = 0

  ## Set this false to enable logging to stderr, true to disable logging
  disable_logging = false

  ## Data format to consume. Currently only "ltsv" is supported.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "ltsv"

  ## Time label to be used to create a timestamp for a measurement.
  time_label = "time"

  ## Time format for parsing timestamps.
  ## Please see https://golang.org/pkg/time/#Parse for the format string.
  time_format = "2006-01-02T15:04:05Z07:00"

  ## Labels for string fields.
  str_field_labels = ["str1"]

  ## Labels for integer (64bit signed decimal integer) fields.
  ## For acceptable integer values, please refer to:
  ## https://golang.org/pkg/strconv/#ParseInt
  int_field_labels = ["int1", "int2"]

  ## Labels for float (64bit float) fields.
  ## For acceptable float values, please refer to:
  ## https://golang.org/pkg/strconv/#ParseFloat
  float_field_labels = ["float1"]

  ## Labels for boolean fields.
  ## For acceptable boolean values, please refer to:
  ## https://golang.org/pkg/strconv/#ParseBool
  bool_field_labels = ["bool1", "bool2"]

  ## Labels for tags to be added
  tag_labels = ["tag1", "tag2"]

  ## Method to modify duplicated measurement points.
  ## Must be one of "add_uniq_tag", "increment_time", "no_op".
  ## This will be used to modify duplicated points.
  ## For detail, please see https://docs.influxdata.com/influxdb/v0.10/troubleshooting/frequently_encountered_issues/#writing-duplicate-points
  ## NOTE: For modifier methods other than "no_op" to work correctly, the log lines
  ## MUST be sorted by timestamps in ascending order.
  duplicate_points_modifier_method = "add_uniq_tag"

  ## When duplicate_points_modifier_method is "increment_time",
  ## this will be added to the time of the previous measurement
  ## if the time of current time is equal to or less than the
  ## time of the previous measurement.
  ##
  ## NOTE: You need to set this value equal to or greater than
  ## precisions of your output plugins. Otherwise the times will
  ## become the same value!
  ## For the precision of the InfluxDB plugin, please see
  ## https://github.com/influxdata/telegraf/blob/v0.10.1/plugins/outputs/influxdb/influxdb.go#L40-L42
  ## For the duration string format, please see
  ## https://golang.org/pkg/time/#ParseDuration
  duplicate_points_increment_duration = "1us"

  ## When duplicate_points_modifier_method is "add_uniq_tag",
  ## this will be the label of the tag to be added to ensure uniqueness of points.
  ## NOTE: The uniq tag will be only added to the successive points of duplicated
  ## points, it will not be added to the first point of duplicated points.
  ## If you want to always add the uniq tag, add a tag with the same name as
  ## duplicate_points_modifier_uniq_tag and the string value "0" to [inputs.tail.tags].
  duplicate_points_modifier_uniq_tag = "uniq"

  ## Defaults tags to be added to measurements.
  [inputs.tail.tags]
    log_host = "log.example.com"
```
