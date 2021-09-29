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
```

Server endpoint may be specified without UDP or TCP scheme (eg. "127.0.0.1:12201").
In such case, UDP protocol is assumed.
