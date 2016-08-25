# RabbitMQ Producer Output Plugin

This plugin writes to a [RabbitMQ Broker](https://www.rabbitmq.com/), acting as a RabbitMQ Producer

```
# An output for publishing to RabbitMQ
[[outputs.rabbitmq]]
  # The following options form a connection string to rabbitmq:
  # amqp://{username}:{password}@{rabbitmq_host}:{rabbitmq_port}
  username = "guest"
  password = "guest"
  rabbitmq_host = "localhost"
  rabbitmq_port = "5672"
  # name of the queue to publish to
  queue = "task_queue"

  data_format = "influx"
```