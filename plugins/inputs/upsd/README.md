# UPSD Input Plugin

This plugin reads data of one or more Uninterruptible Power Supplies 
from an upsd daemon using its NUT network protocol.

### Requirements

upsd should be installed and it's daemon should be running.

### Configuration

```toml
[[inputs.upsd]]
  ## A running NUT server to connect to.
  # If not provided will default to 127.0.0.1
  # server = "127.0.0.1"
  # username = "user"
  # password = "password"
  ## Timeout for dialing server.
  # connectionTimeout = "10s"
  ## Read/write operation timeout.
  # opTimeout = "10s"
```

## Metrics
This implementation tries to maintain compatibility with the apcupsd metrics:

- upsd
  - tags:
    - serial
    - ups_name
    - model
  - fields:
    - status_flags ([status-bits][])
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

## Example Output

```
upsd,serial=AS1231515,ups_name=name1 load_percent=9.7,time_left_ns=9800000,output_voltage=230.4,internal_temp=32.4,battery_voltage=27.4,input_frequency=50.2,input_voltage=230.4,battery_charge_percent=100,status_flags=8i 1490035922000000000
```
