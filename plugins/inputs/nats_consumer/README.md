# NATS Consumer Input Plugin

The [NATS][nats] consumer plugin reads from the specified NATS subjects and
creates metrics using one of the supported [input data formats][].

A [Queue Group][queue group] is used when subscribing to subjects so multiple
instances of telegraf can read from a NATS cluster in parallel.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Read metrics from NATS subject(s)
[[inputs.nats_consumer]]
  ## urls of NATS servers
  servers = ["nats://localhost:4222"]

  ## subject(s) to consume
  ## If you use jetstream you need to set the subjects
  ## in jetstream_subjects
  subjects = ["telegraf"]

  ## jetstream subjects
  ## jetstream is a streaming technology inside of nats.
  ## With jetstream the nats-server persists messages and 
  ## a consumer can consume historical messages. This is
  ## useful when telegraf needs to restart it don't miss a 
  ## message. You need to configure the nats-server.
  ## https://docs.nats.io/nats-concepts/jetstream.
  jetstream_subjects = ["js_telegraf"]

  ## name a queue group
  queue_group = "telegraf_consumers"

  ## Optional credentials
  # username = ""
  # password = ""

  ## Optional NATS 2.0 and NATS NGS compatible user credentials
  # credentials = "/etc/telegraf/nats.creds"

  ## Use Transport Layer Security
  # secure = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Sets the limits for pending msgs and bytes for each subscription
  ## These shouldn't need to be adjusted except in very high throughput scenarios
  # pending_message_limit = 65536
  # pending_bytes_limit = 67108864

  ## Maximum messages to read from the broker that have not been written by an
  ## output.  For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## subject is added as tag subject or the value in subject_tag, if subject_tag is empty string this tag is ignored
  # subject_tag = ""

  ## Enable extracting tag values from NATS subjects
  ## _ denotes an ignored entry in the topic path
  # [[inputs.nats_consumer.subject_parsing]]
  #   subject = ""
  #   measurement = ""
  #   tags = ""
  #   fields = ""
  ## Value supported is int, float, unit
  #   [[inputs.nats_consumer.subject_parsing]]
  #      key = type
```

## About Subject Parsing

The NATS subject as a whole is stored as a tag, but this can be far too coarse
to be easily used when utilizing the data further down the line. This
change allows tag values to be extracted from the NATS subject letting you
store the information provided in the subject in a meaningful way.
An `_` denotes an ignored entry in the subject path.
Please see the following example.

## Example Configuration for subject parsing

```toml
[[inputs.nats_consumer]]
  ## urls of NATS servers
  servers = ["nats://localhost:4222"]

 ## subject(s) to consume
  subjects = [
    "telegraf.*.cpu.23",
  ]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "value"
  data_type = "float"

  subject_tag = "subject"

  [[inputs.nats_consumer.subject_parsing]]
    subject = "telegraf.one.cpu.23"
    measurement = "_._.measurement._"
    tags = "tag._._._"
    fields = "_._._.test"
    [inputs.nats_consumer.subject_parsing.types]
      test = "int"
```

Result:

```shell
cpu,host=pop-os,tag=telegraf,subject=telegraf.one.cpu.23 value=45,test=23i 1637014942460689291
```

[nats]: https://www.nats.io/about/
[input data formats]: /docs/DATA_FORMATS_INPUT.md
[queue group]: https://www.nats.io/documentation/concepts/nats-queueing/

## Metrics

Which data you will get depends on the subjects you consume from nats

## Example Output

Depends on the nats subject input
nats_consumer,host=[] value=1.9 1655972309339341000
