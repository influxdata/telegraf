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

  ## Optional authentication with nkey seed file (NATS 2.0)
  # nkey_seed = "/etc/telegraf/seed.txt"

  ## NATS subject for producer messages.
  ## This field can be a static string or a Go template, see README for details.
  ## Incompatible with `use_batch_format
  subject = "telegraf"

  ## Use Transport Layer Security
  # secure = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Use batch serialization instead of per metric. The batch format allows for the
  ## production of batch output formats and may more efficiently encode and write metrics.
  # use_batch_format = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## Jetstream specific configuration. If not nil, it will assume Jetstream context.
  ## Since this is a table, it should be present at the end of the plugin section. Else you can use inline table format.
  # [outputs.nats.jetstream]
    ## Name of the stream, required when using jetstream.
    # name = ""
    ## List of subjects to register on the stream
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

### Subject Configuration

The `subject` setting determines where producer messages will be published
in NATS. This can be a static subject (e.g., "telegraf"), or a dynamic
subject template using Go’s text/template syntax.

Dynamic templates allow you to construct subjects based on properties of
each metric, such as tags, name and fields. This enables fine-grained
routing and filtering across NATS or JetStream subscribers.

This feature is incompatible with `use_batch_format`

#### Examples

Routing based on tags and metric name:

```toml
subject = '{{ .Tag "region" }}.{{ .Tag "datacenter" }}.{{ .Tag "host" }}.{{ .Name }}'
```

Routing based on tags, metric name and field name:

```toml
subject = 'telegraf.metrics.{{ .Tag "datacenter" }}.{{ .Tag "host" }}.{{ .Name }}.{{ .Field "Value1" }}'
```

If you’re using JetStream the value of subject determines where messages
are published.

> [!IMPORTANT]
> When using a dynamic subject template, Telegraf does not automatically
> register the generated subjects with the JetStream stream.

For dynamic `subject`s you must explicitly define matching subjects in
`outputs.nats.jetstream.subjects` to ensure your stream can receive and
retain those messages correctly.
