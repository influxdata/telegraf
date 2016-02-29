# UDP listener service input plugin

The UDP listener is a service input plugin that listens for messages on a UDP
socket and adds those messages to InfluxDB.
The plugin expects messages in the
[Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

### Configuration:

This is a sample configuration for the plugin.

```toml
[[inputs.udp_listener]]
  ## Address and port to host UDP listener on
  service_address = ":8125"

  ## Number of UDP messages allowed to queue up. Once filled, the
  ## UDP listener will start dropping packets.
  allowed_pending_messages = 10000

  ## UDP packet size for the server to listen for. This will depend
  ## on the size of the packets that the client is sending, which is
  ## usually 1500 bytes.
  udp_packet_size = 1500

  ## Data format to consume. This can be "json", "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```
