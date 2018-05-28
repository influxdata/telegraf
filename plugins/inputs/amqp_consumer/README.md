# AMQP Consumer Input Plugin

This plugin provides a consumer for use with AMQP 0-9-1, a promenent implementation of this protocol being [RabbitMQ](https://www.rabbitmq.com/).

Metrics are read from a topic exchange using the configured queue and binding_key.

Message payload should be formatted in one of the [Telegraf Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

For an introduction to AMQP see:
- https://www.rabbitmq.com/tutorials/amqp-concepts.html
- https://www.rabbitmq.com/getstarted.html

The following defaults are known to work with RabbitMQ:

```toml
# AMQP consumer plugin
[[inputs.amqp_consumer]]
  ## AMQP url
  url = "amqp://localhost:5672/influxdb"
  ## AMQP exchange
  exchange = "telegraf"
  ## AMQP queue name
  queue = "telegraf"
  ## Binding Key
  binding_key = "#"

  ## Controls how many messages the server will try to keep on the network
  ## for consumers before receiving delivery acks.
  #prefetch_count = 50

  ## Auth method. PLAIN and EXTERNAL are supported.
  ## Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
  ## described here: https://www.rabbitmq.com/plugins.html
  # auth_method = "PLAIN"
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
```
