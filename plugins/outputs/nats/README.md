# NATS Output Plugin

This plugin writes metrics to subjects of a set of [NATS][nats] instances in
one of the supported [data formats][data_formats].

⭐ Telegraf v1.1.0
🏷️ messaging
💻 all

[nats]: https://nats.io
[data_formats]: /docs/DATA_FORMATS_OUTPUT.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username` and
`password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Send telegraf measurements to NATS
[[outputs.nats]]
  ## URLs of NATS servers
  servers = ["nats://localhost:4222"]

  ## Optional client name
  # name = ""

  ## Optional credentials
  # username = ""
  # password = ""

  ## Optional NATS 2.0 and NATS NGS compatible user credentials
  # credentials = "/etc/telegraf/nats.creds"

  ## NATS subject for producer messages.
  ##
  ## This field can be a static subject string (e.g., "telegraf"), or a dynamic subject defined
  ## using Go template syntax. Templates allow you to construct the subject based on metric tags,
  ## name, and field, providing fine-grained routing.
  ##
  ## Example using a dynamic subject:
  ## subject = "{{ .GetTag \"region\" }}.{{ .GetTag \"datacenter\" }}.{{ .GetTag \"host\" }}.{{ .Name }}.{{ .Field }}"
  ##
  ## Including `.Field` in the template will emit one message per field, which can substantially
  ## increase message volume. Use this only when field-level granularity is required.
  ##
  ## For JetStream:
  ## - This value determines the subject where messages will be published.
  ## - **If a dynamic template is used**, this subject is **not** automatically added to the JetStream
  ##   stream’s subject list. You must explicitly define matching subjects under
  ##   `outputs.nats.jetstream.subjects` to ensure proper stream creation or update.
  subject = "telegraf"

  ## Use Transport Layer Security
  # secure = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## Jetstream specific configuration. If not nil, it will assume Jetstream context.
  ## Since this is a table, it should be present at the end of the plugin section. Else you can use inline table format.
  # [outputs.nats.jetstream]
    ## Name of the stream, required when using jetstream. Telegraf will
    ## use the union of the above subject and below the subjects array.
    # name = ""
    # subjects = []

    ## Use asynchronous publishing for higher throughput, but note that it does not guarantee order within batches.
    # async_publish = false

    ## Timeout for wating on acknowledgement on asynchronous publishing
    ## String with valid units "ns", "us" (or "µs"), "ms", "s", "m", "h".
    # async_ack_timeout = "5s"

    ## Full jetstream create stream config, refer: https://docs.nats.io/nats-concepts/jetstream/streams
    # retention = "limits"
    # max_consumers = -1
    # max_msgs_per_subject = -1
    # max_msgs = -1
    # max_bytes = -1
    # max_age = 0
    # max_msg_size = -1
    # storage = "file"
    # discard = "old"
    # num_replicas = 1
    # duplicate_window = 120000000000
    # sealed = false
    # deny_delete = false
    # deny_purge = false
    # allow_rollup_hdrs = false
    # allow_direct = true
    # mirror_direct = false

    ## Disable creating the stream but assume the stream is managed externally
    ## and already exists. This will make the plugin fail if the steam does not exist.
    # disable_stream_creation = false
```
