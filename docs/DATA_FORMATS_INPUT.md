# Telegraf Input Data Formats

Telegraf is able to parse the following input data formats into metrics:

1. [InfluxDB Line Protocol](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#influx)
1. [JSON](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#json)
1. [Graphite](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#graphite)
1. [Value](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#value), ie: 45 or "booyah"
1. [Nagios](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#nagios) (exec input only)
1. [Collectd](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#collectd)
1. [Dropwizard](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#dropwizard)
1. [Grok](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#grok)

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
  ## Each data format has its own unique set of configuration options, read
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
  ## Each data format has its own unique set of configuration options, read
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
  ## Each data format has its own unique set of configuration options, read
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

If the JSON data is an array, then each element of the array is parsed with the configured settings.
Each resulting metric will be output with the same timestamp.

For example, if the following configuration:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["/usr/bin/mycollector --foo=bar"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
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
[
    {
        "a": 5,
        "b": {
            "c": 6
        },
        "my_tag_1": "foo",
        "my_tag_2": "baz"
    },
    {
        "a": 7,
        "b": {
            "c": 8
        },
        "my_tag_1": "bar",
        "my_tag_2": "baz"
    }
]
```

Your Telegraf metrics would get tagged with "my_tag_1" and "my_tag_2"

```
exec_mycollector,my_tag_1=foo,my_tag_2=baz a=5,b_c=6
exec_mycollector,my_tag_1=bar,my_tag_2=baz a=7,b_c=8
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
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "value"
  data_type = "integer" # required
```

# Graphite:

The Graphite data format translates graphite _dot_ buckets directly into
telegraf measurement names, with a single value field, and without any tags.
By default, the separator is left as ".", but this can be changed using the
"separator" argument. For more advanced options,
Telegraf supports specifying "templates" to translate
graphite buckets into Telegraf metrics.

Templates are of the form:

```
"host.mytag.mytag.measurement.measurement.field*"
```

Where the following keywords exist:

1. `measurement`: specifies that this section of the graphite bucket corresponds
to the measurement name. This can be specified multiple times.
2. `field`: specifies that this section of the graphite bucket corresponds
to the field name. This can be specified multiple times.
3. `measurement*`: specifies that all remaining elements of the graphite bucket
correspond to the measurement name.
4. `field*`: specifies that all remaining elements of the graphite bucket
correspond to the field name.

Any part of the template that is not a keyword is treated as a tag key. This
can also be specified multiple times.

NOTE: `field*` cannot be used in conjunction with `measurement*`!

#### Measurement & Tag Templates:

The most basic template is to specify a single transformation to apply to all
incoming metrics. So the following template:

```toml
templates = [
    "region.region.measurement*"
]
```

would result in the following Graphite -> Telegraf transformation.

```
us.west.cpu.load 100
=> cpu.load,region=us.west value=100
```

Multiple templates can also be specified, but these should be differentiated
using _filters_ (see below for more details)

```toml
templates = [
    "*.*.* region.region.measurement", # <- all 3-part measurements will match this one.
    "*.*.*.* region.region.host.measurement", # <- all 4-part measurements will match this one.
]
```

#### Field Templates:

The field keyword tells Telegraf to give the metric that field name.
So the following template:

```toml
separator = "_"
templates = [
    "measurement.measurement.field.field.region"
]
```

would result in the following Graphite -> Telegraf transformation.

```
cpu.usage.idle.percent.eu-east 100
=> cpu_usage,region=eu-east idle_percent=100
```

The field key can also be derived from all remaining elements of the graphite
bucket by specifying `field*`:

```toml
separator = "_"
templates = [
    "measurement.measurement.region.field*"
]
```

which would result in the following Graphite -> Telegraf transformation.

```
cpu.usage.eu-east.idle.percentage 100
=> cpu_usage,region=eu-east idle_percentage=100
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
cpu.load.eu-east 100
=> cpu_load,region=eu-east value=100

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
cpu.usage.idle.eu-east 100
=> cpu_usage,region=eu-east,datacenter=1a idle=100
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
  ## Each data format has its own unique set of configuration options, read
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
  ## 2. filter + template + extra tag(s)
  ## 3. filter + template with field key
  ## 4. default template
  templates = [
    "*.app env.service.resource.measurement",
    "stats.* .host.measurement* region=eu-east,agent=sensu",
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
  commands = ["/usr/lib/nagios/plugins/check_load -w 5,6,7 -c 7,8,9"]

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "nagios"
```

# Collectd:

The collectd format parses the collectd binary network protocol.  Tags are
created for host, instance, type, and type instance.  All collectd values are
added as float64 fields.

For more information about the binary network protocol see
[here](https://collectd.org/wiki/index.php/Binary_protocol).

You can control the cryptographic settings with parser options.  Create an
authentication file and set `collectd_auth_file` to the path of the file, then
set the desired security level in `collectd_security_level`.

Additional information including client setup can be found
[here](https://collectd.org/wiki/index.php/Networking_introduction#Cryptographic_setup).

You can also change the path to the typesdb or add additional typesdb using
`collectd_typesdb`.

#### Collectd Configuration:

```toml
[[inputs.socket_listener]]
  service_address = "udp://127.0.0.1:25826"
  name_prefix = "collectd_"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "collectd"

  ## Authentication file for cryptographic security levels
  collectd_auth_file = "/etc/collectd/auth_file"
  ## One of none (default), sign, or encrypt
  collectd_security_level = "encrypt"
  ## Path of to TypesDB specifications
  collectd_typesdb = ["/usr/share/collectd/types.db"]

  # Multi-value plugins can be handled two ways.  
  # "split" will parse and store the multi-value plugin data into separate measurements
  # "join" will parse and store the multi-value plugin as a single multi-value measurement.  
  # "split" is the default behavior for backward compatability with previous versions of influxdb.
  collectd_parse_multivalue = "split"
```

# Dropwizard:

The dropwizard format can parse the JSON representation of a single dropwizard metric registry. By default, tags are parsed from metric names as if they were actual influxdb line protocol keys (`measurement<,tag_set>`) which can be overriden by defining custom [measurement & tag templates](./DATA_FORMATS_INPUT.md#measurement--tag-templates). All field value types are supported, `string`, `number` and `boolean`.

A typical JSON of a dropwizard metric registry:

```json
{
	"version": "3.0.0",
	"counters" : {
		"measurement,tag1=green" : {
			"count" : 1
		}
	},
	"meters" : {
		"measurement" : {
			"count" : 1,
			"m15_rate" : 1.0,
			"m1_rate" : 1.0,
			"m5_rate" : 1.0,
			"mean_rate" : 1.0,
			"units" : "events/second"
		}
	},
	"gauges" : {
		"measurement" : {
			"value" : 1
		}
	},
	"histograms" : {
		"measurement" : {
			"count" : 1,
			"max" : 1.0,
			"mean" : 1.0,
			"min" : 1.0,
			"p50" : 1.0,
			"p75" : 1.0,
			"p95" : 1.0,
			"p98" : 1.0,
			"p99" : 1.0,
			"p999" : 1.0,
			"stddev" : 1.0
		}
	},
	"timers" : {
		"measurement" : {
			"count" : 1,
			"max" : 1.0,
			"mean" : 1.0,
			"min" : 1.0,
			"p50" : 1.0,
			"p75" : 1.0,
			"p95" : 1.0,
			"p98" : 1.0,
			"p99" : 1.0,
			"p999" : 1.0,
			"stddev" : 1.0,
			"m15_rate" : 1.0,
			"m1_rate" : 1.0,
			"m5_rate" : 1.0,
			"mean_rate" : 1.0,
			"duration_units" : "seconds",
			"rate_units" : "calls/second"
		}
	}
}
```

Would get translated into 4 different measurements:

```
measurement,metric_type=counter,tag1=green count=1
measurement,metric_type=meter count=1,m15_rate=1.0,m1_rate=1.0,m5_rate=1.0,mean_rate=1.0
measurement,metric_type=gauge value=1
measurement,metric_type=histogram count=1,max=1.0,mean=1.0,min=1.0,p50=1.0,p75=1.0,p95=1.0,p98=1.0,p99=1.0,p999=1.0
measurement,metric_type=timer count=1,max=1.0,mean=1.0,min=1.0,p50=1.0,p75=1.0,p95=1.0,p98=1.0,p99=1.0,p999=1.0,stddev=1.0,m15_rate=1.0,m1_rate=1.0,m5_rate=1.0,mean_rate=1.0
```

You may also parse a dropwizard registry from any JSON document which contains a dropwizard registry in some inner field. 
Eg. to parse the following JSON document:

```json
{
	"time" : "2017-02-22T14:33:03.662+02:00",
	"tags" : {
		"tag1" : "green",
		"tag2" : "yellow"
	},
	"metrics" : {
		"counters" : 	{ 
			"measurement" : {
				"count" : 1
			}
		},
		"meters" : {},
		"gauges" : {},
		"histograms" : {},
		"timers" : {}
	}
}
```
and translate it into:

```
measurement,metric_type=counter,tag1=green,tag2=yellow count=1 1487766783662000000
```

you simply need to use the following additional configuration properties:

```toml
dropwizard_metric_registry_path = "metrics"
dropwizard_time_path = "time"
dropwizard_time_format = "2006-01-02T15:04:05Z07:00"
dropwizard_tags_path = "tags"
## tag paths per tag are supported too, eg.
#[inputs.yourinput.dropwizard_tag_paths]
#  tag1 = "tags.tag1"
#  tag2 = "tags.tag2"
```


For more information about the dropwizard json format see
[here](http://metrics.dropwizard.io/3.1.0/manual/json/).

#### Dropwizard Configuration:

```toml
[[inputs.exec]]
  ## Commands array
  commands = ["curl http://localhost:8080/sys/metrics"]
  timeout = "5s"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "dropwizard"

  ## Used by the templating engine to join matched values when cardinality is > 1
  separator = "_"

  ## Each template line requires a template pattern. It can have an optional
  ## filter before the template and separated by spaces. It can also have optional extra
  ## tags following the template. Multiple tags should be separated by commas and no spaces
  ## similar to the line protocol format. There can be only one default template.
  ## Templates support below format:
  ## 1. filter + template
  ## 2. filter + template + extra tag(s)
  ## 3. filter + template with field key
  ## 4. default template
  ## By providing an empty template array, templating is disabled and measurements are parsed as influxdb line protocol keys (measurement<,tag_set>)
  templates = []

  ## You may use an appropriate [gjson path](https://github.com/tidwall/gjson#path-syntax) 
  ## to locate the metric registry within the JSON document
  # dropwizard_metric_registry_path = "metrics"
  
  ## You may use an appropriate [gjson path](https://github.com/tidwall/gjson#path-syntax) 
  ## to locate the default time of the measurements within the JSON document
  # dropwizard_time_path = "time"
  # dropwizard_time_format = "2006-01-02T15:04:05Z07:00"
  
  ## You may use an appropriate [gjson path](https://github.com/tidwall/gjson#path-syntax) 
  ## to locate the tags map within the JSON document
  # dropwizard_tags_path = "tags"

  ## You may even use tag paths per tag
  # [inputs.exec.dropwizard_tag_paths]
  #   tag1 = "tags.tag1"
  #   tag2 = "tags.tag2"
```

#### Grok
Parse logstash-style "grok" patterns. Patterns can be added to patterns, or custom patterns read from custom_pattern_files.

# View logstash grok pattern docs here:
#   https://www.elastic.co/guide/en/logstash/current/plugins-filters-grok.html
# All default logstash patterns are supported, these can be viewed here:
#   https://github.com/logstash-plugins/logstash-patterns-core/blob/master/patterns/grok-patterns

# Available modifiers:
#   string   (default if nothing is specified)
#   int
#   float
#   duration (ie, 5.23ms gets converted to int nanoseconds)
#   tag      (converts the field into a tag)
#   drop     (drops the field completely)
# Timestamp modifiers:
#   ts-ansic         ("Mon Jan _2 15:04:05 2006")
#   ts-unix          ("Mon Jan _2 15:04:05 MST 2006")
#   ts-ruby          ("Mon Jan 02 15:04:05 -0700 2006")
#   ts-rfc822        ("02 Jan 06 15:04 MST")
#   ts-rfc822z       ("02 Jan 06 15:04 -0700")
#   ts-rfc850        ("Monday, 02-Jan-06 15:04:05 MST")
#   ts-rfc1123       ("Mon, 02 Jan 2006 15:04:05 MST")
#   ts-rfc1123z      ("Mon, 02 Jan 2006 15:04:05 -0700")
#   ts-rfc3339       ("2006-01-02T15:04:05Z07:00")
#   ts-rfc3339nano   ("2006-01-02T15:04:05.999999999Z07:00")
#   ts-httpd         ("02/Jan/2006:15:04:05 -0700")
#   ts-epoch         (seconds since unix epoch)
#   ts-epochnano     (nanoseconds since unix epoch)
#   ts-"CUSTOM"
# CUSTOM time layouts must be within quotes and be the representation of the
# "reference time", which is Mon Jan 2 15:04:05 -0700 MST 2006
# See https://golang.org/pkg/time/#Parse for more details.

# Example log file pattern, example log looks like this:
#   [04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs
# Breakdown of the DURATION pattern below:
#   NUMBER  is a builtin logstash grok pattern matching float & int numbers.
#   [nuµm]? is a regex specifying 0 or 1 of the characters within brackets.
#   s       is also regex, this pattern must end in "s".
# so DURATION will match something like '5.324ms' or '6.1µs' or '10s'
DURATION %{NUMBER}[nuµm]?s
RESPONSE_CODE %{NUMBER:response_code:tag}
RESPONSE_TIME %{DURATION:response_time_ns:duration}
EXAMPLE_LOG \[%{HTTPDATE:ts:ts-httpd}\] %{NUMBER:myfloat:float} %{RESPONSE_CODE} %{IPORHOST:clientip} %{RESPONSE_TIME}

# Wider-ranging username matching vs. logstash built-in %{USER}
NGUSERNAME [a-zA-Z0-9\.\@\-\+_%]+
NGUSER %{NGUSERNAME}
# Wider-ranging client IP matching
CLIENT (?:%{IPORHOST}|%{HOSTPORT}|::1)

##
## COMMON LOG PATTERNS
##

# apache & nginx logs, this is also known as the "common log format"
#   see https://en.wikipedia.org/wiki/Common_Log_Format
COMMON_LOG_FORMAT %{CLIENT:client_ip} %{NOTSPACE:ident} %{NOTSPACE:auth} \[%{HTTPDATE:ts:ts-httpd}\] "(?:%{WORD:verb:tag} %{NOTSPACE:request}(?: HTTP/%{NUMBER:http_version:float})?|%{DATA})" %{NUMBER:resp_code:tag} (?:%{NUMBER:resp_bytes:int}|-)

# Combined log format is the same as the common log format but with the addition
# of two quoted strings at the end for "referrer" and "agent"
#   See Examples at http://httpd.apache.org/docs/current/mod/mod_log_config.html
COMBINED_LOG_FORMAT %{COMMON_LOG_FORMAT} %{QS:referrer} %{QS:agent}

# HTTPD log formats
HTTPD20_ERRORLOG \[%{HTTPDERROR_DATE:timestamp}\] \[%{LOGLEVEL:loglevel:tag}\] (?:\[client %{IPORHOST:clientip}\] ){0,1}%{GREEDYDATA:errormsg}
HTTPD24_ERRORLOG \[%{HTTPDERROR_DATE:timestamp}\] \[%{WORD:module}:%{LOGLEVEL:loglevel:tag}\] \[pid %{POSINT:pid:int}:tid %{NUMBER:tid:int}\]( \(%{POSINT:proxy_errorcode:int}\)%{DATA:proxy_errormessage}:)?( \[client %{IPORHOST:client}:%{POSINT:clientport}\])? %{DATA:errorcode}: %{GREEDYDATA:message}
HTTPD_ERRORLOG %{HTTPD20_ERRORLOG}|%{HTTPD24_ERRORLOG}

#### Grok Configuration:
```toml
[[inputs.reader]]
  ## This is a list of patterns to check the given log file(s) for.
  ## Note that adding patterns here increases processing time. The most
  ## efficient configuration is to have one pattern per logparser.
  ## Other common built-in patterns are:
  ##   %{COMMON_LOG_FORMAT}   (plain apache & nginx access logs)
  ##   %{COMBINED_LOG_FORMAT} (access logs + referrer & agent)
  grok_patterns = ["%{COMBINED_LOG_FORMAT}"]

  ## Name of the outputted measurement name.
  grok_name_override = "apache_access_log"

  ## Full path(s) to custom pattern files.
  grok_custom_pattern_files = []

  ## Custom patterns can also be defined here. Put one pattern per line.
  grok_custom_patterns = '''
  '''

  ## Timezone allows you to provide an override for timestamps that
  ## don't already include an offset
  ## e.g. 04/06/2016 12:41:45 data one two 5.43µs
  ##
  ## Default: "" which renders UTC
  ## Options are as follows:
  ##   1. Local             -- interpret based on machine localtime
  ##   2. "Canada/Eastern"  -- Unix TZ values like those found in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
  ##   3. UTC               -- or blank/unspecified, will return timestamp in UTC
  grok_timezone = "Canada/Eastern"
```