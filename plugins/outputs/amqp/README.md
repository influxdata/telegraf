# AMQP Output Plugin

This plugin writes to a AMQP 0-9-1 Exchange, a promenent implementation of this protocol being [RabbitMQ](https://www.rabbitmq.com/).

This plugin does not bind the exchange to a queue.

For an introduction to AMQP see:
- https://www.rabbitmq.com/tutorials/amqp-concepts.html
- https://www.rabbitmq.com/getstarted.html

### Configuration:
```toml
# Publishes metrics to an AMQP broker
[[outputs.amqp]]
  ## Broker to publish to.
  ##   deprecated in 1.7; use the brokers option
  # url = "amqp://localhost:5672/influxdb"

  ## Brokers to publish to.  If multiple brokers are specified a random broker
  ## will be selected anytime a connection is established.  This can be
  ## helpful for load balancing when not using a dedicated load balancer.
  brokers = ["amqp://localhost:5672/influxdb"]

  ## Maximum messages to send over a connection.  Once this is reached, the
  ## connection is closed and a new connection is made.  This can be helpful for
  ## load balancing when not using a dedicated load balancer.
  # max_messages = 0

  ## Exchange to declare and publish to.
  exchange = "telegraf"

  ## Exchange type; common types are "direct", "fanout", "topic", "header", "x-consistent-hash".
  # exchange_type = "topic"

  ## If true, exchange will be passively declared.
  # exchange_declare_passive = false

  ## Exchange durability can be either "transient" or "durable".
  # exchange_durability = "durable"

  ## Additional exchange arguments.
  # exchange_arguments = { }
  # exchange_arguments = {"hash_propery" = "timestamp"}

  ## Authentication credentials for the PLAIN auth_method.
  # username = ""
  # password = ""

  ## Auth method. PLAIN and EXTERNAL are supported
  ## Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
  ## described here: https://www.rabbitmq.com/plugins.html
  # auth_method = "PLAIN"

  ## Metric tag to use as a routing key.
  ##   ie, if this tag exists, its value will be used as the routing key
  # routing_tag = "host"

  ## Static routing key.  Used when no routing_tag is set or as a fallback
  ## when the tag specified in routing tag is not found.
  # routing_key = ""
  # routing_key = "telegraf"

  ## Delivery Mode controls if a published message is persistent.
  ##   One of "transient" or "persistent".
  # delivery_mode = "transient"

  ## InfluxDB database added as a message header.
  ##   deprecated in 1.7; use the headers option
  # database = "telegraf"

  ## InfluxDB retention policy added as a message header
  ##   deprecated in 1.7; use the headers option
  # retention_policy = "default"

  ## Static headers added to each published message.
  # headers = { }
  # headers = {"database" = "telegraf", "retention_policy" = "default"}

  ## Connection timeout.  If not provided, will default to 5s.  0s means no
  ## timeout (not recommended).
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## If true use batch serialization format instead of line based delimiting.
  ## Only applies to data formats which are not line based such as JSON.
  ## Recommended to set to true.
  # use_batch_format = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
```

#### Routing

If `routing_tag` is set, and the tag is defined on the metric, the value of
the tag is used as the routing key.  Otherwise the value of `routing_key` is
used directly.  If both are unset the empty string is used.

Exchange types that do not use a routing key, `direct` and `header`, always
use the empty string as the routing key.

Metrics are published in batches based on the final routing key.
