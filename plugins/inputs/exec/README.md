# Exec Input Plugin

The exec plugin can execute arbitrary commands which output JSON or
InfluxDB [line-protocol](https://docs.influxdata.com/influxdb/v0.9/write_protocols/line/).

If using JSON, only numeric values are parsed and turned into floats. Booleans
and strings will be ignored.

### Configuration

```
# Read flattened metrics from one or more commands that output JSON to stdout
[[inputs.exec]]
  # the command to run
  command = "/usr/bin/mycollector --foo=bar"

  # Data format to consume. This can be "json" or "influx" (line-protocol)
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "json"

  # measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"
```

Other options for modifying the measurement names are:

```
name_override = "measurement_name"
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
  # the command to run
  command = "/usr/bin/line_protocol_collector"

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
