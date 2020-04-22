# Multiplay - Google Cloud PubSub Output Plugin

The GCP PubSub plugin publishes metrics to a [Google Cloud PubSub][pubsub] topic
as one of the supported [output data formats][].

It is based upon the standard Telegraf [cloud_pubsub][] plugin but loads the
credentials for Google Pub/Sub from an encrypted string in the config file.


### Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage mp_cloud_pubsub`.

```toml
[[outputs.mp_cloud_pubsub]]
  ## Required. Name of Google Cloud Platform (GCP) Project that owns
  ## the given PubSub topic.
  project = "my-project"

  ## Required. Name of PubSub topic to publish metrics to.
  topic = "my-topic"

  ## Required. Data format to consume.
  ## Each data format has its own unique set of configuration options.
  ## Read more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Required. Credentials to authenticate with Google Cloud Pub/Sub.
  ## The credentials should be an encrypted string containing the full contents
  ## of a Google Cloud service account file that has access to push messages
  ## to the configured Pub/Sub topic.
  credentials = ""

  ## Optional. If true, will send all metrics per write in one PubSub message.
  # send_batched = true

  ## The following publish_* parameters specifically configures batching
  ## requests made to the GCP Cloud PubSub API via the PubSub Golang library. Read
  ## more here: https://godoc.org/cloud.google.com/go/pubsub#PublishSettings

  ## Optional. Send a request to PubSub (i.e. actually publish a batch)
  ## when it has this many PubSub messages. If send_batched is true,
  ## this is ignored and treated as if it were 1.
  # publish_count_threshold = 1000

  ## Optional. Send a request to PubSub (i.e. actually publish a batch)
  ## when it has this many PubSub messages. If send_batched is true,
  ## this is ignored and treated as if it were 1
  # publish_byte_threshold = 1000000

  ## Optional. Specifically configures requests made to the PubSub API.
  # publish_num_go_routines = 2

  ## Optional. Specifies a timeout for requests to the PubSub API.
  # publish_timeout = "30s"
  
  ## Optional. If true, published PubSub message data will be base64-encoded.
  # base64_data = false
  
  ## Optional. PubSub attributes to add to metrics.
  # [[inputs.pubsub.attributes]]
  #   my_attr = "tag_value"
```

[pubsub]: https://cloud.google.com/pubsub
[output data formats]: /docs/DATA_FORMATS_OUTPUT.md
[cloud_pubsub]: /plugins/outputs/cloud_pubsub
