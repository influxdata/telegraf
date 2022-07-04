# STOMP Producer Output Plugin

This plugin writes to a [Active MQ Broker](http://activemq.apache.org/)
for STOMP <http://stomp.github.io>.

It also support Amazon MQ  <https://aws.amazon.com/amazon-mq/>

## Configuration

```toml @sample.conf
  ## Host of Active MQ broker
  host = "localhost:61613"

  ## Queue name for producer messages
  # queueName = "telegraf"

  ## Username and password if required by the Active MQ server.
  # username = ""
  # password = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Data format to output.
  # data_format = "json"
```
