# Sensors Input Plugin

This plugin collects metrics from hardware sensors: [lm-sensors][lmsensors] on
Linux and the [OpenBSD sensors framework][openbsd_sensors] via
`sysctl hw.sensors`.

> [!NOTE]
> On Linux this plugin requires the lm-sensors package to be installed and
> `sensors` to be executable from Telegraf. On OpenBSD it requires `sysctl`.

⭐ Telegraf v0.10.1
🏷️ hardware, system
💻 linux, openbsd

[lmsensors]: https://en.wikipedia.org/wiki/Lm_sensors
[openbsd_sensors]: https://man.openbsd.org/sysctl.2#HW_SENSORS

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Monitor hardware sensors (lm-sensors on Linux, hw.sensors on OpenBSD)
[[inputs.sensors]]
  ## Linux only: remove numbers from field names.
  ## If true, a field name like 'temp1_input' will be changed to 'temp_input'.
  # remove_numbers = true

  ## Metric version (Linux only). Default 1 keeps tags chip and feature.
  ## Version 2 uses unified tags device, sensor, and type.
  # metric_version = 1

  ## OpenBSD only: sysctl binary and optional device glob filter
  # binary = "/sbin/sysctl"
  # devices = ["cpu*", "acpitz0"]

  ## Timeout is the maximum amount of time that the sensors command can run.
  # timeout = "5s"
```

Linux `metric_version = 1` (the default) keeps the historical `chip` and
`feature` tags so existing dashboards keep working. Set `metric_version = 2`
to use the unified `device`, `sensor`, and `type` tags shared with OpenBSD.
OpenBSD always emits version 2.

## Metrics

### Linux (metric_version = 1, default)

Fields are created dynamically depending on the sensors. All fields are float.

- sensors:
  - tags:
    - chip
    - feature
  - fields:
    - depending on the available sensor information (float)

### Linux (metric_version = 2)

Same fields as version 1. Tags use the unified names; `type` is the feature
name with trailing digits removed (e.g. `temp1` -> `temp`).

- sensors:
  - tags:
    - device (formerly `chip`)
    - sensor (formerly `feature`)
    - type
  - fields:
    - depending on the available sensor information (float)

### OpenBSD

One metric is emitted per sensor. Values are reported in the natural unit of
the sensor type as printed by `sysctl` (e.g. degrees Celsius for `temp`,
Hz for `frequency`, volts for `volt`, RPM for `fan`, percent for `percent`
and `humidity`, seconds for `timedelta`).

- sensors
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

### Linux default (metric_version = 1)

```text
sensors,chip=power_meter-acpi-0,feature=power1 power_average=0,power_average_interval=300 1466751326000000000
sensors,chip=k10temp-pci-00c3,feature=temp1 temp_crit=70,temp_crit_hyst=65,temp_input=29,temp_max=70 1466751326000000000
sensors,chip=k10temp-pci-00cb,feature=temp1 temp_input=29,temp_max=70 1466751326000000000
sensors,chip=k10temp-pci-00d3,feature=temp1 temp_input=27.5,temp_max=70 1466751326000000000
sensors,chip=k10temp-pci-00db,feature=temp1 temp_crit=70,temp_crit_hyst=65,temp_input=29.5,temp_max=70 1466751326000000000
```

### Linux with remove_numbers=false

```text
sensors,chip=power_meter-acpi-0,feature=power1 power1_average=0,power1_average_interval=300 1466753424000000000
sensors,chip=k10temp-pci-00c3,feature=temp1 temp1_crit=70,temp1_crit_hyst=65,temp1_input=29.125,temp1_max=70 1466753424000000000
sensors,chip=k10temp-pci-00cb,feature=temp1 temp1_input=29,temp1_max=70 1466753424000000000
sensors,chip=k10temp-pci-00d3,feature=temp1 temp1_input=29.5,temp1_max=70 1466753424000000000
sensors,chip=k10temp-pci-00db,feature=temp1 temp1_crit=70,temp1_crit_hyst=65,temp1_input=30,temp1_max=70 1466753424000000000
```

### Linux metric_version = 2

```text
sensors,device=k10temp-pci-00c3,sensor=temp1,type=temp temp_crit=70,temp_input=29,temp_max=70 1466751326000000000
sensors,device=power_meter-acpi-0,sensor=power1,type=power power_average=0,power_average_interval=300 1466751326000000000
```

### OpenBSD example

```text
sensors,device=cpu0,sensor=temp0,type=temp value=43 1752956161000000000
sensors,device=cpu0,sensor=frequency0,type=frequency value=2250000000 1752956161000000000
sensors,device=acpitz0,sensor=temp0,type=temp,description=zone\ temperature value=27.8 1752956161000000000
sensors,device=softraid0,sensor=drive0,type=drive,description=sd2 value=4,state="online",status="OK" 1752956161000000000
sensors,device=nmea0,sensor=indicator0,type=indicator,description=Signal value=1,status="OK" 1752956161000000000
sensors,device=foo0,sensor=temp1,type=temp status="UNKNOWN",status_code=4 1752956161000000000
```
