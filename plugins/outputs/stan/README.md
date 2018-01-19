# NATS Streaming Output Plugin

This plugin writes to a (list of) specified NATS Streaming instance(s).

```
[[outputs.stan]]
  ## NATS Streaming Cluster ID
  cluster_id = "test-cluster"
  ## Client ID
  client_id = "telegraf-client-id"
  ## URLs of NATS servers
  servers = ["nats://localhost:4222"]
  ## Optional credentials
  # username = ""
  # password = ""
  ## NATS subject for producer messages
  subject = "telegraf"
  ## Optional NATS Streaming discover prefix
  # discover_prefix = "_STAN.discover"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

### Required parameters:

* `cluster_id`: The NATS Streaming cluster ID.
* `client_id`: A unique identifier of this client sent to the NATS server.
* `servers`:  List of strings, this is for NATS clustering support. Each URL should start with `nats://`.
* `subject`: The NATS subject to publish to.

### Optional parameters:

* `username`: Username for NATS
* `password`: Password for NATS
* `tls_ca`: TLS CA
* `insecure_skip_verify`: Use SSL but skip chain & host verification (default: false)
* `discover_prefix`: The prefix used to discover NATS Streaming servers.
