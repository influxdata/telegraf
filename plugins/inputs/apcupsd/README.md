# apcupsd Input Plugin

This plugin reads data from apcupsd daemon over its NIS network protocol

## Requirements

apcupsd should be installed and its daemon running

## Configuration

```toml
[[inputs.apcupsd]]
  # a list of running apcupsd server to connect to. 
  # If not provided will default to 127.0.0.1:3551
  servers = ["127.0.0.1:3551"]
```

## Measurements

- apcupsd
  - status
  - input_voltage
  - load_percent
  - battery_charge_percent
  - time_left_minutes
  - output_voltage
  - internal_temp
  - battery_voltage
  - input_frequency
  - time_on_battery

Tags:
- serial
- ups_name



## Example output

```
> apcupsd,serial=AS1231515,ups_name=name1,host=server1 time_on_battery=0,load_percent=9.7,time_left_minutes=98,output_voltage=230.4,internal_temp=32.4,battery_voltage=27.4,input_frequency=50.2,online=true,input_voltage=230.4,battery_charge_percent=100 1490035922000000000
```
