# ActiveMQ STOMP Output Plugin

This plugin writes metrics to an [Active MQ Broker][activemq] for [STOMP][stomp]
but also supports [Amazon MQ][amazonmq] brokers. Metrics can be written in one
of the supported [data formats][data_formats].

⭐ Telegraf v1.24.0
🏷️ messaging
💻 all

[activemq]: http://activemq.apache.org/
[stomp]: https://stomp.github.io
[amazonmq]:https://aws.amazon.com/amazon-mq
[data_formats]: /docs/DATA_FORMATS_OUTPUT.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username` and
`password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Configuration for active mq with stomp protocol to send metrics to
[[outputs.stomp]]
  host = "localhost:61613"

  ## Queue name for producer messages
  queueName = "telegraf"

  ## Username and password if required by the Active MQ server.
  # username = ""
  # password = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Data format to output.
  data_format = "json"
```
