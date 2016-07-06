# Telegraf Service Plugin: statsd

### Configuration

```toml
# Statsd Server
[[inputs.statsd]]
  ## Address and port to host UDP listener on
  service_address = ":8125"
  ## Delete gauges every interval (default=false)
  delete_gauges = false
  ## Delete counters every interval (default=false)
  delete_counters = false
  ## Delete sets every interval (default=false)
  delete_sets = false
  ## Delete timings & histograms every interval (default=true)
  delete_timings = true
  ## Percentiles to calculate for timing & histogram stats
  percentiles = [90]

  ## separator to use between elements of a statsd metric
  metric_separator = "_"

  ## Parses tags in the datadog statsd format
  ## http://docs.datadoghq.com/guides/dogstatsd/
  parse_data_dog_tags = false

  ## Statsd data translation templates, more info can be read here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#graphite
  # templates = [
  #     "cpu.* measurement*"
  # ]

  ## Number of UDP messages allowed to queue up, once filled,
  ## the statsd server will start dropping packets
  allowed_pending_messages = 10000

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000
```

### Description

The statsd plugin is a special type of plugin which runs a backgrounded statsd
listener service while telegraf is running.

The format of the statsd messages was based on the format described in the
original [etsy statsd](https://github.com/etsy/statsd/blob/master/docs/metric_types.md)
implementation. In short, the telegraf statsd listener will accept:

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
    - `users.unique:102|s` <- would result in a count of 2 for `users.unique`
- Timings & Histograms
    - `load.time:320|ms`
    - `load.time.nanoseconds:1|h`
    - `load.time:200|ms|@0.1` <- sampled 1/10 of the time

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


### Influx Statsd

In order to take advantage of InfluxDB's tagging system, we have made a couple
additions to the standard statsd protocol. First, you can specify
tags in a manner similar to the line-protocol, like this:

```
users.current,service=payroll,region=us-west:32|g
```

COMING SOON: there will be a way to specify multiple fields.
<!-- TODO Second, you can specify multiple fields within a measurement:

```
current.users,service=payroll,server=host01:west=10,east=10,central=2,south=10|g
``` -->

### Measurements:

Meta:
- tags: `metric_type=<gauge|set|counter|timing|histogram>`

Outputted measurements will depend entirely on the measurements that the user
sends, but here is a brief rundown of what you can expect to find from each
metric type:

- Gauges
    - Gauges are a constant data type. They are not subject to averaging, and they
    don’t change unless you change them. That is, once you set a gauge value, it
    will be a flat line on the graph until you change it again.
- Counters
    - Counters are the most basic type. They are treated as a count of a type of
    event. They will continually increase unless you set `delete_counters=true`.
- Sets
    - Sets count the number of unique values passed to a key. For example, you
    could count the number of users accessing your system using `users:<user_id>|s`.
    No matter how many times the same user_id is sent, the count will only increase
    by 1.
- Timings & Histograms
    - Timers are meant to track how long something took. They are an invaluable
    tool for tracking application performance.
    - The following aggregate measurements are made for timers:
        - `statsd_<name>_lower`: The lower bound is the lowest value statsd saw
        for that stat during that interval.
        - `statsd_<name>_upper`: The upper bound is the highest value statsd saw
        for that stat during that interval.
        - `statsd_<name>_mean`: The mean is the average of all values statsd saw
        for that stat during that interval.
        - `statsd_<name>_stddev`: The stddev is the sample standard deviation
        of all values statsd saw for that stat during that interval.
        - `statsd_<name>_count`: The count is the number of timings statsd saw
        for that stat during that interval. It is not averaged.
        - `statsd_<name>_percentile_<P>` The `Pth` percentile is a value x such
        that `P%` of all the values statsd saw for that stat during that time
        period are below x. The most common value that people use for `P` is the
        `90`, this is a great number to try to optimize.

### Plugin arguments

- **service_address** string: Address to listen for statsd UDP packets on
- **delete_gauges** boolean: Delete gauges on every collection interval
- **delete_counters** boolean: Delete counters on every collection interval
- **delete_sets** boolean: Delete set counters on every collection interval
- **delete_timings** boolean: Delete timings on every collection interval
- **percentiles** []int: Percentiles to calculate for timing & histogram stats
- **allowed_pending_messages** integer: Number of messages allowed to queue up
waiting to be processed. When this fills, messages will be dropped and logged.
- **percentile_limit** integer: Number of timing/histogram values to track
per-measurement in the calculation of percentiles. Raising this limit increases
the accuracy of percentiles but also increases the memory usage and cpu time.
- **templates** []string: Templates for transforming statsd buckets into influx
measurements and tags.
- **parse_data_dog_tags** boolean: Enable parsing of tags in DataDog's dogstatsd format (http://docs.datadoghq.com/guides/dogstatsd/)

### Statsd bucket -> InfluxDB line-protocol Templates

The plugin supports specifying templates for transforming statsd buckets into
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

There are many more options available,
[More details can be found here](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#graphite)
