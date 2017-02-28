# AMQP Consumer Input Plugin

This plugin reads data from an AMQP Queue ([RabbitMQ](https://www.rabbitmq.com/) being an example) formatted in one of
the [Telegraf Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

The following defaults are set to work with RabbitMQ:

```
# AMQP consumer plugin
[[inputs.amqp_consumer]]
  ## AMQP url
  url = "amqp://localhost:5672/influxdb"
  ## AMQP exchange
  exchange = "telegraf"
  ## Auth method. PLAIN and EXTERNAL are supported
  # auth_method = "PLAIN"
  ## Binding Key
  binding_key = "#"

  ## Maximum number of messages server should give to the worker.
  prefetch = 50

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
