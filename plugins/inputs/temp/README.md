# Temperature Input Plugin

This plugin gathers metrics on system temperatures.

⭐ Telegraf v1.8.0
🏷️ hardware, system
💻 linux, macos, windows

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about temperature
[[inputs.temp]]
  ## Desired output format (Linux only)
  ## Available values are
  ##   v1 -- use pre-v1.22.4 sensor naming, e.g. coretemp_core0_input
  ##   v2 -- use v1.22.4+ sensor naming, e.g. coretemp_core_0_input
  # metric_format = "v2"

  ## Add device tag to distinguish devices with the same name (Linux only)
  # add_device_tag = false
```

## Troubleshooting

On **Windows**, the plugin uses a WMI call that is can be replicated with the
following command:

```shell
wmic /namespace:\\root\wmi PATH MSAcpi_ThermalZoneTemperature
```

If the result is "Not Supported" you may be running in a virtualized environment
and not a physical machine. Additionally, if you still get this result your
motherboard or system may not support querying these values. Finally, you may
be required to run as admin to get the values.

## Metrics

- temp
  - tags:
    - sensor
  - fields:
    - temp (float, celcius)

## Example Output

```text
temp,sensor=coretemp_physicalid0_crit temp=100 1531298763000000000
temp,sensor=coretemp_physicalid0_critalarm temp=0 1531298763000000000
temp,sensor=coretemp_physicalid0_input temp=100 1531298763000000000
temp,sensor=coretemp_physicalid0_max temp=100 1531298763000000000
```
