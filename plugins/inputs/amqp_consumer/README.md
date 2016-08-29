# AMQP Consumer Input Plugin

This plugin reads data from an AMQP Queue ([RabbitMQ](https://www.rabbitmq.com/) being an example) formatted in one of the [Telegraf Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

The following defaults are set to work with RabbitMQ:

```
# AMQP consumer plugin
[[inputs.amqp_consumer]]
  # The following options form a connection string to amqp:
  # amqp://{username}:{password}@{amqp_host}:{amqp_port}
  username = "guest"
  password = "guest"
  amqp_host = "localhost"
  amqp_port = "5672"
  # name of the queue to consume from
  queue = "task_queue"

  data_format = "influx"
```