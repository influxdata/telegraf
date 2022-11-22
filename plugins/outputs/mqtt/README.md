# MQTT Producer Output Plugin

This plugin writes to a [MQTT Broker](http://http://mqtt.org/) acting as a mqtt
Producer. It supports MQTT protocols `3.1.1` and `5`.

## Mosquitto v2.0.12+ and `identifier rejected`

In v2.0.12+ of the mosquitto MQTT server, there is a
[bug](https://github.com/eclipse/mosquitto/issues/2117) which requires the
`keep_alive` value to be set non-zero in your telegraf configuration. If not
set, the server will return with `identifier rejected`.

As a reference `eclipse/paho.golang` sets the `keep_alive` to 30.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Configuration for MQTT server to send metrics to
[[outputs.mqtt]]
  ## MQTT Brokers
  ## The list of brokers should only include the hostname or IP address and the
  ## port to the broker. This should follow the format `[{scheme}://]{host}:{port}`. For
  ## example, `localhost:1883` or `mqtt://localhost:1883`.
  ## Scheme can be any of the following: tcp://, mqtt://, tls://, mqtts://
  ## non-TLS and TLS servers can not be mix-and-matched.
  servers = ["localhost:1883", ] # or ["mqtts://tls.example.com:1883"]

  ## Protocol can be `3.1.1` or `5`. Default is `3.1.1`
  # procotol = "3.1.1"

  ## MQTT Topic for Producer Messages
  ## MQTT outputs send metrics to this topic format:
  ## <topic_prefix>/<hostname>/<pluginname>/ (e.g. prefix/web01.example.com/mem)
  topic_prefix = "telegraf"

  ## QoS policy for messages
  ## The mqtt QoS policy for sending messages.
  ## See https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_9.0.0/com.ibm.mq.dev.doc/q029090_.htm
  ##   0 = at most once
  ##   1 = at least once
  ##   2 = exactly once
  # qos = 2

  ## Keep Alive
  ## Defines the maximum length of time that the broker and client may not
  ## communicate. Defaults to 0 which turns the feature off.
  ##
  ## For version v2.0.12 and later mosquitto there is a bug
  ## (see https://github.com/eclipse/mosquitto/issues/2117), which requires
  ## this to be non-zero. As a reference eclipse/paho.mqtt.golang defaults to 30.
  # keep_alive = 0

  ## username and password to connect MQTT server.
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"

  ## client ID
  ## The unique client id to connect MQTT server. If this parameter is not set
  ## then a random ID is generated.
  # client_id = ""

  ## Timeout for write operations. default: 5s
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## When true, metrics will be sent in one MQTT message per flush. Otherwise,
  ## metrics are written one metric per MQTT message.
  # batch = false

  ## When true, metric will have RETAIN flag set, making broker cache entries until someone
  ## actually reads it
  # retain = false

  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```
