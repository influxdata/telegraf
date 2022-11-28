# APCUPSD Input Plugin

This plugin reads data from an apcupsd daemon over its NIS network protocol.

## Requirements

apcupsd should be installed and it's daemon should be running.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Monitor APC UPSes connected to apcupsd
[[inputs.apcupsd]]
  # A list of running apcupsd server to connect to.
  # If not provided will default to tcp://127.0.0.1:3551
  servers = ["tcp://127.0.0.1:3551"]

  ## Timeout for dialing server.
  timeout = "5s"
```

## Metrics

- apcupsd
  - tags:
    - serial
    - ups_name
    - status (string representing the set status_flags)
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
    - time_on_battery_ns
    - cumulative_time_on_battery_ns
    - nominal_input_voltage
    - nominal_battery_voltage
    - nominal_power
    - firmware
    - battery_date
    - last_transfer
    - number_transfers

## Example Output

```shell
apcupsd,serial=AS1231515,status=ONLINE,ups_name=name1 time_on_battery=0,load_percent=9.7,time_left_minutes=98,output_voltage=230.4,internal_temp=32.4,battery_voltage=27.4,input_frequency=50.2,input_voltage=230.4,battery_charge_percent=100,status_flags=8i 1490035922000000000
```

[status-bits]: http://www.apcupsd.org/manual/manual.html#status-bits
