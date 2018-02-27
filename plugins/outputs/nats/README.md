# NATS Output Plugin

This plugin writes to a (list of) specified NATS instance(s).

```
[[outputs.nats]]
  ## URLs of NATS servers
  servers = ["nats://localhost:4222"]
  ## Optional credentials
  # username = ""
  # password = ""
  ## NATS subject for producer messages
  subject = "telegraf"
  ## Optional TLS Config
  ## CA certificate used to self-sign NATS server(s) TLS certificate(s)
  # tls_ca = "/etc/telegraf/ca.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

### Required parameters:

* `servers`:  List of strings, this is for NATS clustering support. Each URL should start with `nats://`.
* `subject`: The NATS subject to publish to.

### Optional parameters:

* `username`: Username for NATS
* `password`: Password for NATS
* `tls_ca`: TLS CA
* `insecure_skip_verify`: Use SSL but skip chain & host verification (default: false)
