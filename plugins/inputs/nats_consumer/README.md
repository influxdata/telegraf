# NATS Consumer

The [NATS](http://www.nats.io/about/) consumer plugin reads from 
specified NATS subjects and adds messages to InfluxDB. The plugin expects messages
in the [Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/DATA_FORMATS_INPUT.md). 
A [Queue Group](http://www.nats.io/documentation/concepts/nats-queueing/)
is used when subscribing to subjects so multiple instances of telegraf can read
from a NATS cluster in parallel.

## Configuration
```
# Read metrics from NATS subject(s)
[[inputs.nats_consumer]]
  ### urls of NATS servers
  servers = ["nats://localhost:4222"]
  ### Use Transport Layer Security
  secure = false
  ### subject(s) to consume
  subjects = ["telegraf"]
  ### name a queue group
  queue_group = "telegraf_consumers"
  ### Maximum number of points to buffer between collection intervals
  point_buffer = 100000
  
  ### Data format to consume. This can be "json", "influx" or "graphite"
  ### Each data format has it's own unique set of configuration options, read
  ### more about them here:
  ### https://github.com/influxdata/telegraf/blob/master/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## Testing

To run tests:

```
go test
```