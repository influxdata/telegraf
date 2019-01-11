# Apache Pulsar Output Plugin

This plugin writes to Apache Pulsar. The sample configuration is:

```
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## URL to Pulsar cluster
  ## If you use SSL, then the protocol should be "pulsar+ssl"
  url = "pulsar://localhost:6650"

  ## Number of threads to be used for handling connections to brokers
  # iothreads = 1

  ## Producer-create, subscribe and unsubscribe operations will be retried until this interval, after which the
  ## operation will be marked as failed
  # operation_timeout = "30s"

  ## Set the number of threads to be used for message listeners
  # message_listener_threads = 1

  ## Number of concurrent lookup-requests allowed to send on each broker-connection to prevent overload on broker.
  ## It should be configured with higher value only in case of it requires to produce/subscribe
  ## on thousands of topic using created Pulsar Client
  # concurrent_lookup_requests = 5000

  ## Set the path to the trusted TLS certificate file
  # tls_trust_certs_path = ""

  ## Configure whether the Pulsar client accept untrusted TLS certificate from broker
  # tls_allow_insecure_connection = false

  ## Set the interval between each stat info. Stats will be activated with positive
  ## stats_interval_in_seconds It should be set to at least 1 second
  # stats_interval = 60

  ## Configure the authentication provider.
  # [auth]

  ## Create new Athenz Authentication provider with configuration in JSON form
  #   athenz = ""

  ## Create new Authentication provider with specified TLS certificate and private key
  #   cert_path = ""
  #   key_path = ""

  ## Configure the producer
  [producer]

  ## Set topic of the message, required
	topic = ""

  ## Specify a name for the producer
  ## If not assigned, the system will generate a globally unique name.
  ## When specifying a name, it is up to the user to ensure that, for a given topic, the producer name is unique
  ## across all Pulsar's clusters. Brokers will enforce that only a single producer a given name can be publishing on
  ## a topic.
  #    name = ""
  
  ## Attach a set of application defined properties to the producer
  ## This properties will be visible in the topic stats
  #    properties = { foo = "bar" }

  ## Set the send timeout.
  ## If a message is not acknowledged by the server before the send_timeout expires, an error will be reported.
  #    send_timeout = "30s"

  ## Set the max size of the queue holding the messages pending to receive an acknowledgment from the broker.
  ## When the queue is full, by default, all calls will fail unless block_if_queue_full is set to true.
  #    max_pending_messages = 64

  ## Set the number of max pending messages across all the partitions
  ## This setting will be used to lower the max pending messages for each partition, if the total exceeds the configured value.
  #    max_pending_messages_across_partitions = 512

  ## Set whether the send operations should block when the outgoing message queue is full. If set to false, send operations will immediately fail
  ## when there is no space left in pending queue.
  #    block_if_queue_full = false

  ## Set the message routing mode for the partitioned producer.
  ## 0 = Round robin
  ## 1 = Use single partition
  ## 2 = Custom partition
  #    message_routing_mode = 0

  ## Change the hashing scheme used to chose the partition on where to publish a particular message.
  ## 0 = Java String.hashCode() equivalent
  ## 1 = Use Murmur3 hashing function
  ## 2 = C++ based boost::hash
  #    hashing_scheme = 0

  ## Set the compression type for the producer.
  ## 0 = No compression
  ## 1 = LZ4
  ## 2 = ZLIB
  #    compression_type = 0

  ## Control whether automatic batching of messages is enabled for the producer.
  #    batching = false

  ## Set the time period within which the messages sent will be batched if batch messages are
  ## enabled. If set, messages will be queued until this time interval or until.
  #    batching_max_publish_delay = "10ms"

  ## Set the maximum number of messages permitted in a batch. If set,
  ## messages will be queued until this threshold is reached or batch interval has elapsed
  #    batching_max_messages = 1000
```
