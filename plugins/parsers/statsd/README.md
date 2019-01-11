# Statsd

The statsd data format tranlsates statsd *dot* buckets directly into telegraf measurement names, with a single value field, more than one tags.
This parser default use [Template Patterns](/docs/TEMPLATE_PATTERN) with separator `_`.

### Configuration
```toml
[[inputs.socket_listener]]
  ## URL to listen on
  # service_address = "tcp://:8094"
  # service_address = "tcp://127.0.0.1:http"
  # service_address = "tcp4://:8094"
  # service_address = "tcp6://:8094"
  # service_address = "tcp6://[2001:db8::1]:8094"
  # service_address = "udp://:8094"
  # service_address = "udp4://:8094"
  # service_address = "udp6://:8094"
  # service_address = "unix:///tmp/telegraf.sock"
  # service_address = "unixgram:///tmp/telegraf.sock"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

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

### Description
The format of the statsd messages was based on the format described in the
original [etsy statsd](https://github.com/etsy/statsd/blob/master/docs/metric_types.md)
implementation. In short, the telegraf statsd parser will accept:

- Gauges
    - `users.current.den001.myapp:32|g` <- standard
    - `users.current.den001.myapp:+10|g` <- additive
    - `users.current.den001.myapp:-10|g`
- Counters
    - `deploys.test.myservice:1|c` <- increments by 1
    - `deploys.test.myservice:101|c` <- increments by 101
    - `deploys.test.myservice:1|c|@0.1` <- with sample rate, increments by 10
- Sets
    - `users.unique:101|s`
    - `users.unique:101|s`
    - `users.unique:102|s`
- Timings & Histograms
    - `load.time:320|ms`
    - `load.time.nanoseconds:1|h`
    - `load.time:200|ms|@0.1` <- sampled 1/10 of the time, result in 10 metric

It is possible to omit repetitive names and merge individual stats into a
single line by separating them with additional colons:

  - `users.current.den001.myapp:32|g:+10|g:-10|g`
  - `deploys.test.myservice:1|c:101|c:1|c|@0.1`
  - `users.unique:101|s:101|s:102|s`
  - `load.time:320|ms:200|ms|@0.1`

This also allows for mixed types in a single line:

  - `foo:1|c:200|ms`

The string `foo:1|c:200|ms` is internally split into two individual metrics
`foo:1|c` and `foo:200|ms` which are added to the aggregator separately.


### Statsd bucket with tags
In order to take advantage of InfluxDB's tagging system, we have made a couple
additions to the standard statsd protocol. First, you can specify
tags in a manner similar to the line-protocol, like this:

```
users.current,service=payroll,region=us-west:32|g
```


### Statsd bucket -> InfluxDB line-protocol Templates
The plugin supports specifying [Template Patterns](/docs/TEMPLATE_PATTERN.md) for transforming statsd buckets into
InfluxDB measurement names and tags. The templates have a _measurement_ keyword,
which can be used to specify parts of the bucket that are to be used in the
measurement name. Other words in the template are used as tag names. For example,
the following template:

```
templates = [
    "measurement.measurement.region"
]
```

would result in the following transformation:

```
cpu.load.us-west:100|g
=> cpu_load,region=us-west 100
```

Users can also filter the template to use based on the name of the bucket,
using glob matching, like so:

```
templates = [
    "cpu.* measurement.measurement.region",
    "mem.* measurement.measurement.host"
]
```

which would result in the following transformation:

```
cpu.load.us-west:100|g
=> cpu_load,region=us-west 100

mem.cached.localhost:256|g
=> mem_cached,host=localhost 256
```

Consult the [Template Patterns](/docs/TEMPLATE_PATTERN.md) documentation for
additional details.
