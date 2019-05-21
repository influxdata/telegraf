# Logz.io Output Plugin

This plugin sends metrics to [Logz.io](https://logz.io/) over HTTPs.

### Configuration:

```toml
[[outputs.logzio]]
  ## Logz.io account token
  token = "your Logz.io token" # required

  ## Use your listener URL for your Logz.io account region.
  # url = "https://listener.logz.io:8071"
  
  ## Timeout for HTTP requests
  # timeout = "5s"
  
  ## Optional TLS Config for use on HTTP connections
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```