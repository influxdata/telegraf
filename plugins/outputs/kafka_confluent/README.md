# Kafka Confluent Producer Output Plugin

This plugin writes to a [Librdkafka](https://docs.confluent.io/2.0.0/clients/librdkafka/index.html) acting as a Kafka Producer.

```
[[outputs.kafka]]
  ## URLs of kafka brokers
  brokers = "localhost:9092,localhost:9093"
  ## Kafka topic for producer messages
  topic = "telegraf"

```

### Required parameters:

* `brokers`: A string with all kafka brokers. URL should just include host and port e.g. -> `"{host}:{port},{host2}:{port2}"`
* `topic`: The `kafka` topic to publish to.