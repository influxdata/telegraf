# Exec Input Plugin

The exec plugin can execute arbitrary commands which output:

* JSON
* InfluxDB [line-protocol](https://docs.influxdata.com/influxdb/v0.9/write_protocols/line/)
* Graphite [graphite-protocol](http://graphite.readthedocs.org/en/latest/feeding-carbon.html)

> Graphite understands messages with this format:

> ``` 
metric_path value timestamp\n 
```

> __metric_path__ is the metric namespace that you want to populate.

> __value__ is the value that you want to assign to the metric at this time.

> __timestamp__ is the unix epoch time.


If using JSON, only numeric values are parsed and turned into floats. Booleans
and strings will be ignored.

### Configuration

```
# Read flattened metrics from one or more commands that output JSON to stdout
[[inputs.exec]]
  # Shell/commands array
  # compatible with old version
  # we can still use the old command configuration
  # command = "/usr/bin/mycollector --foo=bar"
  commands = ["/tmp/test.sh","/tmp/test2.sh"]

  # Data format to consume. This can be "json", "influx" or "graphite" (line-protocol)
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "json"

  # measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ### Below configuration will be used for data_format = "graphite", can be ignored for other data_format
  ### If matching multiple measurement files, this string will be used to join the matched values.
  #separator = "."

  ### Each template line requires a template pattern.  It can have an optional
  ### filter before the template and separated by spaces.  It can also have optional extra
  ### tags following the template.  Multiple tags should be separated by commas and no spaces
  ### similar to the line protocol format.  The can be only one default template.
  ### Templates support below format:
  ### 1. filter + template
  ### 2. filter + template + extra tag
  ### 3. filter + template with field key
  ### 4. default template
  #templates = [
  #  "*.app env.service.resource.measurement",
  #  "stats.* .host.measurement* region=us-west,agent=sensu",
  #  "stats2.* .host.measurement.field",
  #  "measurement*"
  #]
```

Other options for modifying the measurement names are:

```
name_prefix = "prefix_"
```

### Example 1

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

### Example 2

Now let's say we have the following configuration:

```
[[inputs.exec]]
  # Shell/commands array
  # compatible with old version
  # we can still use the old command configuration
  # command = "/usr/bin/line_protocol_collector"
  commands = ["/usr/bin/line_protocol_collector","/tmp/test2.sh"]

  # Data format to consume. This can be "json" or "influx" (line-protocol)
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "influx"
```

And line_protocol_collector outputs the following line protocol:

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


### Example 3

We can also change the data_format to "graphite" to use the metrics collecting scripts such as (compatible with graphite):

* Nagios [Mertics Plugins] (https://exchange.nagios.org/directory/Plugins)
* Sensu [Mertics Plugins] (https://github.com/sensu-plugins) 

#### Configuration
```
# Read flattened metrics from one or more commands that output JSON to stdout
[[inputs.exec]]
  # Shell/commands array
  commands = ["/tmp/test.sh","/tmp/test2.sh"]

  # Data format to consume. This can be "json", "influx" or "graphite" (line-protocol)
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "graphite"

  # measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ### Below configuration will be used for data_format = "graphite", can be ignored for other data_format
  ### If matching multiple measurement files, this string will be used to join the matched values.
  separator = "."

  ### Each template line requires a template pattern.  It can have an optional
  ### filter before the template and separated by spaces.  It can also have optional extra
  ### tags following the template.  Multiple tags should be separated by commas and no spaces
  ### similar to the line protocol format.  The can be only one default template.
  ### Templates support below format:
  ### 1. filter + template
  ### 2. filter + template + extra tag
  ### 3. filter + template with field key
  ### 4. default template
  templates = [
    "*.app env.service.resource.measurement",
    "stats.* .host.measurement* region=us-west,agent=sensu",
    "stats2.* .host.measurement.field",
    "measurement*"
  ]
```

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

More detail information about templates, please refer to [The graphite Input] (https://github.com/influxdata/influxdb/blob/master/services/graphite/README.md)
 
