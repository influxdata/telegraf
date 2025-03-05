# Apache Kafka Consumer Input Plugin

This service plugin consumes messages from [Kafka brokers][kafka] in one of the
supported [data formats][data_formats]. The plugin uses
[consumer groups][consumer_groups] when talking to the Kafka cluster so multiple
instances of Telegraf can consume messages from the same topic in parallel.

⭐ Telegraf v0.2.3
🏷️ messaging
💻 all

[kafka]: https://kafka.apache.org
[consumer_groups]: http://godoc.org/github.com/wvanbergen/kafka/consumergroup
[data_formats]: /docs/DATA_FORMATS_INPUT.md

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Startup error behavior options <!-- @/docs/includes/startup_error_behavior.md -->

In addition to the plugin-specific and global configuration settings the plugin
supports options for specifying the behavior when experiencing startup errors
using the `startup_error_behavior` setting. Available values are:

- `error`:  Telegraf with stop and exit in case of startup errors. This is the
            default behavior.
- `ignore`: Telegraf will ignore startup errors for this plugin and disables it
            but continues processing for all other plugins.
- `retry`:  Telegraf will try to startup the plugin in every gather or write
            cycle in case of startup errors. The plugin is disabled until
            the startup succeeds.

## Secret-store support

This plugin supports secrets from secret-stores for the `sasl_username`,
`sasl_password` and `sasl_access_token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Read metrics from Kafka topics
[[inputs.kafka_consumer]]
  ## Kafka brokers.
  brokers = ["localhost:9092"]

  ## Set the minimal supported Kafka version. Should be a string contains
  ## 4 digits in case if it is 0 version and 3 digits for versions starting
  ## from 1.0.0 separated by dot. This setting enables the use of new
  ## Kafka features and APIs.  Must be 0.10.2.0(used as default) or greater.
  ## Please, check the list of supported versions at
  ## https://pkg.go.dev/github.com/Shopify/sarama#SupportedVersions
  ##   ex: kafka_version = "2.6.0"
  ##   ex: kafka_version = "0.10.2.0"
  # kafka_version = "0.10.2.0"

  ## Topics to consume.
  topics = ["telegraf"]

  ## Topic regular expressions to consume.  Matches will be added to topics.
  ## Example: topic_regexps = [ "*test", "metric[0-9A-z]*" ]
  # topic_regexps = [ ]

  ## When set this tag will be added to all metrics with the topic as the value.
  # topic_tag = ""

  ## The list of Kafka message headers that should be pass as metric tags
  ## works only for Kafka version 0.11+, on lower versions the message headers
  ## are not available
  # msg_headers_as_tags = []

  ## The name of kafka message header which value should override the metric name.
  ## In case when the same header specified in current option and in msg_headers_as_tags
  ## option, it will be excluded from the msg_headers_as_tags list.
  # msg_header_as_metric_name = ""

  ## Set metric(s) timestamp using the given source.
  ## Available options are:
  ##   metric -- do not modify the metric timestamp
  ##   inner  -- use the inner message timestamp (Kafka v0.10+)
  ##   outer  -- use the outer (compressed) block timestamp (Kafka v0.10+)
  # timestamp_source = "metric"

  ## Optional Client id
  # client_id = "Telegraf"

  ## Optional TLS Config
  # enable_tls = false
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Period between keep alive probes.
  ## Defaults to the OS configuration if not specified or zero.
  # keep_alive_period = "15s"

  ## SASL authentication credentials.  These settings should typically be used
  ## with TLS encryption enabled
  # sasl_username = "kafka"
  # sasl_password = "secret"

  ## Optional SASL:
  ## one of: OAUTHBEARER, PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI
  ## (defaults to PLAIN)
  # sasl_mechanism = ""

  ## used if sasl_mechanism is GSSAPI
  # sasl_gssapi_service_name = ""
  # ## One of: KRB5_USER_AUTH and KRB5_KEYTAB_AUTH
  # sasl_gssapi_auth_type = "KRB5_USER_AUTH"
  # sasl_gssapi_kerberos_config_path = "/"
  # sasl_gssapi_realm = "realm"
  # sasl_gssapi_key_tab_path = ""
  # sasl_gssapi_disable_pafxfast = false

  ## used if sasl_mechanism is OAUTHBEARER
  # sasl_access_token = ""

  ## SASL protocol version.  When connecting to Azure EventHub set to 0.
  # sasl_version = 1

  # Disable Kafka metadata full fetch
  # metadata_full = false

  ## Name of the consumer group.
  # consumer_group = "telegraf_metrics_consumers"

  ## Compression codec represents the various compression codecs recognized by
  ## Kafka in messages.
  ##  0 : None
  ##  1 : Gzip
  ##  2 : Snappy
  ##  3 : LZ4
  ##  4 : ZSTD
  # compression_codec = 0
  ## Initial offset position; one of "oldest" or "newest".
  # offset = "oldest"

  ## Consumer group partition assignment strategy; one of "range", "roundrobin" or "sticky".
  # balance_strategy = "range"

  ## Maximum number of retries for metadata operations including
  ## connecting. Sets Sarama library's Metadata.Retry.Max config value. If 0 or
  ## unset, use the Sarama default of 3,
  # metadata_retry_max = 0

  ## Type of retry backoff. Valid options: "constant", "exponential"
  # metadata_retry_type = "constant"

  ## Amount of time to wait before retrying. When metadata_retry_type is
  ## "constant", each retry is delayed this amount. When "exponential", the
  ## first retry is delayed this amount, and subsequent delays are doubled. If 0
  ## or unset, use the Sarama default of 250 ms
  # metadata_retry_backoff = 0

  ## Maximum amount of time to wait before retrying when metadata_retry_type is
  ## "exponential". Ignored for other retry types. If 0, there is no backoff
  ## limit.
  # metadata_retry_max_duration = 0

  ## When set to true, this turns each bootstrap broker address into a set of
  ## IPs, then does a reverse lookup on each one to get its canonical hostname.
  ## This list of hostnames then replaces the original address list.
  ## resolve_canonical_bootstrap_servers_only = false

  ## Maximum length of a message to consume, in bytes (default 0/unlimited);
  ## larger messages are dropped
  max_message_len = 1000000

  ## Max undelivered messages
  ## This plugin uses tracking metrics, which ensure messages are read to
  ## outputs before acknowledging them to the original broker to ensure data
  ## is not lost. This option sets the maximum messages to read from the
  ## broker that have not been written by an output.
  ##
  ## This value needs to be picked with awareness of the agent's
  ## metric_batch_size value as well. Setting max undelivered messages too high
  ## can result in a constant stream of data batches to the output. While
  ## setting it too low may never flush the broker's messages.
  # max_undelivered_messages = 1000

  ## Maximum amount of time the consumer should take to process messages. If
  ## the debug log prints messages from sarama about 'abandoning subscription
  ## to [topic] because consuming was taking too long', increase this value to
  ## longer than the time taken by the output plugin(s).
  ##
  ## Note that the effective timeout could be between 'max_processing_time' and
  ## '2 * max_processing_time'.
  # max_processing_time = "100ms"

  ## The default number of message bytes to fetch from the broker in each
  ## request (default 1MB). This should be larger than the majority of
  ## your messages, or else the consumer will spend a lot of time
  ## negotiating sizes and not actually consuming. Similar to the JVM's
  ## `fetch.message.max.bytes`.
  # consumer_fetch_default = "1MB"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## Metrics

The plugin accepts arbitrary input and parses it according to the `data_format`
setting. There is no predefined metric format.

## Example Output

There is no predefined metric format, so output depends on plugin input.
