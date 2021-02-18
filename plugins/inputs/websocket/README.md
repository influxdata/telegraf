# Websocket Input Plugin

The websocket input plugin collects metrics from a websocket endpoint and parses the received message in one of the
supported [input data formats](../../../docs/DATA_FORMATS_INPUT.md).
This plugin supports two operation modes, an event-based mode and a sampling mode. The event-based mode is enabled in
case the `trigger_body` is empty (default), waiting for the server to actively push metrics through the websocket.
The sampling mode is used if `trigger_body` is not empty. In this case the content of `trigger_body` is sent to the
server in the regular telegraf `interval` and the returned metric is gathered and parsed using the configured
data-format.

### Configuration:

```toml
# Read formatted metrics from a websocket endpoint
[[inputs.websocket]]
  ## URL to read the metrics from (mandatory)
  url = "ws://localhost:8080"

  ## Messages to send to the websocket in order to initialize the connection.
  ## If an empty message is found, the sending is paused for "handshake_pause"
  ## long before sending the next message.
  ## If set to empty (default), nothing will be sent.
  # handshake_bodies = []
  # handshake_pause = "100ms"

  ## Message to send to the websocket in order to trigger sending of a metric
  ## If set to empty (default), this plugin will wait for the server to send
  ## messages in an event-based fashion. Otherwise, the content of this option
  ## will be sent in each gather interval actively triggering a metric.
  # trigger_body = ""

  ## Amount of time allowed to complete a request
  # timeout = "5s"

  ## HTTP Proxy support
  # http_proxy_url = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"
```

### Metrics:

The metrics collected by this input plugin will depend on the configured `data_format` and the payload returned by the websocket endpoint.

The default values below are added if the input format does not specify a value:

- websocket
  - tags:
    - url
