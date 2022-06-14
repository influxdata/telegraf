# Temperature Input Plugin

The temp input plugin gather metrics on system temperature.  This plugin is
meant to be multi platform and uses platform specific collection methods.

Currently supports Linux and Windows.

## Configuration

```toml @sample.conf
# Read metrics about temperature
[[inputs.temp]]
  # no configuration
```

## Metrics

- temp
  - tags:
    - sensor
  - fields:
    - temp (float, celcius)

## Troubleshooting

On **Windows**, the plugin uses a WMI call that is can be replicated with the
following command:

```shell
wmic /namespace:\\root\wmi PATH MSAcpi_ThermalZoneTemperature
```

## Example Output

For `output = "measurement" (default) the output will look like this

```shell
temp,sensor=coretemp_physicalid0_crit temp=100 1531298763000000000
temp,sensor=coretemp_physicalid0_critalarm temp=0 1531298763000000000
temp,sensor=coretemp_physicalid0_input temp=100 1531298763000000000
temp,sensor=coretemp_physicalid0_max temp=100 1531298763000000000
```

For `output = "field" the output will look like this

```shell
temp,sensor=coretemp_physicalid0 crit=100 temp=100 high=100 1531298763000000000
```
