# NSQ Consumer Input Plugin

The [NSQ](http://nsq.io/) consumer plugin polls a specified NSQD
topic and adds messages to InfluxDB. This plugin allows a message to be in any of the supported `data_format` types. 

## Configuration

```toml
# Read metrics from NSQD topic(s)
[[inputs.nsq_consumer]]
  ## An array of NSQD HTTP API endpoints
  server = "localhost:4150"
  topic = "telegraf"
  channel = "consumer"
  max_in_flight = 100

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## Testing
The `nsq_consumer_test` mocks out the interaction with `NSQD`. It requires no outside dependencies.
