# Temp Input plugin

This input plugin collect temperature.

### Configuration:

```
[[inputs.temp]]
```

### Measurements & Fields:

All fields are float64.

- temp ( unit: °Celsius)

### Tags:

- All measurements have the following tags:
    - host
    - sensor

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter temp --test
* Plugin: temp, Collection 1
> temp,host=localhost,sensor=coretemp_physicalid0_crit temp=100 1531298763000000000
> temp,host=localhost,sensor=coretemp_physicalid0_critalarm temp=0 1531298763000000000
> temp,host=localhost,sensor=coretemp_physicalid0_input temp=100 1531298763000000000
> temp,host=localhost,sensor=coretemp_physicalid0_max temp=100 1531298763000000000
```
