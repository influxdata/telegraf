# NATS Streaming Output Plugin

This plugin writes to a (list of) specified NATS Streaming instance(s).

```toml
[[outputs.stan]]
  ## URLs of NATS Streaming servers
  servers = ["nats://localhost:4222"]

  ## Optional credentials
  # username = ""
  # password = ""

  ## Optional NATS 2.0 and NATS NGS compatible user credentials
  # credentials = "/etc/telegraf/nats.creds"

  ## NATS Streaming subject for producer messages
  subject = "telegraf"

  ## NATS Streaming cluster id
  cluster_id = "test-cluster"

  ## NATS Streaming client id
  client_id = "telegraf"

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
