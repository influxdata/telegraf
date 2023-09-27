# MQTT Consumer Input Plugin

The [MQTT][mqtt] consumer plugin reads from the specified MQTT topics
and creates metrics using one of the supported [input data formats][].

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `usernane` and
`password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

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

  ## Max undelivered messages
  ## This plugin uses tracking metrics, which ensure messages are read to
  ## outputs before acknowledging them to the original broker to ensure data
  ## is not lost. This option sets the maximum messages to read from the
  ## broker that have not been written by an output.
  ##
  ## This value needs to be picked with awareness of the agent's
  ## metric_batch_size value as well. Setting max undelivered messages too high
  ## can result in a constant stream of data batches to the output. While
  ## setting it too low may never flush the broker's messages.
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

  ## Client trace messages
  ## When set to true, and debug mode enabled in the agent settings, the MQTT
  ## client's messages are included in telegraf logs. These messages are very
  ## noisey, but essential for debugging issues.
  # client_trace = false

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

## Example Output

```text
mqtt_consumer,host=pop-os,topic=telegraf/host01/cpu value=45i 1653579140440951943
mqtt_consumer,host=pop-os,topic=telegraf/host01/cpu value=100i 1653579153147395661
```

## About Topic Parsing

The MQTT topic as a whole is stored as a tag, but this can be far too coarse to
be easily used when utilizing the data further down the line. This change allows
tag values to be extracted from the MQTT topic letting you store the information
provided in the topic in a meaningful way. An `_` denotes an ignored entry in
the topic path. Please see the following example.

### Topic Parsing Example

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

Will result in the following metric:

```text
cpu,host=pop-os,tag=telegraf,topic=telegraf/one/cpu/23 value=45,test=23i 1637014942460689291
```

## Field Pivoting Example

You can use the pivot processor to rotate single
valued metrics into a multi field metric.
For more info check out the pivot processors
[here][1].

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
        tags = "/_/site/version/device_name/field"
[[processors.pivot]]
    tag_key = "field"
    value_key = "value"
```

Will result in the following metric:

```text
sensors,site=CLE,version=v1,device_name=device5 temp=390,rpm=45.0,ph=1.45
```

[1]: <https://github.com/influxdata/telegraf/tree/master/plugins/processors/pivot> "Pivot Processor"

## Metrics

- All measurements are tagged with the incoming topic, ie
`topic=telegraf/host01/cpu`

- example when [[inputs.mqtt_consumer.topic_parsing]] is set

- when [[inputs.internal]] is set:
  - payload_size (int): get the cumulative size in bytes that have been received from incoming messages
  - messages_received (int): count of the number of messages that have been received from mqtt

This will result in the following metric:

```text
internal_mqtt_consumer host=pop-os version=1.24.0 messages_received=622i payload_size=37942i 1657282270000000000
```

[mqtt]: https://mqtt.org
[input data formats]: /docs/DATA_FORMATS_INPUT.md
