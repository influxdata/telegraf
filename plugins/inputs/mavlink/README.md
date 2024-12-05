# MavLink Input Plugin

The `mavlink` plugin connects to a MavLink-compatible flight controller such as
 [ArduPilot](https://ardupilot.org/) or [PX4](https://px4.io/). and translates
all incoming messages into metrics.

The purpose of this plugin is to allow Telegraf to be used to ingest live
 flight metrics from unmanned systems (drones, planes, boats, etc.)

Warning: This input plugin potentially generates a large amount of data! Use
the configuration to limit the set of messages, or use another telegraf plugin
to filter the output.

This plugin currently uses the ArduPilot-specific Mavlink dialect. See the
[Mavlink docs](https://mavlink.io/en/messages/ardupilotmega.html) for more
info on dialects and the various messages that this plugin can receive.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from a Mavlink connection to a flight controller.
[[inputs.mavlink]]
     ## Flight controller URL. Must be a valid Mavlink connection string in one
     ## of the following formats:
     ##
     ## - Serial port:  serial://<device name>:<baud rate> 
     ##            eg: "serial:///dev/ttyACM0:57600"
     ## 
     ## - TCP client:   tcp://<target ip or hostname>:<port>
     ##            eg: "tcp://192.168.1.12:14550"
     ## 
     ## - UDP client:   udp://<target ip or hostname>:<port>
     ##            eg: "udp://192.168.1.12:14550"
     ## 
     ## - UDP server:   udp://:<listen port>
     ##            eg: "udp://:14540"
     ## 
     ## The meaning of each of these modes is documented by
     ## https://mavsdk.mavlink.io/v1.4/en/cpp/guide/connections.html.
     fcu_url = "udp://:14540"

     ## Filter to specific messages. Only the messages in this list will be parsed.
     ## If blank or unset, all messages will be accepted.
     ## Each message in this list should be lowercase camel_case, with "message_"
     ## prefix removed, eg: "global_position_int", "attitude"
     # message_filter = []

     ## Mavlink system ID for Telegraf
     ## Only used if the mavlink plugin is sending messages, eg.
     ## when `stream_request_enable` is enabled (see below.)
     system_id = 254

     ## Determines whether the plugin sends requests to stream telemetry,
     ## and if enabled, the requested frequency of telemetry in Hz.
     ## This setting should be disabled if your software controls rates using
     ## REQUEST_DATA_STREAM or MAV_CMD_SET_MESSAGE_INTERVAL
     ## (See https://mavlink.io/en/mavgen_python/howto_requestmessages.html#how-to-request--stream-messages)
     stream_request_enable = true
     stream_request_frequency = 4
```

## Metrics

Each supported Mavlink message translates to one metric group, and fields
on the Mavlink message are converted to fields in telegraf.

The name of the Mavlink message is translated into lowercase and any
leading text `message_` is dropped.

For example, [MESSAGE_ATTITUDE](https://mavlink.io/en/messages/common.html)
will become an `attitude` metric, with all fields copied from its Mavlink
message definition.

## Example Output

_`mavlink` input plugin connected to ArduPilot SITL and the `file` output
plugin:_

```text
system_time,fcu_url=udp://:5760,sys_id=1 time_unix_usec=1732901334516981i,time_boot_ms=1731552i 0
simstate,fcu_url=udp://:5760,sys_id=1 roll=0,pitch=0,yaw=-1.2217304706573486,xacc=0,yacc=0,zacc=-9.806650161743164,xgyro=0,ygyro=0,zgyro=0,lat=450469223i,lng=-834024728i 0
ekf_status_report,fcu_url=udp://:5760,sys_id=1 velocity_variance=0.006436665542423725,pos_horiz_variance=0.006062425673007965,pos_vert_variance=0.0029854460153728724,compass_variance=0.010930062271654606,terrain_alt_variance=0,airspeed_variance=0 0
local_position_ned,fcu_url=udp://:5760,sys_id=1 time_boot_ms=1731552i,x=-0.010437906719744205,y=-0.02162001095712185,z=-0.0037050051614642143,vx=-0.011906237341463566,vy=-0.02467793971300125,vz=0.012739507481455803 0
vibration,fcu_url=udp://:5760,sys_id=1 time_usec=1731552102i,vibration_x=0.0028534166049212217,vibration_y=0.002792230574414134,vibration_z=0.0028329004999250174,clipping_0=0i,clipping_1=0i,clipping_2=0i 0
battery_status,fcu_url=udp://:5760,sys_id=1 id=0i,temperature=32767i,current_battery=0i,current_consumed=0i,energy_consumed=0i,battery_remaining=100i,time_remaining=0i 0
statustext,fcu_url=udp://:5760,sys_id=1 text="Field Elevation Set: 0m",id=0i,chunk_seq=0i 0
ahrs,fcu_url=udp://:5760,sys_id=1 omegaix=-0.0012698185164481401,omegaiy=-0.0011798597406595945,omegaiz=-0.0017210562946274877,accel_weight=0,renorm_val=0,error_rp=0.002372326795011759,error_yaw=0.0014012008905410767 0
ahrs2,fcu_url=udp://:5760,sys_id=1 roll=-0.0015893152449280024,pitch=-0.0018129277741536498,yaw=-1.2297048568725586,altitude=0.22999998927116394,lat=450469223i,lng=-834024728i 0
attitude,fcu_url=udp://:5760,sys_id=1 time_boot_ms=1731811i,roll=-0.0011288427049294114,pitch=-0.0013485358795151114,yaw=-1.2430261373519897,rollspeed=-0.00023304438218474388,pitchspeed=-0.00023194786626845598,yawspeed=-0.0008081073756329715 0
global_position_int,fcu_url=udp://:5760,sys_id=1 time_boot_ms=1731811i,lat=450469223i,lon=-834024730i,alt=0i,relative_alt=-115i,vx=-1i,vy=-2i,vz=1i,hdg=28878i 0
vfr_hud,fcu_url=udp://:5760,sys_id=1 airspeed=0,groundspeed=0.027561495080590248,heading=288i,throttle=0i,alt=0,climb=-0.011526756919920444 0
sys_status,fcu_url=udp://:5760,sys_id=1 load=0i,voltage_battery=12600i,current_battery=0i,battery_remaining=100i,drop_rate_comm=0i,errors_comm=0i,errors_count1=0i,errors_count2=0i,errors_count3=0i,errors_count4=0i 0
power_status,fcu_url=udp://:5760,sys_id=1 vcc=5000i,vservo=0i 0
meminfo,fcu_url=udp://:5760,sys_id=1 brkval=0i,freemem=65535i,freemem32=131072i 0
mission_current,fcu_url=udp://:5760,sys_id=1 seq=0i,total=0i,mission_mode=0i,mission_id=0i,fence_id=0i,rally_points_id=0i 0
servo_output_raw,fcu_url=udp://:5760,sys_id=1 time_usec=1731811998i,port=0i,servo1_raw=1500i,servo2_raw=0i,servo3_raw=1500i,servo4_raw=0i,servo5_raw=0i,servo6_raw=0i,servo7_raw=0i,servo8_raw=0i,servo9_raw=0i,servo10_raw=0i,servo11_raw=0i,servo12_raw=0i,servo13_raw=0i,servo14_raw=0i,servo15_raw=0i,servo16_raw=0i 0
rc_channels,fcu_url=udp://:5760,sys_id=1 time_boot_ms=1731811i,chancount=8i,chan1_raw=1500i,chan2_raw=1500i,chan3_raw=1500i,chan4_raw=1500i,chan5_raw=1800i,chan6_raw=1000i,chan7_raw=1000i,chan8_raw=1800i,chan9_raw=0i,chan10_raw=0i,chan11_raw=0i,chan12_raw=0i,chan13_raw=0i,chan14_raw=0i,chan15_raw=0i,chan16_raw=0i,chan17_raw=0i,chan18_raw=0i,rssi=255i 0
raw_imu,fcu_url=udp://:5760,sys_id=1 time_usec=1731811998i,xacc=0i,yacc=0i,zacc=-1001i,xgyro=1i,ygyro=0i,zgyro=0i,xmag=84i,ymag=159i,zmag=508i,id=0i,temperature=4493i 0
scaled_imu2,fcu_url=udp://:5760,sys_id=1 time_boot_ms=1731811i,xacc=0i,yacc=0i,zacc=-1001i,xgyro=1i,ygyro=0i,zgyro=1i,xmag=84i,ymag=159i,zmag=508i,temperature=4493i 0
scaled_imu3,fcu_url=udp://:5760,sys_id=1 time_boot_ms=1731811i,xacc=0i,yacc=0i,zacc=0i,xgyro=0i,ygyro=0i,zgyro=0i,xmag=84i,ymag=159i,zmag=508i,temperature=0i 0
scaled_pressure,fcu_url=udp://:5760,sys_id=1 time_boot_ms=1731811i,press_abs=1013.2387084960938,press_diff=0,temperature=3499i,temperature_press_diff=0i 0
scaled_pressure2,fcu_url=udp://:5760,sys_id=1 time_boot_ms=1731811i,press_abs=1013.2310791015625,press_diff=0,temperature=3499i,temperature_press_diff=0i 0
gps_raw_int,fcu_url=udp://:5760,sys_id=1 time_usec=1731635000i,lat=450469223i,lon=-834024728i,alt=0i,eph=121i,epv=200i,vel=0i,cog=0i,satellites_visible=10i,alt_ellipsoid=0i,hacc=300i,vacc=300i,vel_acc=40i,hdg_acc=0i,yaw=0i 0
```
