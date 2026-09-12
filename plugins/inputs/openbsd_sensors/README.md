# OpenBSD Sensors Input Plugin

This plugin collects hardware sensor data such as temperatures, voltages, fan
speeds or drive states from the [OpenBSD sensors framework][sensors] using the
`sysctl` command to query the `hw.sensors` tree.

> [!NOTE]
> This plugin only supports OpenBSD as the sensors framework is exposed via
> the `hw.sensors` sysctl tree which does not exist on other platforms.

⭐ Telegraf v1.40.0
🏷️ hardware, system
💻 openbsd

[sensors]: https://man.openbsd.org/sysctl.2#HW_SENSORS

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Collect hardware sensor data from the OpenBSD sensors framework
# This plugin ONLY supports OpenBSD
[[inputs.openbsd_sensors]]
  ## Location of the sysctl binary
  # binary = "/sbin/sysctl"

  ## Collect sensors of the listed devices only, supports glob patterns.
  ## Sensors of all devices are collected if unset or empty.
  # devices = ["cpu*", "acpitz0"]

  ## Maximum time the sysctl binary is allowed to run
  # timeout = "5s"
```

## Metrics

One metric is emitted per sensor. Values are reported in the natural unit of
the sensor type as printed by `sysctl` (e.g. degrees Celsius for `temp`,
Hz for `frequency`, volts for `volt`, RPM for `fan`, percent for `percent`
and `humidity`, seconds for `timedelta`).

- openbsd_sensors
  - tags:
    - device (sensor device, e.g. `cpu0`, `acpitz0` or `softraid0`)
    - sensor (sensor name, e.g. `temp0` or `drive1`)
    - type (sensor type, e.g. `temp`, `frequency` or `drive`)
    - description (sensor description, e.g. `zone temperature`;
      omitted when absent)
  - fields:
    - value (sensor value; `indicator` sensors report `On` as 1 and
      `Off` as 0; `drive` sensors report the kernel `SENSOR_DRIVE_*`
      enum, e.g. `online` is 4 and `failed` is 9; `float`)
    - state (drive state, e.g. `online` or `failed`; `drive` sensors
      only; `string`)
    - status (sensor status, one of `OK`, `WARNING`, `CRITICAL` or
      `UNKNOWN`; omitted when the sensor does not report a status; `string`)
    - status_code (numeric form of `status`, matching OpenBSD
      `SENSOR_S_*`: OK=1, WARNING=2, CRITICAL=3, UNKNOWN=4; only set
      when the sensor value is unknown but a status is present; `float`)

## Example Output

```text
openbsd_sensors,device=cpu0,sensor=temp0,type=temp value=43 1752956161000000000
openbsd_sensors,device=cpu0,sensor=frequency0,type=frequency value=2250000000 1752956161000000000
openbsd_sensors,device=acpitz0,sensor=temp0,type=temp,description=zone\ temperature value=27.8 1752956161000000000
openbsd_sensors,device=softraid0,sensor=drive0,type=drive,description=sd2 value=4,state="online",status="OK" 1752956161000000000
openbsd_sensors,device=nmea0,sensor=indicator0,type=indicator,description=Signal value=1,status="OK" 1752956161000000000
openbsd_sensors,device=nmea0,sensor=timedelta0,type=timedelta,description=GPS\ differential value=-0.000104,status="OK" 1752956161000000000
openbsd_sensors,device=nmea0,sensor=angle0,type=angle,description=Latitude value=22.7643,status="OK" 1752956161000000000
openbsd_sensors,device=foo0,sensor=temp1,type=temp status="UNKNOWN",status_code=4 1752956161000000000
```
