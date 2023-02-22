# Websocket Output Plugin

This plugin can write to a WebSocket endpoint.

It can output data in any of the [supported output formats][formats].

[formats]: ../../../docs/DATA_FORMATS_OUTPUT.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# A plugin that can transmit metrics over WebSocket.
[[outputs.websocket]]
  ## URL is the address to send metrics to. Make sure ws or wss scheme is used.
  url = "ws://127.0.0.1:3000/telegraf"

  ## Timeouts (make sure read_timeout is larger than server ping interval or set to zero).
  # connect_timeout = "30s"
  # write_timeout = "30s"
  # read_timeout = "30s"

  ## Optionally turn on using text data frames (binary by default).
  # use_text_frames = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional SOCKS5 proxy to use
  # socks5_enabled = true
  # socks5_address = "127.0.0.1:1080"
  # socks5_username = "alice"
  # socks5_password = "pass123"

  ## Optional HTTP proxy to use
  # use_system_proxy = false
  # http_proxy_url = "http://localhost:8888"

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"

  ## Additional HTTP Upgrade headers
  # [outputs.websocket.headers]
  #   Authorization = "Bearer <TOKEN>"
```
