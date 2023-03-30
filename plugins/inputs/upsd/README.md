# UPSD Input Plugin

This plugin reads data of one or more Uninterruptible Power Supplies
from an upsd daemon using its NUT network protocol.

## Requirements

upsd should be installed and it's daemon should be running.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Monitor UPSes connected via Network UPS Tools
[[inputs.upsd]]
  ## A running NUT server to connect to.
  ## IPv6 addresses must be enclosed in brackets (e.g. "[::1]")
  # server = "127.0.0.1"
  # port = 3493
  # username = "user"
  # password = "password"

  ## Force parsing numbers as floats
  ## It is highly recommended to enable this setting to parse numbers
  ## consistently as floats to avoid database conflicts where some numbers are
  ## parsed as integers and others as floats.
  # force_float = false
```

## Metrics

This implementation tries to maintain compatibility with the apcupsd metrics:

- upsd
  - tags:
    - serial
    - ups_name
    - model
  - fields:
    - status_flags ([status-bits][rfc9271-sec5.1])
    - input_voltage
    - load_percent
    - battery_charge_percent
    - time_left_ns
    - output_voltage
    - internal_temp
    - battery_voltage
    - input_frequency
    - battery_date
    - nominal_input_voltage
    - nominal_battery_voltage
    - nominal_power
    - firmware

With the exception of:

- tags:
  - status (string representing the set status_flags)
- fields:
  - time_on_battery_ns

[rfc9271-sec5.1]: https://www.rfc-editor.org/rfc/rfc9271.html#section-5.1

## Example Output

```text
upsd,serial=AS1231515,ups_name=name1 load_percent=9.7,time_left_ns=9800000,output_voltage=230.4,internal_temp=32.4,battery_voltage=27.4,input_frequency=50.2,input_voltage=230.4,battery_charge_percent=100,status_flags=8i 1490035922000000000
```
