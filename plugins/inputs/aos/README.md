# Apstra AOS Input Plugin
The [AOS](https://www.apstra.com/products/aos-overview/) input plugin parses
protobuf messages sent by the AOS server and formats them as time series 
database metrics, e.g., Influxdb. These messages contain telemetry data
that AOS generates and consist of performance monitoring (perfmon), alerts and events.

The plugin configures the telegraf instance as a protobuf streaming endpoint to
receive either perfmon, alert or event messages, or any combination thereof.
The message data is also augmented with information collected from the AOS server
via the REST API.

## Configuration
```
[[inputs.aos]]
  # TCP Port to listen for incoming sessions from the AOS Server.
  port = 7777

  # Address of the server running Telegraf, it needs to be reachable from AOS.
  address = "192.168.59.1"

  # Interval to refresh content from the AOS server (in sec).
  # refresh_interval = 30

  # Streaming Type Can be "perfmon", "alerts" or "events".
  streaming_type = [ "perfmon", "alerts" ]

  # Define parameters to join the AOS Server using the REST API.
  aos_server = "192.168.59.250"
  aos_port = 443
  aos_login = "admin"
  aos_password = "admin"
  aos_protocol = "https"

```

## Support 
This plugin supports AOS up to version 3.1.0.