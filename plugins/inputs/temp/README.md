# Temp Input plugin

The temp input plugin gather metrics on system temperature.  This plugin is
meant to be multi platform and uses platform specific collection methods.

Currently supports Linux and Windows.

### Configuration:

```
[[inputs.temp]]
```

### Metrics:

- temp
  - tags:
    - sensor
  - fields:
    - temp (float, celcius)

### Example Output:

```
temp,sensor=coretemp_physicalid0_crit temp=100 1531298763000000000
temp,sensor=coretemp_physicalid0_critalarm temp=0 1531298763000000000
temp,sensor=coretemp_physicalid0_input temp=100 1531298763000000000
temp,sensor=coretemp_physicalid0_max temp=100 1531298763000000000
```
