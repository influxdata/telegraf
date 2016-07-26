# Kafka New Consumer Input Plugin

The [Kafka](http://kafka.apache.org/) consumer plugin polls a specified Kafka
topic and adds messages to InfluxDB. The plugin assumes messages follow the
line protocol. [Sarama-cluster](http://godoc.org/github.com/bsm/sarama-cluster)
is used to talk to the Kafka cluster so multiple instances of telegraf can read
from the same topic in parallel.

This plugin is compatible with Kafka 0.9+

### Configuration:

```toml
# Description
[[inputs.kafka_consumer_plus]]
    ## topic(s) to consume
    topics = ["telegraf"]
    ## an array of kafka 0.9+ brokers
    broker_list = ["localhost:9092"]
    ## the name of the consumer group
    consumer_group = "telegraf_kafka_consumer_group"
    ## from beginning
    from_beginning = true

    ## Optional SSL Config
    # ssl_ca = "/etc/telegraf/ca.pem"
    ssl_cert = "/etc/telegraf/cert.pem"
    ssl_key = "/etc/telegraf/cert.key"
    ## Use SSL but skip chain & host verification
    insecure_skip_verify = true

    ## Data format to consume.
    ## Each data format has it's own unique set of configuration options, read
    ## more about them here:
    ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
    data_format = "json"
  
```

## Testing

Running integration tests requires running Zookeeper & Kafka. See Makefile
for kafka container command.