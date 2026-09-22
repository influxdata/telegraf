# UPower Input Plugin

This plugin gathers battery and power-supply information of the local machine
and connected peripherals (e.g. wireless mice, keyboards or headsets) from the
[UPower][upower] daemon via D-Bus.

⭐ Telegraf v1.41.0
🏷️ hardware, system
💻 linux, freebsd

[upower]: https://upower.freedesktop.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather battery and power-supply information via UPower
# This plugin does NOT support Windows and macOS
[[inputs.upower]]
  ## Device types to collect, all devices are collected if empty.
  ## Available types are: line_power, battery, ups, monitor, mouse, keyboard,
  ## pda, phone, media_player, tablet, computer, gaming_input, pen, touchpad,
  ## modem, network, headset, speakers, headphones, video, other_audio,
  ## remote_control, printer, scanner, camera, wearable, toy, bluetooth_generic
  # device_types = []

  ## Timeout for collecting the device information
  # timeout = "5s"
```

## Metrics

Line-power devices (e.g. AC adapters) only report the `online` field, all
other device types report the battery related fields.

- upower
  - tags:
    - native_path (string) - Native path of the device (e.g. `BAT0`)
    - type (string) - Device type (e.g. `battery`, `line_power` or `mouse`)
    - vendor (string) - Vendor of the device, omitted if unknown
    - model (string) - Model of the device, omitted if unknown
    - serial (string) - Serial number of the device, omitted if unknown
    - state (string) - Battery state, one of `unknown`, `charging`,
      `discharging`, `empty`, `fully_charged`, `pending_charge` or
      `pending_discharge`
    - technology (string) - Battery technology, e.g. `lithium_ion`
    - warning_level (string) - Warning level, one of `unknown`, `none`,
      `discharging`, `low`, `critical` or `action`
  - fields:
    - online (boolean) - Whether the line-power device is connected
    - percentage (float) - Charge level in percent
    - energy_wh (float) - Current energy in watt-hours
    - energy_empty_wh (float) - Energy considered empty in watt-hours
    - energy_full_wh (float) - Energy when fully charged in watt-hours
    - energy_full_design_wh (float) - Design energy when fully charged
      in watt-hours
    - energy_rate_w (float) - Charge or discharge rate in watts
    - voltage_v (float) - Voltage in volts
    - temperature_c (float) - Temperature in degrees Celsius
    - capacity_percent (float) - Battery health as percentage of the design
      capacity
    - charge_cycles (int) - Number of charge cycles, `-1` if unknown
    - time_to_empty_sec (int) - Estimated seconds until empty, `0` if unknown
    - time_to_full_sec (int) - Estimated seconds until full, `0` if unknown
    - is_present (boolean) - Whether the battery is present
    - is_rechargeable (boolean) - Whether the battery is rechargeable
    - power_supply (boolean) - Whether the device powers the system

## Example Output

```text
upower,host=laptop,model=5B10W13973,native_path=BAT0,serial=1234,state=discharging,technology=lithium_polymer,type=battery,vendor=LGC,warning_level=none capacity_percent=87.86,charge_cycles=213i,energy_empty_wh=0,energy_full_design_wh=57,energy_full_wh=50.08,energy_rate_w=7.891,energy_wh=41.57,is_present=true,is_rechargeable=true,percentage=83,power_supply=true,temperature_c=0,time_to_empty_sec=18966i,time_to_full_sec=0i,voltage_v=11.98 1758542400000000000
upower,host=laptop,native_path=AC,type=line_power online=false 1758542400000000000
upower,host=laptop,model=MX\ Master\ 3,native_path=hidpp_battery_0,state=discharging,technology=unknown,type=mouse,vendor=Logitech,warning_level=none capacity_percent=0,charge_cycles=0i,energy_empty_wh=0,energy_full_design_wh=0,energy_full_wh=0,energy_rate_w=0,energy_wh=0,is_present=true,is_rechargeable=true,percentage=55,power_supply=false,temperature_c=0,time_to_empty_sec=0i,time_to_full_sec=0i,voltage_v=0 1758542400000000000
```
