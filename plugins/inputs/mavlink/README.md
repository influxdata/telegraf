# MavLink Input Plugin

The `mavlink` plugin connects to a MavLink-compatible flight controller such as
 [ArduPilot](https://ardupilot.org/) or [PX4](https://px4.io/). and translates
all incoming messages into metrics.

The purpose of this plugin is to allow Telegraf to be used to ingest live
 flight metrics from unmanned systems (drones, planes, boats, etc.)

Warning: This input plugin potentially generates a large amount of data! Use
the configuration to limit the set of messages, or use another telegraf plugin
to filter the output.

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
     # Only used if the mavlink plugin is sending messages, eg.
     # when `stream_request_enable` is enabled (see below.)
     system_id = 254

     # Determines whether the plugin sends requests to stream telemetry,
     # and if enabled, the requested frequency of telemetry in Hz.
     # This setting should be disabled if your software controls rates using
     # REQUEST_DATA_STREAM or MAV_CMD_SET_MESSAGE_INTERVAL
     # (See https://mavlink.io/en/mavgen_python/howto_requestmessages.html#how-to-request--stream-messages)
     stream_request_enable = true
     stream_request_frequency = 4
```

### Note: Mavlink Dialects

This plugin currently only uses the ArduPilot-specific dialect, which also
includes messages from the common Mavlink dialect.

See the [Mavlink docs](https://mavlink.io/en/messages/ardupilotmega.html) for
more info on dialects.

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
system_time,host=chris-ubuntu time_boot_ms=320734i,time_unix_usec=0i 1732249197549336855
ahrs,host=chris-ubuntu omegaix=-0.00021079527505207807,omegaiy=0.0015763355186209083,omegaiz=0.0000307745867758058,accel_weight=0,renorm_val=0,error_rp=0.0011759063927456737,error_yaw=1 1732249197549382215
ahrs2,host=chris-ubuntu altitude=0,lat=0i,lng=0i,roll=0.027588961645960808,pitch=-0.017094312235713005,yaw=-0.40916287899017334 1732249197549417595
hwstatus,host=chris-ubuntu vcc=0i,i2cerr=0i 1732249197549436175
ekf_status_report,host=chris-ubuntu terrain_alt_variance=0,airspeed_variance=0,velocity_variance=0,pos_horiz_variance=0.003580203279852867,pos_vert_variance=0.008646958507597446,compass_variance=0 1732249197549567906
vibration,host=chris-ubuntu clipping_2=0i,time_usec=320734842i,vibration_x=0.022737687453627586,vibration_y=0.0202490147203207,vibration_z=0.026936473324894905,clipping_0=0i,clipping_1=0i 1732249197549646256
battery_status,host=chris-ubuntu current_battery=14i,time_remaining=0i,energy_consumed=0i,id=0i,temperature=32767i,battery_remaining=99i,current_consumed=12i 1732249197550220839
attitude,host=chris-ubuntu yaw=-0.4206164479255676,rollspeed=-0.00021804751304443926,pitchspeed=0.00012629013508558273,yawspeed=-0.001034539774991572,time_boot_ms=320954i,roll=0.02792513370513916,pitch=-0.017007455229759216 1732249197769545832
global_position_int,host=chris-ubuntu vy=0i,lat=0i,lon=0i,alt=2110i,relative_alt=2110i,vx=0i,vz=0i,hdg=33591i,time_boot_ms=320954i 1732249197769617423
sys_status,host=chris-ubuntu current_battery=14i,errors_count2=0i,load=53i,errors_count1=0i,battery_remaining=99i,drop_rate_comm=0i,errors_comm=0i,errors_count3=0i,errors_count4=0i,voltage_battery=3i 1732249197769801674
power_status,host=chris-ubuntu vcc=0i,vservo=0i 1732249197769840494
meminfo,host=chris-ubuntu brkval=0i,freemem=58176i,freemem32=58176i 1732249197769881134
mission_current,host=chris-ubuntu total=0i,mission_mode=0i,mission_id=0i,fence_id=0i,rally_points_id=0i,seq=0i 1732249197769938424
vfr_hud,host=chris-ubuntu throttle=0i,alt=2.109999895095825,climb=-0,airspeed=0,groundspeed=0,heading=335i 1732249197769991115
servo_output_raw,host=chris-ubuntu servo6_raw=0i,servo2_raw=0i,servo3_raw=1500i,servo9_raw=0i,servo1_raw=0i,servo5_raw=1500i,servo10_raw=0i,servo11_raw=0i,servo14_raw=0i,servo15_raw=0i,servo16_raw=0i,time_usec=320955083i,port=0i,servo4_raw=1500i,servo7_raw=0i,servo8_raw=0i,servo12_raw=0i,servo13_raw=0i 1732249197770103425
rc_channels,host=chris-ubuntu chan2_raw=0i,chan5_raw=0i,chan6_raw=0i,chan9_raw=0i,chan10_raw=0i,chan13_raw=0i,chancount=0i,chan1_raw=0i,chan4_raw=0i,chan16_raw=0i,chan18_raw=0i,rssi=0i,time_boot_ms=320955i,chan15_raw=0i,chan17_raw=0i,chan12_raw=0i,chan7_raw=0i,chan8_raw=0i,chan11_raw=0i,chan14_raw=0i,chan3_raw=0i 1732249197770226556
raw_imu,host=chris-ubuntu ymag=0i,zmag=0i,time_usec=320955172i,xacc=-16i,zacc=-1012i,ygyro=-1i,zgyro=0i,yacc=17i,xgyro=0i,xmag=0i,id=0i,temperature=3118i 1732249197770292876
scaled_pressure,host=chris-ubuntu time_boot_ms=320955i,press_abs=994.549560546875,press_diff=0,temperature=2791i,temperature_press_diff=0i 1732249197770337826
gps_raw_int,host=chris-ubuntu cog=0i,satellites_visible=0i,lon=0i,vel=0i,alt_ellipsoid=0i,hacc=0i,vacc=0i,hdg_acc=0i,lat=0i,alt=0i,epv=65535i,yaw=0i,eph=65535i,vel_acc=0i,time_usec=0i 1732249197770433337
system_time,host=chris-ubuntu time_unix_usec=0i,time_boot_ms=320955i 1732249197770457207
ahrs,host=chris-ubuntu error_rp=0.0010012972634285688,error_yaw=1,omegaix=-0.00021079527505207807,omegaiy=0.0015763355186209083,omegaiz=0.0000307745867758058,accel_weight=0,renorm_val=0 1732249197789253764
ahrs2,host=chris-ubuntu lat=0i,lng=0i,roll=0.027666587382555008,pitch=-0.017075257375836372,yaw=-0.4092118442058563,altitude=0 1732249197789308644
hwstatus,host=chris-ubuntu vcc=0i,i2cerr=0i 1732249197789328744
ekf_status_report,host=chris-ubuntu pos_horiz_variance=0.0036376763600856066,pos_vert_variance=0.006598762236535549,compass_variance=0,terrain_alt_variance=0,airspeed_variance=0,velocity_variance=0 1732249197789400524
vibration,host=chris-ubuntu vibration_z=0.029913151636719704,clipping_0=0i,clipping_1=0i,clipping_2=0i,time_usec=320974603i,vibration_x=0.025675609707832336,vibration_y=0.022661570459604263 1732249197789445705
battery_status,host=chris-ubuntu current_battery=14i,energy_consumed=0i,temperature=32767i,current_consumed=12i,battery_remaining=99i,time_remaining=0i,id=0i 1732249197789605765
attitude,host=chris-ubuntu yawspeed=-0.0006323250127024949,time_boot_ms=321214i,roll=0.027931861579418182,pitch=-0.017001383006572723,yaw=-0.42084062099456787,rollspeed=-0.000111618239316158,pitchspeed=0.00003287754952907562 1732249198028859780
global_position_int,host=chris-ubuntu relative_alt=2107i,vz=0i,hdg=33589i,time_boot_ms=321214i,lat=0i,lon=0i,alt=2100i,vx=0i,vy=0i 1732249198028926881
sys_status,host=chris-ubuntu voltage_battery=0i,drop_rate_comm=0i,errors_comm=0i,errors_count1=0i,errors_count2=0i,errors_count3=0i,battery_remaining=99i,current_battery=14i,load=51i,errors_count4=0i 1732249198029084052
power_status,host=chris-ubuntu vcc=0i,vservo=0i 1732249198029116442
meminfo,host=chris-ubuntu brkval=0i,freemem=58176i,freemem32=58176i 1732249198029155772
mission_current,host=chris-ubuntu seq=0i,total=0i,mission_mode=0i,mission_id=0i,fence_id=0i,rally_points_id=0i 1732249198029206172
vfr_hud,host=chris-ubuntu alt=2.0999999046325684,climb=-0,airspeed=0,groundspeed=0,heading=335i,throttle=0i 1732249198029248892
servo_output_raw,host=chris-ubuntu servo2_raw=0i,servo11_raw=0i,servo13_raw=0i,servo14_raw=0i,time_usec=321214595i,servo1_raw=0i,servo4_raw=1500i,servo5_raw=1500i,servo6_raw=0i,servo9_raw=0i,servo10_raw=0i,servo12_raw=0i,port=0i,servo8_raw=0i,servo15_raw=0i,servo16_raw=0i,servo7_raw=0i,servo3_raw=1500i 1732249198029348773
rc_channels,host=chris-ubuntu chan18_raw=0i,rssi=0i,chan9_raw=0i,chan13_raw=0i,chan14_raw=0i,chan2_raw=0i,chan15_raw=0i,chan11_raw=0i,chancount=0i,chan5_raw=0i,chan8_raw=0i,chan4_raw=0i,chan6_raw=0i,chan7_raw=0i,chan10_raw=0i,chan12_raw=0i,time_boot_ms=321214i,chan1_raw=0i,chan3_raw=0i,chan16_raw=0i,chan17_raw=0i 1732249198029460344
raw_imu,host=chris-ubuntu temperature=3118i,time_usec=321214685i,xacc=-16i,ygyro=-1i,zgyro=0i,ymag=0i,id=0i,yacc=17i,zacc=-1012i,xgyro=0i,xmag=0i,zmag=0i 1732249198029519524
scaled_pressure,host=chris-ubuntu press_diff=0,temperature=2791i,temperature_press_diff=0i,time_boot_ms=321214i,press_abs=994.5499267578125 1732249198029564314
gps_raw_int,host=chris-ubuntu alt=0i,vel=0i,alt_ellipsoid=0i,vacc=0i,eph=65535i,cog=0i,satellites_visible=0i,vel_acc=0i,yaw=0i,time_usec=0i,epv=65535i,hdg_acc=0i,lat=0i,lon=0i,hacc=0i 1732249198029637994
```
