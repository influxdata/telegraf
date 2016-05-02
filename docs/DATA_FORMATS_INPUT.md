# Telegraf Input Data Formats

Telegraf is able to parse the following input data formats into metrics:

1. [InfluxDB Line Protocol](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#influx)
1. [JSON](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#json)
1. [Graphite](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#graphite)
1. [Value](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#value), ie: 45 or "booyah"
1. [Nagios](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#nagios) (exec input only)

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

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"

  ## Additional configuration options go here
```

Each data_format has an additional set of configuration options available, which
I'll go over below.

# Influx:

There are no additional configuration options for InfluxDB line-protocol. The
metrics are parsed directly into Telegraf metrics.

#### Influx Configuration:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/tmp/test.sh", "/usr/bin/mycollector --foo=bar"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

# JSON:

The JSON data format flattens JSON into metric _fields_.
NOTE: Only numerical values are converted to fields, and they are converted
into a float. strings are ignored unless specified as a tag_key (see below).

So for example, this JSON:

```json
{
    "a": 5,
    "b": {
        "c": 6
    },
    "ignored": "I'm a string"
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

  ## Data format to consume.
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

# Value:

The "value" data format translates single values into Telegraf metrics. This
is done by assigning a measurement name and setting a single field ("value")
as the parsed metric.

#### Value Configuration:

You **must** tell Telegraf what type of metric to collect by using the
`data_type` configuration option. Available options are:

1. integer
2. float or long
3. string
4. boolean

**Note:** It is also recommended that you set `name_override` to a measurement
name that makes sense for your metric, otherwise it will just be set to the
name of the plugin.

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["cat /proc/sys/kernel/random/entropy_avail"]

  ## override the default metric name of "exec"
  name_override = "entropy_available"

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "value"
  data_type = "integer" # required
```

# Graphite:

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
    "measurement.measurement.field.field.region"
]
```

would result in the following Graphite -> Telegraf transformation.

```
cpu.usage.idle.percent.us-west 100
=> cpu_usage,region=us-west idle_percent=100
```

The field key can also be derived from the second "half" of the input metric-name by specifying ```field*```:
```toml
templates = [
    "measurement.measurement.region.field*"
]
```

would result in the following Graphite -> Telegraf transformation.

```
cpu.usage.us-west.idle.percentage 100
=> cpu_usage,region=us-west idle_percentage=100
```
(This cannot be used in conjunction with "measurement*"!)

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

  ## Data format to consume.
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

# Nagios:

There are no additional configuration options for Nagios line-protocol. The
metrics are parsed directly into Telegraf metrics.

Note: Nagios Input Data Formats is only supported in `exec` input plugin.

#### Nagios Configuration:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/usr/lib/nagios/plugins/check_load", "-w 5,6,7 -c 7,8,9"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "nagios"
```
