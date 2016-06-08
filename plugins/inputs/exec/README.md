# Exec Input Plugin

Please also see: [Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md)

### Example 1 - JSON

#### Configuration

In this example a script called ```/tmp/test.sh```, a script called ```/tmp/test2.sh```, and
all scripts matching glob pattern ```/tmp/collect_*.sh``` are configured for ```[[inputs.exec]]```
in JSON format. Glob patterns are matched on every run, so adding new scripts that match the pattern
will cause them to be picked up immediately.

```toml
# Read flattened metrics from one or more commands that output JSON to stdout
[[inputs.exec]]
  # Shell/commands array
  # Full command line to executable with parameters, or a glob pattern to run all matching files.
  commands = ["/tmp/test.sh", "/tmp/test2.sh", "/tmp/collect_*.sh"]

  ## Timeout for each command to complete.
  timeout = "5s"

  # Data format to consume.
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "json"

  # measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"
```

Other options for modifying the measurement names are:

```
name_prefix = "prefix_"
```

Let's say that we have the above configuration, and mycollector outputs the
following JSON:

```json
{
    "a": 0.5,
    "b": {
        "c": 0.1,
        "d": 5
    }
}
```

The collected metrics will be stored as fields under the measurement
"exec_mycollector":

```
exec_mycollector a=0.5,b_c=0.1,b_d=5 1452815002357578567
```
If using JSON, only numeric values are parsed and turned into floats. Booleans
and strings will be ignored.

### Example 2 - Influx Line-Protocol

In this example an application called ```/usr/bin/line_protocol_collector```
and a script called ```/tmp/test2.sh``` are configured for ```[[inputs.exec]]```
in influx line-protocol format.

#### Configuration

```toml
[[inputs.exec]]
  # Shell/commands array
  # compatible with old version
  # we can still use the old command configuration
  # command = "/usr/bin/line_protocol_collector"
  commands = ["/usr/bin/line_protocol_collector","/tmp/test2.sh"]

  ## Timeout for each command to complete.
  timeout = "5s"

  # Data format to consume.
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "influx"
```

The line_protocol_collector application outputs the following line protocol:

```
cpu,cpu=cpu0,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu1,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu2,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu3,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu4,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu5,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu6,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
```

You will get data in InfluxDB exactly as it is defined above,
tags are cpu=cpuN, host=foo, and datacenter=us-east with fields usage_idle
and usage_busy. They will receive a timestamp at collection time.
Each line must end in \n, just as the Influx line protocol does.


### Example 3 - Graphite

We can also change the data_format to "graphite" to use the metrics collecting scripts such as (compatible with graphite):

* Nagios [Metrics Plugins](https://exchange.nagios.org/directory/Plugins)
* Sensu [Metrics Plugins](https://github.com/sensu-plugins)

In this example a script called /tmp/test.sh and a script called /tmp/test2.sh are configured for [[inputs.exec]] in graphite format.

#### Configuration

```toml
# Read flattened metrics from one or more commands that output JSON to stdout
[[inputs.exec]]
  # Shell/commands array
  commands = ["/tmp/test.sh","/tmp/test2.sh"]

  ## Timeout for each command to complete.
  timeout = "5s"

  # Data format to consume.
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "graphite"

  # measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Below configuration will be used for data_format = "graphite", can be ignored for other data_format
  ## If matching multiple measurement files, this string will be used to join the matched values.
  separator = "."

  ## Each template line requires a template pattern.  It can have an optional
  ## filter before the template and separated by spaces.  It can also have optional extra
  ## tags following the template.  Multiple tags should be separated by commas and no spaces
  ## similar to the line protocol format.  The can be only one default template.
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
Graphite messages are in this format:

```
metric_path value timestamp\n
```

__metric_path__ is the metric namespace that you want to populate.

__value__ is the value that you want to assign to the metric at this time.

__timestamp__ is the unix epoch time.

And test.sh/test2.sh will output:

```
sensu.metric.net.server0.eth0.rx_packets 461295119435 1444234982
sensu.metric.net.server0.eth0.tx_bytes 1093086493388480 1444234982
sensu.metric.net.server0.eth0.rx_bytes 1015633926034834 1444234982
sensu.metric.net.server0.eth0.tx_errors 0 1444234982
sensu.metric.net.server0.eth0.rx_errors 0 1444234982
sensu.metric.net.server0.eth0.tx_dropped 0 1444234982
sensu.metric.net.server0.eth0.rx_dropped 0 1444234982
```

The templates configuration will be used to parse the graphite metrics to support influxdb/opentsdb tagging store engines.

More detail information about templates, please refer to [The graphite Input](https://github.com/influxdata/influxdb/blob/master/services/graphite/README.md)
