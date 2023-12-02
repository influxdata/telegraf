# NATS Output Plugin

This plugin writes to a (list of) specified NATS instance(s).

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

  ## NATS subject for producer messages
  ## For jetstream this is also the subject where messages will be published
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

  ## If the value is non-empty, enable jetstream based publishing.
  ## Name of the stream where nats jetstream will publish the messages.
  ## If the stream already exists, it will Update it using the fields specified in the jetstream section.
  ## Else it will create it.
  # jetstream_stream = "telegraf-metrics-stream"

  ## Jetstream specific configuration
  ## If this section is empty, and jetstream_stream is specified, the stream_create config would have
  ## just the two fields- Name(jetstream_stream) and Subjects([]string{subject})
  ## Since this is a table, it should be present at the end of the plugin section. Else you can use inline table format.
  # [outputs.nats.jetstream]
    ## Full jetstream create stream config, refer: https://docs.nats.io/nats-concepts/jetstream/streams
    ## The `name` and `subjects` fields from configuration will be ignored, and the values will be determined as follows:
    ## The stream name (`name`) will be taken from the `jetstream_stream` field in the `outputs.nats` section of the Telegraf configuration.
    ## The subjects (`Subjects`) for the stream will be derived from the `subject` field in the `outputs.nats` section of the Telegraf configuration.
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
```
