# NSQ Input Plugin

This plugin gathers metrics from [NSQ](https://nsq.io/).

See the [NSQD API docs](https://nsq.io/components/nsqd.html) for endpoints that
the plugin can read.

## Configuration

```toml @sample.conf
# Read NSQ topic and channel statistics.
[[inputs.nsq]]
  ## An array of NSQD HTTP API endpoints
  endpoints  = ["http://localhost:4151"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```
