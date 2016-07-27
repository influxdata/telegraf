# Kafka Consumer Input Plugin

The [Kafka](http://kafka.apache.org/) consumer plugin polls a specified Kafka
topic and adds messages to InfluxDB. The plugin assumes messages follow the
line protocol. [Consumer Group](http://godoc.org/github.com/wvanbergen/kafka/consumergroup)
is used to talk to the Kafka cluster so multiple instances of telegraf can read
from the same topic in parallel.

Now supports kafka new consumer (version 0.9+) with TLS

## Configuration[0.8]

```toml
# Read metrics from Kafka topic(s)
[[inputs.kafka_consumer]]
  ## is new consumer?
  new_consumer = false
  ## topic(s) to consume
  topics = ["telegraf"]
  ## an array of Zookeeper connection strings
  zookeeper_peers = ["localhost:2181"]
  ## the name of the consumer group
  consumer_group = "telegraf_metrics_consumers"
  ## Maximum number of metrics to buffer between collection intervals
  metric_buffer = 100000
  ## Offset (must be either "oldest" or "newest")
  offset = "oldest"

  ## Data format to consume.

  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```



## Configuration[0.9+]

```toml
# Read metrics from Kafka topic(s)
[[inputs.kafka_consumer]]
  ## is new consumer?
  new_consumer = true
  ## topic(s) to consume
  topics = ["telegraf"]
  ## an array of kafka 0.9+ brokers
  broker_list = ["localhost:9092"]
  ## the name of the consumer group
  consumer_group = "telegraf_kafka_consumer_group"
  ## Offset (must be either "oldest" or "newest")
  offset = "oldest"
  
  ## Optional SSL Config
  ssl_ca = "/etc/telegraf/ca.pem"
  ssl_cert = "/etc/telegraf/cert.pem"
  ssl_key = "/etc/telegraf/cert.key"
  ## Use SSL but skip chain & host verification
  insecure_skip_verify = false

  ## Data format to consume.
  
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```



## Testing

Running integration tests requires running Zookeeper & Kafka. See Makefile
for kafka container command.
