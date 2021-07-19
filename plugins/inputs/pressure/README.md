# Pressure input plugin

This plugin gather information about system pressure stall on resources (cpu, memory or io). Only supports linux
OSes with kernel version 4.20 or higher compiled with CONFIG_PSI option

### Configuration:

```toml
# Gather system metrics about Pressure Stall (PSI)
[[inputs.pressure]]
```

### Metrics:

- metric
    - tags:
        - resource (cpu, memory or io)
        - type (some or full)
    - fields:
        - avg10 (float64, gauge)
        - avg60 (float64, gauge)
        - avg300 (float64, gauge)
        - total (uint64, gauge)
### Example Output:

This section shows example output in Line Protocol format.

```
> pressure,host=c6e73f77bce3,resource=cpu,type=some avg10=0,avg300=0,avg60=0,total=75375584i 1613481208000000000
> pressure,host=c6e73f77bce3,resource=memory,type=some avg10=0,avg300=0,avg60=0,total=0i 1613481208000000000
> pressure,host=c6e73f77bce3,resource=memory,type=full avg10=0,avg300=0,avg60=0,total=0i 1613481208000000000
> pressure,host=c6e73f77bce3,resource=io,type=some avg10=0,avg300=0.01,avg60=0,total=46715138i 1613481208000000000
> pressure,host=c6e73f77bce3,resource=io,type=full avg10=0,avg300=0.01,avg60=0,total=44824261i 1613481208000000000
```
