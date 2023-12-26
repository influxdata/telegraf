# Pressure Stall Information (PSI) Input Plugin

A plugin to gather Pressure Stall Information from the Linux kernel from `/proc/pressure/{cpu,memory,io}`.

Kernel version 4.20 or later is required.

Examples:

```shell
# /proc/pressure/cpu
some avg10=1.53 avg60=1.87 avg300=1.73 total=1088168194

# /proc/pressure/memory
some avg10=0.00 avg60=0.00 avg300=0.00 total=3463792
full avg10=0.00 avg60=0.00 avg300=0.00 total=1429641

# /proc/pressure/io
some avg10=0.00 avg60=0.00 avg300=0.00 total=68568296
full avg10=0.00 avg60=0.00 avg300=0.00 total=54982338
```

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about Pressure Stall Information (PSI)
# Requires Linux kernel v4.20+
[[inputs.psi]]
  # no configuration
```

## Metrics

- pressure
  - tags:
    - resource: cpu, memory, or io
    - type: some or full
  - fields: avg10, avg60, avg300 for 10-second, 1-minute, and 5-minute averages
- pressureTotal
  - tags: same as pressure
  - fields: total

Note that the combination for `resource=cpu,type=full` is omitted because it is always zero.

## Example Output

```text
pressure,resource=cpu,type=some avg10=0.00,avg60=0.00,avg300=0.00 1700000000000000000
pressure,resource=memory,type=some avg10=0.00,avg60=0.00,avg300=0.00 1700000000000000000
pressure,resource=memory,type=full avg10=0.00,avg60=0.00,avg300=0.00 1700000000000000000
pressure,resource=io,type=some avg10=21.35,avg60=19.23,avg300=13.48 1700000000000000000
pressure,resource=io,type=full avg10=21.26,avg60=19.10,avg300=13.39 1700000000000000000
pressureTotal,resource=cpu,type=some total=657544125 1700000000000000000
pressureTotal,resource=memory,type=some total=62167126 1700000000000000000
pressureTotal,resource=memory,type=full total=60968885 1700000000000000000
pressureTotal,resource=io,type=some total=6507666709 1700000000000000000
pressureTotal,resource=io,type=full total=6399637024 1700000000000000000
```

## Credits

Part of this plugin was derived from [gridscale/linux-psi-telegraf-plugin](https://github.com/gridscale/linux-psi-telegraf-plugin), available under the same MIT license.
