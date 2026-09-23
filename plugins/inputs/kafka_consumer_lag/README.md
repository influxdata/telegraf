# Kafka Consumer Lag Input Plugin

This plugin collects the lag of [Kafka][kafka] consumer groups directly from
the brokers using the Kafka protocol, without requiring an external component
such as Burrow or a Prometheus exporter.

The lag of a partition is the difference between the partition's log end
offset (fetched via `ListOffsets`) and the offset committed by the consumer
group (fetched via `OffsetFetch`). Metrics are emitted per partition and
aggregated per topic and per group, each level can be disabled.

> [!NOTE]
> Consumer group lag cannot be obtained from broker JMX metrics as brokers do
> not expose committed offsets there. This plugin fetches the committed
> offsets the same way `kafka-consumer-groups.sh --describe` does.

⭐ Telegraf v1.41.0
🏷️ messaging
💻 all

[kafka]: https://kafka.apache.org

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

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
- `probe`:  Telegraf will probe the plugin's function (if possible) and disables
            the plugin in case probing fails. If the plugin does not support
            probing, Telegraf will behave as if `ignore` was set instead.

## Secret store support

This plugin supports secrets from secret stores for the `sasl_username`,
`sasl_password` and `sasl_access_token` option.
See the [secret store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Collect consumer group lag directly from Kafka brokers
[[inputs.kafka_consumer_lag]]
  ## Kafka brokers.
  brokers = ["localhost:9092"]

  ## Set the minimal supported Kafka version. Setting this enables the use of
  ## new Kafka features and APIs. Must be 0.11.0.0 or greater. Please check
  ## the list of supported versions at
  ## https://pkg.go.dev/github.com/IBM/sarama#SupportedVersions
  ##   ex: version = "3.6.0"
  # version = ""

  ## Consumer groups to monitor, specified as glob patterns.
  ## By default all groups are monitored.
  # groups_include = ["*"]
  # groups_exclude = []

  ## Topics to monitor, specified as glob patterns.
  ## By default all topics with committed offsets are monitored.
  # topics_include = ["*"]
  # topics_exclude = []

  ## Levels at which lag metrics are emitted. Valid entries are "partition",
  ## "topic" and "group". The partition level emits one metric per
  ## (group, topic, partition), the topic and group levels emit the sum and
  ## the maximum over the contained partitions.
  # metric_levels = ["partition", "topic", "group"]

  ## Only collect groups whose coordinator is the broker with this ID. This
  ## is intended for deployments running Telegraf on every broker host, so
  ## that each group is collected exactly once and collection follows
  ## coordinator moves automatically. Groups are then listed from this broker
  ## only, so "brokers" may contain just the local broker.
  ## Set to -1 to collect all groups.
  # coordinator_broker_id = -1

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

  ## SASL authentication credentials. These settings should typically be used
  ## with TLS encryption enabled
  # sasl_username = ""
  # sasl_password = ""

  ## Optional SASL, one of:
  ##   OAUTHBEARER, PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI, AWS-MSK-IAM
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

  ## used if sasl_mechanism is AWS-MSK-IAM
  # sasl_aws_msk_iam_region = ""
  ## for profile based auth
  ## sasl_aws_msk_iam_profile = ""
  ## for role based auth
  ## sasl_aws_msk_iam_role = ""
  ## sasl_aws_msk_iam_session = ""

  ## Arbitrary key value string pairs to pass as a TOML table. For example:
  ## {logicalCluster = "cluster-042", poolId = "pool-027"}
  # sasl_extensions = {}

  ## SASL protocol version. When connecting to Azure EventHub set to 0.
  # sasl_version = 1

  ## Fetch metadata for all topics instead of only the ones in use.
  ## Set to false on clusters with a very large number of topics.
  # metadata_full = true

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
```

### Required permissions

When the cluster uses ACLs, the configured principal needs `Describe`
permission on the monitored consumer groups and on the topics they consume.
No `Read` permission is required as the plugin never joins a group or
consumes messages.

Missing `Describe` permission on a topic is not reported as an error. The
brokers silently omit the committed offsets of such topics, so their
partitions are missing from the metrics.

### Running on every broker host

When Telegraf runs on every broker host, each instance would by default
collect the lag of every group in the cluster, resulting in duplicate
metrics. Set `coordinator_broker_id` to the ID of the local broker so that
each instance only collects the groups coordinated by its broker. Kafka
distributes groups across coordinators and moves them when a broker fails,
so the collection is sharded and fails over automatically without any
coordination between the Telegraf instances.

With `coordinator_broker_id` set, the group list, the group descriptions and
the committed offsets are fetched from the local broker, so `brokers` can
simply point to it, e.g. `["localhost:9092"]`. Log end offsets are still
fetched from the leader of each partition, which may be another broker.
Metadata and coordinator lookups may be sent to any available broker. Other
brokers being unavailable therefore only affects the partitions they lead,
not the collection as a whole.

### Unavailable brokers

Without `coordinator_broker_id`, the consumer groups are listed from every
broker as each broker only knows the groups it coordinates. A broker failing
to answer is reported as an error and only its groups are missing from that
collection, the groups of the remaining brokers are still collected.

### Number of requests

Committed offsets are fetched with a single `OffsetFetch` request per
coordinator covering all of its groups, so the number of requests per
collection does not grow with the number of consumer groups. This requires
brokers running Kafka 3.0.0 or later and `version` set accordingly. On older
brokers the plugin transparently falls back to one request per group. Log end
offsets are always fetched with one `ListOffsets` request per partition leader.

### Cardinality

The partition level emits one series per group, topic and partition, so the
number of series grows with the product of the three. On large clusters with
many consumer groups this can reach hundreds of thousands of series. In this
case drop the partition level from `metric_levels` and rely on the `lag_max`
field of the topic and group levels to detect skewed partitions, or enable the
partition level for selected groups only using a second plugin instance with
`groups_include`. Note that dropping the partition level only reduces the
number of emitted series, the same offsets are still fetched from Kafka.

To reduce the data fetched from Kafka, use the filters instead.
`groups_include` and `groups_exclude` limit the groups whose committed offsets
are fetched. `topics_include` and `topics_exclude` limit the partitions whose
log end offsets are fetched. The topic filter is applied to the offsets
returned by the brokers, as the committed offsets of a group are always
fetched for all topics it committed on. Listing the groups is not affected by
the group filter either, as the group names are only known afterwards.

Short-lived groups such as the ones created by `kafka-console-consumer` or by
batch jobs using a fresh group ID on every run create new series each time
which linger until their offsets expire. Exclude such groups, for example with
`groups_exclude = ["console-consumer-*"]`, to avoid series churn.

### Partitions without committed offsets

A group only has a lag for partitions it committed an offset for. Partitions
which have never been committed are skipped as there is no offset to compare
against the log end offset. Partitions of topics which were deleted but still
have offsets stored for the group are skipped as well. The plugin never
creates topics, even if the brokers allow automatic topic creation.

## Metrics

- `kafka_consumer_group_partition` (one metric per group, topic and partition)
  - tags:
    - `group` (name of the consumer group)
    - `topic` (name of the topic)
    - `partition` (partition number)
  - fields:
    - `committed_offset` (int64, offset committed by the group)
    - `log_end_offset` (int64, high watermark of the partition, i.e. the
      offset following the last message available to consumers)
    - `lag` (int64, `log_end_offset - committed_offset`, zero if negative)

- `kafka_consumer_group_topic` (one metric per group and topic)
  - tags:
    - `group` (name of the consumer group)
    - `topic` (name of the topic)
  - fields:
    - `lag_sum` (int64, sum of the lag over all partitions)
    - `lag_max` (int64, maximum lag of a single partition)
    - `partitions` (int64, number of partitions with a computed lag)

- `kafka_consumer_group` (one metric per group)
  - tags:
    - `group` (name of the consumer group)
  - fields:
    - `lag_sum` (int64, sum of the lag over all partitions)
    - `lag_max` (int64, maximum lag of a single partition)
    - `partitions` (int64, number of partitions with a computed lag)
    - `topics` (int64, number of topics with a computed lag)
    - `members` (int64, number of members currently joined to the group)

The aggregates only cover partitions whose lag could be computed. Partitions
whose log end offset could not be fetched are reported as errors and are
missing from `lag_sum`, `lag_max`, `partitions` and `topics`.

The `log_end_offset` is the same value `kafka-consumer-groups.sh` reports as
`LOG-END-OFFSET`. The committed offset can exceed it, e.g. if a topic was
recreated or its log was truncated after the offsets were committed, in which
case the `lag` is reported as zero.

The `members` field is omitted if the group could not be described. This is
always the case for groups using the consumer protocol introduced with
[KIP-848][kip848] (`group.protocol=consumer`), as they cannot be described
with the classic `DescribeGroups` API. Groups which only commit offsets
without joining, e.g. Flink jobs or offsets synced by MirrorMaker 2, always
have zero members.

[kip848]: https://cwiki.apache.org/confluence/display/KAFKA/KIP-848%3A+The+Next+Generation+of+the+Consumer+Rebalance+Protocol

## Example Output

```text
kafka_consumer_group_partition,group=order-svc,partition=0,topic=orders committed_offset=999950i,lag=50i,log_end_offset=1000000i 1726560000000000000
kafka_consumer_group_partition,group=order-svc,partition=1,topic=orders committed_offset=980000i,lag=20000i,log_end_offset=1000000i 1726560000000000000
kafka_consumer_group_topic,group=order-svc,topic=orders lag_max=20000i,lag_sum=20050i,partitions=2i 1726560000000000000
kafka_consumer_group,group=order-svc lag_max=20000i,lag_sum=20050i,members=2i,partitions=2i,topics=1i 1726560000000000000
```
