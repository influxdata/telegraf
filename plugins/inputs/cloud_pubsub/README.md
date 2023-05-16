# Google Cloud PubSub Input Plugin

The GCP PubSub plugin ingests metrics from [Google Cloud PubSub][pubsub]
and creates metrics using one of the supported [input data formats][].

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

## Configuration

```toml @sample.conf
# Read metrics from Google PubSub
[[inputs.cloud_pubsub]]
  ## Required. Name of Google Cloud Platform (GCP) Project that owns
  ## the given PubSub subscription.
  project = "my-project"

  ## Required. Name of PubSub subscription to ingest metrics from.
  subscription = "my-subscription"

  ## Required. Data format to consume.
  ## Each data format has its own unique set of configuration options.
  ## Read more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Optional. Filepath for GCP credentials JSON file to authorize calls to
  ## PubSub APIs. If not set explicitly, Telegraf will attempt to use
  ## Application Default Credentials, which is preferred.
  # credentials_file = "path/to/my/creds.json"

  ## Optional. Number of seconds to wait before attempting to restart the
  ## PubSub subscription receiver after an unexpected error.
  ## If the streaming pull for a PubSub Subscription fails (receiver),
  ## the agent attempts to restart receiving messages after this many seconds.
  # retry_delay_seconds = 5

  ## Optional. Maximum byte length of a message to consume.
  ## Larger messages are dropped with an error. If less than 0 or unspecified,
  ## treated as no limit.
  # max_message_len = 1000000

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

  ## The following are optional Subscription ReceiveSettings in PubSub.
  ## Read more about these values:
  ## https://godoc.org/cloud.google.com/go/pubsub#ReceiveSettings

  ## Optional. Maximum number of seconds for which a PubSub subscription
  ## should auto-extend the PubSub ACK deadline for each message. If less than
  ## 0, auto-extension is disabled.
  # max_extension = 0

  ## Optional. Maximum number of unprocessed messages in PubSub
  ## (unacknowledged but not yet expired in PubSub).
  ## A value of 0 is treated as the default PubSub value.
  ## Negative values will be treated as unlimited.
  # max_outstanding_messages = 0

  ## Optional. Maximum size in bytes of unprocessed messages in PubSub
  ## (unacknowledged but not yet expired in PubSub).
  ## A value of 0 is treated as the default PubSub value.
  ## Negative values will be treated as unlimited.
  # max_outstanding_bytes = 0

  ## Optional. Max number of goroutines a PubSub Subscription receiver can spawn
  ## to pull messages from PubSub concurrently. This limit applies to each
  ## subscription separately and is treated as the PubSub default if less than
  ## 1. Note this setting does not limit the number of messages that can be
  ## processed concurrently (use "max_outstanding_messages" instead).
  # max_receiver_go_routines = 0

  ## Optional. If true, Telegraf will attempt to base64 decode the
  ## PubSub message data before parsing. Many GCP services that
  ## output JSON to Google PubSub base64-encode the JSON payload.
  # base64_data = false

  ## Content encoding for message payloads, can be set to "gzip" or
  ## "identity" to apply no encoding.
  # content_encoding = "identity"

  ## If content encoding is not "identity", sets the maximum allowed size, 
  ## in bytes, for a message payload when it's decompressed. Can be increased 
  ## for larger payloads or reduced to protect against decompression bombs.
  ## Acceptable units are B, KiB, KB, MiB, MB...
  # max_decompression_size = "500MB"
```

### Multiple Subscriptions and Topics

This plugin assumes you have already created a PULL subscription for a given
PubSub topic. To learn how to do so, see [how to create a subscription][pubsub
create sub].

Each plugin agent can listen to one subscription at a time, so you will
need to run multiple instances of the plugin to pull messages from multiple
subscriptions/topics.

[pubsub]: https://cloud.google.com/pubsub
[pubsub create sub]: https://cloud.google.com/pubsub/docs/admin#create_a_pull_subscription
[input data formats]: /docs/DATA_FORMATS_INPUT.md

## Metrics

## Example Output
