# ctrlX Data Layer Input Plugin

The `ctrlx_datalayer` plugin gathers data from the ctrlX Data Layer,
a communication middleware runnning on
[ctrlX CORE devices](https://ctrlx-core.com) from
[Bosch Rexroth](https://boschrexroth.com). The platform is used for
professional automation applications like industrial automation, building
automation, robotics, IoT Gateways or as classical PLC. For more
information, see [ctrlX AUTOMATION](https://ctrlx-automation.com).

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# A ctrlX Data Layer server sent event input plugin
[[inputs.ctrlx_datalayer]]
   ## Hostname or IP address of the ctrlX CORE Data Layer server
   ##  example: server = "localhost"        # Telegraf is running directly on the device
   ##           server = "192.168.1.1"      # Connect to ctrlX CORE remote via IP
   ##           server = "host.example.com" # Connect to ctrlX CORE remote via hostname
   ##           server = "10.0.2.2:8443"    # Connect to ctrlX CORE Virtual from development environment
   server = "localhost"

   ## Authentication credentials
   username = "boschrexroth"
   password = "boschrexroth"

   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false

   ## Timeout for HTTP requests. (default: "10s")
   # timeout = "10s"


   ## Create a ctrlX Data Layer subscription.
   ## It is possible to define multiple subscriptions per host. Each subscription can have its own
   ## sampling properties and a list of nodes to subscribe to.
   ## All subscriptions share the same credentials.
   [[inputs.ctrlx_datalayer.subscription]]
      ## The name of the measurement. (default: "ctrlx")
      measurement = "memory"

      ## Configure the ctrlX Data Layer nodes which should be subscribed.
      ## address - node address in ctrlX Data Layer (mandatory)
      ## name    - field name to use in the output (optional, default: base name of address)
      ## tags    - extra node tags to be added to the output metric (optional)
      ## Note: 
      ## Use either the inline notation or the bracketed notation, not both.
      ## The tags property is only supported in bracketed notation due to toml parser restrictions
      ## Examples:
      ## Inline notation 
      nodes=[
         {name="available", address="framework/metrics/system/memavailable-mb"},
         {name="used", address="framework/metrics/system/memused-mb"},
      ]
      ## Bracketed notation
      # [[inputs.ctrlx_datalayer.subscription.nodes]]
      #    name   ="available"
      #    address="framework/metrics/system/memavailable-mb"
      #    ## Define extra tags related to node to be added to the output metric (optional)
      #    [inputs.ctrlx_datalayer.subscription.nodes.tags]
      #       node_tag1="node_tag1"
      #       node_tag2="node_tag2"
      # [[inputs.ctrlx_datalayer.subscription.nodes]]
      #    name   ="used"
      #    address="framework/metrics/system/memused-mb"

      ## The switch "output_json_string" enables output of the measurement as json. 
      ## That way it can be used in in a subsequent processor plugin, e.g. "Starlark Processor Plugin".
      # output_json_string = false

      ## Define extra tags related to subscription to be added to the output metric (optional)
      # [inputs.ctrlx_datalayer.subscription.tags]
      #    subscription_tag1 = "subscription_tag1"
      #    subscription_tag2 = "subscription_tag2"

      ## The interval in which messages shall be sent by the ctrlX Data Layer to this plugin. (default: 1s)
      ## Higher values reduce load on network by queuing samples on server side and sending as a single TCP packet.
      # publish_interval = "1s"

      ## The interval a "keepalive" message is sent if no change of data occurs. (default: 60s)
      ## Only used internally to detect broken network connections.
      # keep_alive_interval = "60s"

      ## The interval an "error" message is sent if an error was received from a node. (default: 10s)
      ## Higher values reduce load on output target and network in case of errors by limiting frequency of error messages.
      # error_interval = "10s"

      ## The interval that defines the fastest rate at which the node values should be sampled and values captured. (default: 1s)
      ## The sampling frequency should be adjusted to the dynamics of the signal to be sampled.
      ## Higher sampling frequence increases load on ctrlX Data Layer.
      ## The sampling frequency can be higher, than the publish interval. Captured samples are put in a queue and sent in publish interval.
      ## Note: The minimum sampling interval can be overruled by a global setting in the ctrlX Data Layer configuration ('datalayer/subscriptions/settings').
      # sampling_interval = "1s"

      ## The requested size of the node value queue. (default: 10)
      ## Relevant if more values are captured than can be sent.
      # queue_size = 10

      ## The behaviour of the queue if it is full. (default: "DiscardOldest")
      ## Possible values: 
      ## - "DiscardOldest"
      ##   The oldest value gets deleted from the queue when it is full.
      ## - "DiscardNewest"
      ##   The newest value gets deleted from the queue when it is full.
      # queue_behaviour = "DiscardOldest"

      ## The filter when a new value will be sampled. (default: 0.0)
      ## Calculation rule: If (abs(lastCapturedValue - newValue) > dead_band_value) capture(newValue).
      # dead_band_value = 0.0

      ## The conditions on which a sample should be captured and thus will be sent as a message. (default: "StatusValue")
      ## Possible values:
      ## - "Status"
      ##   Capture the value only, when the state of the node changes from or to error state. Value changes are ignored.
      ## - "StatusValue" 
      ##   Capture when the value changes or the node changes from or to error state.
      ##   See also 'dead_band_value' for what is considered as a value change.
      ## - "StatusValueTimestamp": 
      ##   Capture even if the value is the same, but the timestamp of the value is newer.
      ##   Note: This might lead to high load on the network because every sample will be sent as a message
      ##   even if the value of the node did not change.
      # value_change = "StatusValue"
      
```

## Metrics

All measurements are tagged with the server address of the device and the
corresponding node address as defined in the ctrlX Data Layer.

- measurement name
  - tags:
    - `source` (ctrlX Data Layer server where the metrics are gathered from)
    - `node` (Address of the ctrlX Data Layer node)
  - fields:
    - `{name}` (for nodes with simple data types)
    - `{name}_{index}`(for nodes with array data types)
    - `{name}_{jsonflat.key}` (for nodes with object data types)

### Output Format

The switch "output_json_string" determines the format of the output metric.

#### Output default format

With the output default format

```toml
output_json_string=false
```

the output is formatted automatically as follows depending on the data type:

##### Simple data type

The value is passed 'as it is' to a metric with pattern:

```text
{name}={value}
```

Simple data types of ctrlX Data Layer:

```text
bool8,int8,uint8,int16,uint16,int32,uint32,int64,uint64,float,double,string,timestamp
```

##### Array data type

Every value in the array is passed to a metric with pattern:

```text
{name}_{index}={value[index]}
```

example:

```text
myarray=[1,2,3] -> myarray_1=1, myarray_2=2, myarray_3=3
```

Array data types of ctrlX Data Layer:

```text
arbool8,arint8,aruint8,arint16,aruint16,arint32,aruint32,arint64,aruint64,arfloat,ardouble,arstring,artimestamp
```

##### Object data type (JSON)

Every value of the flattened json is passed to a metric with pattern:

```text
{name}_{jsonflat.key}={jsonflat.value}
```

example:

```text
myobj={"a":1,"b":2,"c":{"d": 3}} -> myobj_a=1, myobj_b=2, myobj_c_d=3
```

#### Output JSON format

With the output JSON format

```toml
output_json_string=true
```

the output is formatted as JSON string:

```text
{name}="{value}"
```

examples:

```text
input=true -> output="true"
```

```text
input=[1,2,3] -> output="[1,2,3]"
```

```text
input={"x":4720,"y":9440,"z":{"d": 14160}} -> output="{\"x\":4720,\"y\":9440,\"z\":14160}"
```

The JSON output string can be passed to a processor plugin for transformation
e.g. [Parser Processor Plugin][PARSER.md]
or [Starlark Processor Plugin][STARLARK.md]

[PARSER.md]: ../../processors/parser/README.md
[STARLARK.md]: ../../processors/starlark/README.md

example:

```toml
[[inputs.ctrlx_datalayer.subscription]]
   measurement = "osci"
   nodes = [
     {address="oscilloscope/instances/Osci_PLC/rec-values/allsignals"},
   ]
   output_json_string = true

[[processors.starlark]]
   namepass = [
      'osci',
   ]
   script = "oscilloscope.star"
```

## Troubleshooting

This plugin was contributed by
[Bosch Rexroth](https://www.boschrexroth.com).
For questions regarding ctrlX AUTOMATION and this plugin feel
free to check out and be part of the
[ctrlX AUTOMATION Community](https://ctrlx-automation.com/community)
to get additional support or leave some ideas and feedback.

Also, join
[InfluxData Community Slack](https://influxdata.com/slack) or
[InfluxData Community Page](https://community.influxdata.com/)
if you have questions or comments for the telegraf engineering teams.

## Example Output

The plugin handles simple, array and object (JSON) data types.

### Example with simple data type

Configuration:

```toml
[[inputs.ctrlx_datalayer.subscription]]
   measurement="memory"
   [inputs.ctrlx_datalayer.subscription.tags]
      sub_tag1="memory_tag1"
      sub_tag2="memory_tag2"

   [[inputs.ctrlx_datalayer.subscription.nodes]]
      name   ="available"
      address="framework/metrics/system/memavailable-mb"
      [inputs.ctrlx_datalayer.subscription.nodes.tags]
         node_tag1="memory_available_tag1"
         node_tag2="memory_available_tag2"

   [[inputs.ctrlx_datalayer.subscription.nodes]]
      name   ="used"
      address="framework/metrics/system/memused-mb"
      [inputs.ctrlx_datalayer.subscription.nodes.tags]
         node_tag1="memory_used_node_tag1"
         node_tag2="memory_used_node_tag2"
```

Source:

```json
"framework/metrics/system/memavailable-mb" : 365.93359375
"framework/metrics/system/memused-mb" : 567.67578125
```

Metrics:

```text
memory,source=192.168.1.1,host=host.example.com,node=framework/metrics/system/memavailable-mb,node_tag1=memory_available_tag1,node_tag2=memory_available_tag2,sub_tag1=memory2_tag1,sub_tag2=memory_tag2 available=365.93359375 1680093310249627400
memory,source=192.168.1.1,host=host.example.com,node=framework/metrics/system/memused-mb,node_tag1=memory_used_node_tag1,node_tag2=memory_used_node_tag2,sub_tag1=memory2_tag1,sub_tag2=memory_tag2 used=567.67578125 1680093310249667600
```

### Example with array data type

Configuration:

```toml
[[inputs.ctrlx_datalayer.subscription]]
   measurement="array"
   nodes=[
      { name="ar_uint8", address="alldata/dynamic/array-of-uint8"},
      { name="ar_bool8", address="alldata/dynamic/array-of-bool8"},
   ]
```

Source:

```json
"alldata/dynamic/array-of-bool8" : [true, false, true]
"alldata/dynamic/array-of-uint8" : [0, 255]
```

Metrics:

```text
array,source=192.168.1.1,host=host.example.com,node=alldata/dynamic/array-of-bool8 ar_bool8_0=true,ar_bool8_1=false,ar_bool8_2=true 1680095727347018800
array,source=192.168.1.1,host=host.example.com,node=alldata/dynamic/array-of-uint8 ar_uint8_0=0,ar_uint8_1=255 1680095727347223300
```

### Example with object data type (JSON)

Configuration:

```toml
[[inputs.ctrlx_datalayer.subscription]]
   measurement="motion"
   nodes=[
      {name="linear", address="motion/axs/Axis_1/state/values/actual"},
      {name="rotational", address="motion/axs/Axis_2/state/values/actual"},
   ]
```

Source:

```json
"motion/axs/Axis_1/state/values/actual" : {"actualPos":65.249329860957,"actualVel":5,"actualAcc":0,"actualTorque":0,"distLeft":0,"actualPosUnit":"mm","actualVelUnit":"mm/min","actualAccUnit":"m/s^2","actualTorqueUnit":"Nm","distLeftUnit":"mm"}
"motion/axs/Axis_2/state/values/actual" : {"actualPos":120,"actualVel":0,"actualAcc":0,"actualTorque":0,"distLeft":0,"actualPosUnit":"deg","actualVelUnit":"rpm","actualAccUnit":"rad/s^2","actualTorqueUnit":"Nm","distLeftUnit":"deg"}
```

Metrics:

```text
motion,source=192.168.1.1,host=host.example.com,node=motion/axs/Axis_1/state/values/actual linear_actualVel=5,linear_distLeftUnit="mm",linear_actualAcc=0,linear_distLeft=0,linear_actualPosUnit="mm",linear_actualAccUnit="m/s^2",linear_actualTorqueUnit="Nm",linear_actualPos=65.249329860957,linear_actualVelUnit="mm/min",linear_actualTorque=0 1680258290342523500
motion,source=192.168.1.1,host=host.example.com,node=motion/axs/Axis_2/state/values/actual rotational_distLeft=0,rotational_actualVelUnit="rpm",rotational_actualAccUnit="rad/s^2",rotational_distLeftUnit="deg",rotational_actualPos=120,rotational_actualVel=0,rotational_actualAcc=0,rotational_actualPosUnit="deg",rotational_actualTorqueUnit="Nm",rotational_actualTorque=0 1680258290342538100
```

If `output_json_string` is set in the configuration:

```toml
  output_json_string = true
```

then the metrics will be generated like this:

```text
motion,source=192.168.1.1,host=host.example.com,node=motion/axs/Axis_1/state/values/actual linear="{\"actualAcc\":0,\"actualAccUnit\":\"m/s^2\",\"actualPos\":65.249329860957,\"actualPosUnit\":\"mm\",\"actualTorque\":0,\"actualTorqueUnit\":\"Nm\",\"actualVel\":5,\"actualVelUnit\":\"mm/min\",\"distLeft\":0,\"distLeftUnit\":\"mm\"}" 1680258290342523500
motion,source=192.168.1.1,host=host.example.com,node=motion/axs/Axis_2/state/values/actual rotational="{\"actualAcc\":0,\"actualAccUnit\":\"rad/s^2\",\"actualPos\":120,\"actualPosUnit\":\"deg\",\"actualTorque\":0,\"actualTorqueUnit\":\"Nm\",\"actualVel\":0,\"actualVelUnit\":\"rpm\",\"distLeft\":0,\"distLeftUnit\":\"deg\"}" 1680258290342538100
```
