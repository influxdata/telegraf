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

  ## Jetstream specific configuration. If specified, telegraf will use Jetstream to publish messages
  # [outputs.nats.jetstream]
    ## Specifies whether telegraf should create the stream at the startup or not. It will only create if it doesn't exist.
    # auto_create_stream = true

    ## Name of the stream where nats jetstream will publish the messages
    # stream = "my-jetstream"

    ## When the `auto_create_stream` option is set to true in the JetStream configuration, 
    ## telegraf dynamically creates the JetStream stream config using the JSON provided. 
    ## In this scenario, the `name` and `subjects` fields from the JSON configuration will be ignored, and the values will be determined as follows:
    ## The stream name (`name`) will be taken from the `stream` field in the `jetstream` section of the Telegraf configuration.
    ## The subjects (`Subjects`) for the stream will be derived from the `subject` field in the `nats` section of the Telegraf configuration.
    # stream_config_json = '''
    # {
    #     "retention": "workqueue",
    #     "max_consumers": 10,
    #     "discard": "old",
    #     "storage": "file",
    #     "max_msgs": 100000,
    #     "max_bytes": 104857600,  // 100 MB
    #     "max_age": 86400000000000, // in the int64 format
    #     "num_replicas": 1
    # }
    # '''

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
```
