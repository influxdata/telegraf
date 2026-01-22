# MavLink Input Plugin

This plugin collects metrics from [MavLink][mavlink]-compatible flight
controllers such as [ArduPilot][ardupilot] or [PX4][px4] to live ingest
flight metrics from unmanned systems (drones, planes, boats, etc.)
Currently the ArduPilot-specific Mavlink dialect is used, check the
[Mavlink documentation][mavlink_docs] for more details and the various
messages available.

> [!WARNING]
> This plugin potentially generates a large amount of data. If your output
> plugin cannot handle the rate of messages, use
> [Metric filters][metric_filters] to limit the metrics written to outputs,
> and/or the `filters` configuration parameter to limit which Mavlink messages
> this plugin parses.

⭐ Telegraf v1.35.0
🏷️ iot
💻 all

[ardupilot]: https://ardupilot.org/
[mavlink]: https://mavlink.io/
[mavlink_docs]: https://mavlink.io/en/messages/ardupilotmega.html
[metric_filters]: ../../../docs/CONFIGURATION.md#metric-filtering
[px4]: https://px4.io/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from a Mavlink flight controller.
[[inputs.mavlink]]
  ## Flight controller URL supporting serial port, UDP and TCP connections.
  ## Options are documented at
  ##   https://mavsdk.mavlink.io/v1.4/en/cpp/guide/connections.html.
  ##
  ## Examples:
  ## - Serial port: serial:///dev/ttyACM0:57600
  ## - TCP client:  tcp://192.168.1.12:5760
  ## - UDP client:  udp://192.168.1.12:14550
  ## - TCP server:  tcpserver://:5760
  ## - UDP server:  udpserver://:14550
  # url = "tcp://127.0.0.1:5760"

  ## Filter to specific messages. Only the messages in this list will be parsed.
  ## If blank or unset, all messages will be accepted. Glob syntax is accepted.
  ## Each message in this list should be lowercase camel_case, with "message_"
  ## prefix removed, eg: "global_position_int", "attitude"
  # filter = []

  ## Mavlink system ID for Telegraf. Only used if the mavlink plugin is sending
  ## messages, eg. when `stream_request_frequency` is 0 (see below.)
  # system_id = 254

  ## Determines whether the plugin sends requests to subscribe to data.
  ## In mavlink, stream rates must be configured before data is received.
  ## This config item sets the rate in Hz, with 0 disabling the request.
  ##
  ## This frequency should be set to 0 if your software already controls the
  ## rates using REQUEST_DATA_STREAM or MAV_CMD_SET_MESSAGE_INTERVAL
  ## (See https://mavlink.io/en/mavgen_python/howto_requestmessages.html)
  # stream_request_frequency = 4
```

## Metrics

Each supported Mavlink message translates to one metric group, and fields
on the Mavlink message are converted to fields in telegraf.

The name of the Mavlink message is translated into lowercase and any
leading text `message_` is dropped.

For example, the message [ATTITUDE][attitude] will become an `attitude` metric,
with all fields copied from its Mavlink message definition.

[attitude]: https://mavlink.io/en/messages/common.html#ATTITUDE

## Example Output

```text
system_time,source=udp://:5760,sys_id=1 time_unix_usec=1732901334516981i,time_boot_ms=1731552i
ekf_status_report,source=udp://:5760,sys_id=1 velocity_variance=0.006436665542423725,pos_horiz_variance=0.006062425673007965,pos_vert_variance=0.0029854460153728724,compass_variance=0.010930062271654606,terrain_alt_variance=0,airspeed_variance=0
local_position_ned,source=udp://:5760,sys_id=1 time_boot_ms=1731552i,x=-0.010437906719744205,y=-0.02162001095712185,z=-0.0037050051614642143,vx=-0.011906237341463566,vy=-0.02467793971300125,vz=0.012739507481455803
vibration,source=udp://:5760,sys_id=1 time_usec=1731552102i,vibration_x=0.0028534166049212217,vibration_y=0.002792230574414134,vibration_z=0.0028329004999250174,clipping_0=0i,clipping_1=0i,clipping_2=0i
battery_status,source=udp://:5760,sys_id=1 id=0i,temperature=32767i,current_battery=0i,current_consumed=0i,energy_consumed=0i,battery_remaining=100i,time_remaining=0i
ahrs,source=udp://:5760,sys_id=1 omegaix=-0.0012698185164481401,omegaiy=-0.0011798597406595945,omegaiz=-0.0017210562946274877,accel_weight=0,renorm_val=0,error_rp=0.002372326795011759,error_yaw=0.0014012008905410767
ahrs2,source=udp://:5760,sys_id=1 roll=-0.0015893152449280024,pitch=-0.0018129277741536498,yaw=-1.2297048568725586,altitude=0.22999998927116394,lat=450469223i,lng=-834024728i
attitude,source=udp://:5760,sys_id=1 time_boot_ms=1731811i,roll=-0.0011288427049294114,pitch=-0.0013485358795151114,yaw=-1.2430261373519897,rollspeed=-0.00023304438218474388,pitchspeed=-0.00023194786626845598,yawspeed=-0.0008081073756329715 0
global_position_int,source=udp://:5760,sys_id=1 time_boot_ms=1731811i,lat=450469223i,lon=-834024730i,alt=0i,relative_alt=-115i,vx=-1i,vy=-2i,vz=1i,hdg=28878i
gps_raw_int,source=udp://:5760,sys_id=1 time_usec=1731635000i,lat=450469223i,lon=-834024728i,alt=0i,eph=121i,epv=200i,vel=0i,cog=0i,satellites_visible=10i,alt_ellipsoid=0i,hacc=300i,vacc=300i,vel_acc=40i,hdg_acc=0i,yaw=0i
```
