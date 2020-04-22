# Pulsar Output Plugin

This plugin writes to a [Apache Pulsar](https://pulsar.apache.org/en/) acting as Pulsar Producer.

### Configuration:
```toml
## Configuration for apache pulsar  to send metrics to 
[[outputs.pulsar]]
 ## URLs of pulsar url
  url = "pulsar://localhost:6650"
 ## Pulsar topic for producer messages
  topic = "persistent://public/default/telegraf"

 ## The value of this tag will be used as the topic.  If not set the 'topic'
 ## option is used.
 # topic_tag = "foo"

 ## If true, the 'topic_tag' will be removed from to the metric.
 # exclude_topic_tag = false

  routing_tag = "host"

 ## The routing key is set as the message key and used to determine which
 ## partition to send the message to.  This value is only used when no
 ## routing_tag is set or as a fallback when the tag specified in routing tag
 ## is not found.
 ##
 ## If set to "random", a random value will be generated for each message.
 ##
 ## When unset, no message key is added and each message is routed to a random
 ## partition.
 ##
 ##   ex: routing_key = "random"
 ##       routing_key = "telegraf"
 # routing_key = ""

 ## Optional Authentication Provider Config Defaults to empty "" NoAuthentication
 ## if set to "token" provide the JWT token
 ## if set to "tls" the please mention tls_cert and tls_key
 # auth_provider = ""

 # For token auth provider
 # auth_token = ""

 # Set the following values for tls auth provider
 # insecure_skip_verify = false
 # tls_ca = ""
 # tls_validate_host_name = true
 # tls_cert = ""
 # tls_key = ""

 ## Optional timeout Config
 # connection_timeout = "30s"
 # operation_timeout = "30s"

 ## Optional Producer Config

 ## CompressionType represents the various compression codecs recognized by
 ## Pulsar in messages.
 ##  0 : No compression
 ##  1 : LZ4 compression
 ##  2 : ZLib compression
 ##  3 : ZSTD compression
 # compression_type = 0

 ## MaxPendingMessages set the max size of the queue holding the messages pending to receive an
 ## acknowledgment from the broker.
 # max_pending_messages = 1000

 ## BatchingMaxMessages set the maximum number of messages permitted in a batch. (default: metric_batch_size)
 ## If set to a value greater than 1, messages will be queued until this threshold is reached or
 ## batch interval has elapsed. By Default it is set to "metric_batch_size"
 # batching_max_messages = 1000

 ## HashingScheme change the "HashingScheme" used to chose the partition on where to publish a particular message.
 ##  JavaStringHash : JavaStringHash Hshing
 ##  Murmur3_32Hash : Murmur3_32Hash Hashing
 # hashing_scheme = "JavaStringHash"


 ## Data format to output.
 ## Each data format has its own unique set of configuration options, read
 ## more about them here:
 ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
 # data_format = "influx"

```
