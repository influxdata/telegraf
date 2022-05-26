# MQTT Consumer Input Plugin

The [MQTT][mqtt] consumer plugin reads from the specified MQTT topics
and creates metrics using one of the supported [input data formats][].

## Configuration

```toml @sample.conf
# Read metrics from MQTT topic(s)
[[inputs.mqtt_consumer]]
  ## Broker URLs for the MQTT server or cluster.  To connect to multiple
  ## clusters or standalone servers, use a separate plugin instance.
  ##   example: servers = ["tcp://localhost:1883"]
  ##            servers = ["ssl://localhost:1883"]
  ##            servers = ["ws://localhost:1883"]
  servers = ["tcp://127.0.0.1:1883"]

  ## Topics that will be subscribed to.
  topics = [
    "telegraf/host01/cpu",
    "telegraf/+/mem",
    "sensors/#",
  ]

  ## The message topic will be stored in a tag specified by this value.  If set
  ## to the empty string no topic tag will be created.
  # topic_tag = "topic"

  ## QoS policy for messages
  ##   0 = at most once
  ##   1 = at least once
  ##   2 = exactly once
  ##
  ## When using a QoS of 1 or 2, you should enable persistent_session to allow
  ## resuming unacknowledged messages.
  # qos = 0

  ## Connection timeout for initial connection in seconds
  # connection_timeout = "30s"

  ## Maximum messages to read from the broker that have not been written by an
  ## output.  For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Persistent session disables clearing of the client session on connection.
  ## In order for this option to work you must also set client_id to identify
  ## the client.  To receive messages that arrived while the client is offline,
  ## also set the qos option to 1 or 2 and don't forget to also set the QoS when
  ## publishing.
  # persistent_session = false

  ## If unset, a random client ID will be generated.
  # client_id = ""

  ## Username and password to connect MQTT server.
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Enable extracting tag values from MQTT topics
  ## _ denotes an ignored entry in the topic path
  # [[inputs.mqtt_consumer.topic_parsing]]
  #   topic = ""
  #   measurement = ""
  #   tags = ""
  #   fields = ""
  ## Value supported is int, float, unit
  #   [[inputs.mqtt_consumer.topic.types]]
  #      key = type
```

## About Topic Parsing

The MQTT topic as a whole is stored as a tag, but this can be far too coarse
to be easily used when utilizing the data further down the line. This
change allows tag values to be extracted from the MQTT topic letting you
store the information provided in the topic in a meaningful way.
An `_` denotes an ignored entry in the topic path. 
Please see the following example.

## Example Configuration for topic parsing

```toml
[[inputs.mqtt_consumer]]
  ## Broker URLs for the MQTT server or cluster.  To connect to multiple
  ## clusters or standalone servers, use a separate plugin instance.
  ##   example: servers = ["tcp://localhost:1883"]
  ##            servers = ["ssl://localhost:1883"]
  ##            servers = ["ws://localhost:1883"]
  servers = ["tcp://127.0.0.1:1883"]

  ## Topics that will be subscribed to.
  topics = [
    "telegraf/+/cpu/23",
  ]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "value"
  data_type = "float"

  [[inputs.mqtt_consumer.topic_parsing]]
    topic = "telegraf/one/cpu/23"
    measurement = "_/_/measurement/_"
    tags = "tag/_/_/_"
    fields = "_/_/_/test"
    [inputs.mqtt_consumer.topic_parsing.types]
      test = "int"
```

Result:

```shell
cpu,host=pop-os,tag=telegraf,topic=telegraf/one/cpu/23 value=45,test=23i 1637014942460689291
```

### Field Pivoting with Eaxmple

You can use the pivot processor to rotate single
valued metrics into a multi field metric.
For more info check out the pivot processors 
[here](https://github.com/influxdata/telegraf/tree/master/plugins/processors/pivot).

For this example these are the topics:

```text
/sensors/CLE/v1/device5/temp
/sensors/CLE/v1/device5/rpm
/sensors/CLE/v1/device5/ph
/sensors/CLE/v1/device5/spin
```

And these are the metrics:

```text
sensors,site=CLE,version=v1,device_name=device5,field=temp value=390
sensors,site=CLE,version=v1,device_name=device5,field=rpm value=45.0
sensors,site=CLE,version=v1,device_name=device5,field=ph value=1.45
```

Using pivot in the config will rotate the metrics into a multi field metric.
The config:

```toml
[[inputs.mqtt_consumer]]
    ....
    topics = "/sensors/#"
    [[inputs.mqtt_consumer.topic_parsing]]
        measurement = "/measurement/_/_/_/_"
        tags = "/_/site/version/device_name/_"
        fields = "/_/_/_/_/field"
[[processors.pivot]]
    tag_key = "field"
    value_key = "value"
```

Result:

```text
sensors,site=CLE,version=v1,device_name=device5 temp=390,rpm=45.0,ph=1.45
```

## Example Output

```
mqtt_consumer,host=pop-os,topic=telegraf/host01/cpu value=45i 1653579140440951943
mqtt_consumer,host=pop-os,topic=telegraf/host01/cpu value=100i 1653579153147395661
```

## Metrics

- All measurements are tagged with the incoming topic, ie
`topic=telegraf/host01/cpu`

- example when [[inputs.mqtt_consumer.topic_parsing]] is set

[mqtt]: https://mqtt.org
[input data formats]: /docs/DATA_FORMATS_INPUT.md
