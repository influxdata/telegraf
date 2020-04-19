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
 ## if set to "tls" the please mention tls_certificate_path and tls_private_key_path
 # auth_provider = ""

 # For token auth provider
 # auth_token = ""
 # Set the following values for tls auth provider
 # tls_allow_insecure_connection = false
 # tls_trust_certs_file_path = ""
 # tls_validate_host_name = true
 # tls_certificate_path = ""
 # tls_private_key_path = ""

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

 ## BatchingMaxPublishDelay set the time period within which the messages sent will be batched (default: 10ms)
 ## if batch messages are enabled. If set to a non zero value, messages will be queued until this time
 ## interval or until
 # batching_max_publish_delay = "10ms"

 ## BatchingMaxMessages set the maximum number of messages permitted in a batch. (default: 1000)
 ## If set to a value greater than 1, messages will be queued until this threshold is reached or
 ## batch interval has elapsed.
 # batching_max_messages = 1000

 ## HashingScheme change the "HashingScheme" used to chose the partition on where to publish a particular message.
 ##  JavaStringHash : JavaStringHash Hshing
 ##  Murmur3_32Hash : Murmur3_32Hash Hashing
 # hashing_scheme = "JavaStringHash"

 ## Disable batching will reduce the throughput
 # disable_batching = false


 ## Data format to output.
 ## Each data format has its own unique set of configuration options, read
 ## more about them here:
 ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
 # data_format = "influx"

 ## Optional topic suffix configuration.
 ## If the section is omitted, no suffix is used.
 ## Following topic suffix methods are supported:
 ##   measurement - suffix equals to separator + measurement's name
 ##   tags        - suffix equals to separator + specified tags' values
 ##                 interleaved with separator
 ## The routing tag specifies a tagkey on the metric whose value is used as
 ## the message key.  The message key is used to determine which partition to
 ## send the message to.  This tag is prefered over the routing_key option.

 ## Suffix equals to "_" + measurement name
 #  [outputs.pulsar.topic_suffix]
 #    method = "measurement"
 #    separator = "-"

 ## Suffix equals to "__" + measurement's "foo" tag value.
 ##   If there's no such a tag, suffix equals to an empty string
 #  [outputs.pulsar.topic_suffix]
 #    method = "tags"
 #    keys = ["foo"]
 #    separator = "-"

 ## Suffix equals to "_" + measurement's "foo" and "bar"
 ##   tag values, separated by "_". If there is no such tags,
 ##   their values treated as empty strings.
 #  [outputs.pulsar.topic_suffix]
 #    method = "tags"
 #    keys = ["foo","bar"]
 #    separator = "-"


```
