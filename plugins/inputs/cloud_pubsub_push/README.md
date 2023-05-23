# Google Cloud PubSub Push Input Plugin

The Google Cloud PubSub Push listener is a service input plugin that listens
for messages sent via an HTTP POST from [Google Cloud PubSub][pubsub].
The plugin expects messages in Google's Pub/Sub JSON Format ONLY. The intent
of the plugin is to allow Telegraf to serve as an endpoint of the
Google Pub/Sub 'Push' service.  Google's PubSub service will **only** send
over HTTPS/TLS so this plugin must be behind a valid proxy or must be
configured to use TLS.

Enable TLS by specifying the file names of a service TLS certificate and key.

Enable mutually authenticated TLS and authorize client connections by signing
certificate authority by including a list of allowed CA certificate file names
in `tls_allowed_cacerts`.

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
# Google Cloud Pub/Sub Push HTTP listener
[[inputs.cloud_pubsub_push]]
  ## Address and port to host HTTP listener on
  service_address = ":8080"

  ## Application secret to verify messages originate from Cloud Pub/Sub
  # token = ""

  ## Path to listen to.
  # path = "/"

  ## Maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## Maximum duration before timing out write of the response. This should be
  ## set to a value large enough that you can send at least 'metric_batch_size'
  ## number of messages within the duration.
  # write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 524,288,00 bytes (500 mebibytes)
  # max_body_size = "500MB"

  ## Whether to add the pubsub metadata, such as message attributes and
  ## subscription as a tag.
  # add_meta = false

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

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

This plugin assumes you have already created a PUSH subscription for a given
PubSub topic.

[pubsub]: https://cloud.google.com/pubsub

## Metrics

## Example Output
