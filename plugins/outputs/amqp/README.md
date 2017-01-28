# AMQP Output Plugin

This plugin writes to a AMQP exchange using tag, defined in configuration file
as RoutingTag, as a routing key.

If RoutingTag is empty, then empty routing key will be used.
Metrics are grouped in batches by RoutingTag.

This plugin doesn't bind exchange to a queue, so it should be done by consumer.

### Configuration:

```
# Configuration for the AMQP server to send metrics to
[[outputs.amqp]]
  ## AMQP url
  url = "amqp://localhost:5672/influxdb"
  ## AMQP exchange
  exchange = "telegraf"
  ## Auth method. PLAIN and EXTERNAL are supported
  # auth_method = "PLAIN"
  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, it's value will be used as the routing key
  routing_tag = "host"

  ## InfluxDB retention policy
  # retention_policy = "default"
  ## InfluxDB database
  # database = "telegraf"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```
