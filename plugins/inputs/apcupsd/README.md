# APC UPSD Input Plugin

This plugin gathers data from one or more [apcupsd daemon][apcupsd_daemon] over
the NIS network protocol. To query a server, the daemon must be running and be
accessible.

⭐ Telegraf v1.12.0
🏷️ hardware, server
💻 all

[apcupsd_daemon]: https://sourceforge.net/projects/apcupsd/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

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

```text
apcupsd,serial=AS1231515,status=ONLINE,ups_name=name1 time_on_battery=0,load_percent=9.7,time_left_minutes=98,output_voltage=230.4,internal_temp=32.4,battery_voltage=27.4,input_frequency=50.2,input_voltage=230.4,battery_charge_percent=100,status_flags=8i 1490035922000000000
```

[status-bits]: http://www.apcupsd.org/manual/manual.html#status-bits
