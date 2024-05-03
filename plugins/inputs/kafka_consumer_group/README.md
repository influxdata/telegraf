# Kafka Consumer Group Input Plugin

The [Kafka][kafka] consumer group plugin reads from Kafka

## Configuration

```toml
[[inputs.kafka_consumer_group]]
  ## Kafka brokers.
  brokers = ["localhost:9092"]

  ## Consumer Groups to monitor.
  consumer_groups = ["telegraf"]
```

[kafka]: https://kafka.apache.org
