# RabbitMQ Consumer Input Plugin

This plugin reads data from a RabbitMQ Queue formatted in one of the [Telegraf Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

```
# RabbitMQ consumer plugin
[[inputs.rabbitmq_consumer]]
  # The following options form a connection string to rabbitmq:
  # amqp://{username}:{password}@{rabbitmq_host}:{rabbitmq_port}
  username = "guest"
  password = "guest"
  rabbitmq_host = "localhost"
  rabbitmq_port = "5672"
  # name of the queue to consume from
  queue = "task_queue"

  data_format = "influx"
```