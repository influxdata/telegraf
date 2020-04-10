# Graylog Output Plugin

This plugin writes to a Graylog instance using the "[GELF][]" format.

[GELF]: https://docs.graylog.org/en/3.1/pages/gelf.html#gelf-payload-specification

### Configuration:

```toml
[[outputs.graylog]]
  ## UDP endpoint for your graylog instances.
  servers = ["127.0.0.1:12201"]

  ## The field to use as the GELF short_message, if unset the static string
  ## "telegraf" will be used.
  ##   example: short_message_field = "message"
  # short_message_field = ""
```
