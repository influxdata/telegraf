# Loki Output Plugin

This plugin sends logs to Loki, using tags as labels and one field as log line 

### Configuration:

```toml
# A plugin that can transmit logs to Loki
[[outputs.loki]]
  ## Connection timeout, defaults to "5s" if not set.
  timeout = "5s"

  ## The URL of Loki
  # url = "https://loki.domain.tld"

  ## Basic auth credential
  # username = "loki"
  # password = "pass"

  ## Additional HTTP headers
  # http_headers = {"X-Scope-OrgID" = "1"}

  ## The field containing the log
  # field_line = "log"

  ## If the request must be gzip encoded
  # gzip_request = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"  
```
