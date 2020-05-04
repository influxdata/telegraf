# SMCIPMITool Input Plugin

Retrieves IPMI sensor data using the SuperMicro specific command line utility
[`SMCIPMITool`](https://www.supermicro.com/en/solutions/management-software/ipmi-utilities).

This plugin is based largely on the [`IPMI Sensor Input Plugin`](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/ipmi_sensor), however it relies on SuperMicro's SMCIPMITool which includes power supply data via the `SMCIPMITool pminfo` command.


Sensor data is retrieved using the following commands:

```
// equilivent to: ipmitool sdr
SMCIPMITool <IP> <USER> <PASS> ipmi sensor

// additional power supply PMBus info
SMCIPMITool <IP> <USER> <PASS> pminfo
```
### Configuration

```toml
# Reads IPMI data via SMCIPMITool
[[inputs.smc_ipmi]]
  ## Path to SMCIPMITool executable
  ## https://www.supermicro.com/en/solutions/management-software/ipmi-utilities
  path = "/usr/bin/smcipmitool/SMCIPMITool"

  # servers = ["USERID:PASSW0RD@(192.168.1.1)"]
  servers = ["userid:password@(192.168.1.1)"]

  ## Retrieve temperature values as celsius "C" or fahrenheit "F"
  ## Defaults to celsius
  # temp_unit = "F"
```

### Measurements

Version 1 schema:
- smc_ipmi:
  - tags:
    - name
    - unit
    - server
  - fields:
    - status (int, 1=ok/0=anything else)
    - value (float)

#### Permissions
IPMI user will need Administrator privileges in order to retrieve sensor data

### Example Output

```
> smc_ipmi,host=dev,name=cpu1_temp,server=192.168.1.1,unit=f status=1i,value=120 1588625062000000000
> smc_ipmi,host=dev,name=cpu2_temp,server=192.168.1.1,unit=f status=1i,value=145 1588625062000000000
> smc_ipmi,host=dev,name=pch_temp,server=192.168.1.1,unit=f status=1i,value=135 1588625062000000000
> smc_ipmi,host=dev,name=system_temp,server=192.168.1.1,unit=f status=1i,value=104 1588625062000000000
...
> smc_ipmi,host=dev,name=pmbus_status,server=192.168.1.1 status=1i 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_input_voltage,server=192.168.1.1,unit=v value=119 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_input_current,server=192.168.1.1,unit=a value=0.95 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_main_output_voltage,server=192.168.1.1,unit=v value=12.09 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_main_output_current,server=192.168.1.1,unit=a value=8 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_temperature_1,server=192.168.1.1,unit=f value=106 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_temperature_2,server=192.168.1.1,unit=f value=113 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_fan_1,server=192.168.1.1,unit=rpm value=6336 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_fan_2,server=192.168.1.1,unit=rpm value=0 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_main_output_power,server=192.168.1.1,unit=w value=96 1588625062000000000
> smc_ipmi,host=dev,name=pmbus_input_power,server=192.168.1.1,unit=w value=113 1588625062000000000
```
