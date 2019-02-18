# Kafka Consumer Input Plugin

The [Kafka][kafka] consumer plugin reads from Kafka
and creates metrics using one of the supported [input data formats][].

For old kafka version (< 0.8), please use the [kafka_consumer_legacy][] input plugin
and use the old zookeeper connection method.

### Configuration

```toml
[[inputs.kafka_consumer]]
  ## kafka servers
  brokers = ["localhost:9092"]
  ## topic(s) to consume
  topics = ["telegraf"]
  ## Add topic as tag if topic_tag is not empty
  # topic_tag = ""

  ## Optional Client id
  # client_id = "Telegraf"

  ## Set the minimal supported Kafka version.  Setting this enables the use of new
  ## Kafka features and APIs.  Of particular interest, lz4 compression
  ## requires at least version 0.10.0.0.
  ##   ex: version = "1.1.0"
  # version = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional SASL Config
  # sasl_username = "kafka"
  # sasl_password = "secret"

  ## the name of the consumer group
  consumer_group = "telegraf_metrics_consumers"
  ## Offset (must be either "oldest" or "newest")
  offset = "oldest"

  ## Maximum length of a message to consume, in bytes (default 0/unlimited);
  ## larger messages are dropped
  max_message_len = 1000000

  ## Maximum messages to read from the broker that have not been written by an
  ## output.  For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

[kafka]: https://kafka.apache.org
[kafka_consumer_legacy]: /plugins/inputs/kafka_consumer_legacy/README.md
[input data formats]: /docs/DATA_FORMATS_INPUT.md
