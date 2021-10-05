# Graylog Output Plugin

This plugin writes to a Graylog instance using the "[GELF][]" format.

[GELF]: https://docs.graylog.org/en/3.1/pages/gelf.html#gelf-payload-specification

### Configuration:

```toml
[[outputs.graylog]]
  ## Endpoints for your graylog instances.
  servers = ["udp://127.0.0.1:12201"]

  ## Connection timeout.
  # timeout = "5s"

  ## The field to use as the GELF short_message, if unset the static string
  ## "telegraf" will be used.
  ##   example: short_message_field = "message"
  # short_message_field = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

Server endpoint may be specified without UDP or TCP scheme (eg. "127.0.0.1:12201").
In such case, UDP protocol is assumed. TLS config is ignored for UDP endpoints.
