# AMQP Output Plugin

This plugin writes to a AMQP 0-9-1 Exchange, a promenent implementation of this protocol being [RabbitMQ](https://www.rabbitmq.com/).

Metrics are written to a topic exchange using tag, defined in configuration file as RoutingTag, as a routing key.

If RoutingTag is empty, then empty routing key will be used.
Metrics are grouped in batches by RoutingTag.

This plugin doesn't bind exchange to a queue, so it should be done by consumer.

For an introduction to AMQP see:
- https://www.rabbitmq.com/tutorials/amqp-concepts.html
- https://www.rabbitmq.com/getstarted.html

### Configuration:

```
# Configuration for the AMQP server to send metrics to
[[outputs.amqp]]
  ## AMQP url
  url = "amqp://localhost:5672/influxdb"
  ## AMQP exchange
  exchange = "telegraf"
  ## Auth method. PLAIN and EXTERNAL are supported
  ## Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
  ## described here: https://www.rabbitmq.com/plugins.html
  # auth_method = "PLAIN"
  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, its value will be used as the routing key
  routing_tag = "host"
  ## Delivery Mode controls if a published message is persistent
  ## Valid options are "transient" and "persistent". default: "transient"
  # delivery_mode = "transient"

  ## InfluxDB retention policy
  # retention_policy = "default"
  ## InfluxDB database
  # database = "telegraf"

  ## Write timeout, formatted as a string.  If not provided, will default
  ## to 5s. 0s means no timeout (not recommended).
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```
