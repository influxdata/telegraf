# AMQP Output Plugin

This plugin writes to a AMQP 0-9-1 Exchange, a promenent implementation of this protocol being [RabbitMQ](https://www.rabbitmq.com/).

Metrics are written to a topic exchange using a routing key defined by:
1. The routing_key config defines a static value
2. The routing_tag config defines a metric tag with a dynamic value, overriding the static routing_key if found
3. If neither option is defined, or the tag is not found in a metric, then the empty routing key will be used

Metrics are grouped in batches by the final routing key.

This plugin doesn't bind exchange to a queue, so it should be done by consumer. The exchange is always defined as type: topic.
To use it for distributing metrics equally among workers (type: direct), set the routing_key to a static value on the exchange,
declare and bind a single queue with the same routing_key, and consume from the same queue in each worker.
To use it to send metrics to many consumers at once (type: fanout), set the routing_key to "#" on the exchange, then declare, bind,
and consume from individual queues in each worker.

For an introduction to AMQP see:
- https://www.rabbitmq.com/tutorials/amqp-concepts.html
- https://www.rabbitmq.com/getstarted.html

### Configuration:

```toml
# Configuration for the AMQP server to send metrics to
[[outputs.amqp]]
  ## Broker to publish to.
  ##   deprecated in 1.7; use the brokers option
  # url = "amqp://localhost:5672/influxdb"

  ## Brokers to publish to.  If multiple brokers are specified a random broker
  ## will be selected anytime a connection is established.  This can be
  ## helpful for load balancing when not using a dedicated load balancer.
  brokers = ["amqp://localhost:5672/influxdb"]

  ## Authentication credentials for the PLAIN auth_method.
  # username = ""
  # password = ""

  ## Exchange to declare and publish to.
  exchange = "telegraf"

  ## Exchange type; common types are "direct", "fanout", "topic", "header", "x-consistent-hash".
  # exchange_type = "topic"

  ## If true, exchange will be passively declared.
  # exchange_passive = false

  ## Exchange durability can be either "transient" or "durable".
  # exchange_durability = "durable"

  ## Additional exchange arguments.
  # exchange_args = { }
  # exchange_args = {"hash_propery" = "timestamp"}

  ## Auth method. PLAIN and EXTERNAL are supported
  ## Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
  ## described here: https://www.rabbitmq.com/plugins.html
  # auth_method = "PLAIN"
  ## Topic routing key
  # routing_key = ""
  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, its value will be used as the routing key
  ##  and override routing_key config even if defined
  routing_tag = "host"
  ## Delivery Mode controls if a published message is persistent
  ## Valid options are "transient" and "persistent". default: "transient"
  delivery_mode = "transient"

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
