# MavLink Input Plugin

The `mavlink` plugin connects to a MavLink-compatible flight controller such as as [ArduPilot](https://ardupilot.org/) or [PX4](https://px4.io/). and translates all incoming messages into metrics.

The purpose of this plugin is to allow Telegraf to be used to ingest live flight metrics from unmanned systems (drones, planes, boats, etc.)  

Warning: This input plugin potentially generates a large amount of data! Use the configuration to limit the set of messages or the rate, or use another telegraf plugin to filter the output.

## Configuration

```toml @sample.conf
# Read metrics from a Mavlink connection to a flight controller.
[[inputs.mavlink]]
     # Flight controller URL. Must be a valid Mavlink connection string in one
     # of the following formats:
     #
     # - Serial port:  serial://<device name>:<baud rate> 
     #            eg: "serial:///dev/ttyACM0:57600"
     # 
     # - TCP client:   tcp://<target ip or hostname>:<port>
     #            eg: "tcp://192.168.1.12:14550"
     # 
     # - UDP client:   udp://<target ip or hostname>:<port>
     #            eg: "udp://192.168.1.12:14550"
     # 
     # - UDP server:   udp://:<listen port>
     #            eg: "udp://:14540"
     # 
     # The meaning of each of these modes is documented by
     # https://mavsdk.mavlink.io/v1.4/en/cpp/guide/connections.html.
     fcu_url = "udp://:14540"

     # Filter to specific messages. Only the messages in this list will be parsed.
     # If blank, all messages will be accepted.
     message_filter = []

     # Mavlink system ID for Telegraf
     system_id = 254
```

### Note: Mavlink Dialects

This plugin currently only supports the [ardupilotmega](https://mavlink.io/en/messages/ardupilotmega.html) ArduPilot-specific dialect, which also includes messages from the common Mavlink dialect.

## Metrics

Each supported Mavlink message translates to one metric group, and fields on the Mavlink message are converted to fields in telegraf.

The name of the Mavlink message is translated into lowercase and any leading text `message_` is dropped.

For example, [MESSAGE_ATTITUDE](https://mavlink.io/en/messages/common.html#ATTITUDE) will become an `attitude` metric, with all fields copied from its Mavlink message definition.

## Example Output

_`mavlink` input plugin connected to ArduPilot SITL and the `file` output plugin:_
```


```