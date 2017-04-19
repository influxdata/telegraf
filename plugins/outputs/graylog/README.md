# Graylog Output Plugin

This plugin writes to a Graylog instance using the "gelf" format.

It requires a `servers` name.

### Configuration:

```toml
# Send telegraf metrics to graylog(s)
[[outputs.graylog]]
  ## UDP endpoint for your graylog instance(s).
  servers = ["127.0.0.1:12201", "192.168.1.1:12201"]
```
