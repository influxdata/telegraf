# Procstat plugin

The procstat plugin can be used to monitor system resource usage by an
individual process.

Processes can be specified either by pid file or by executable name. Procstat
plugin will use `pgrep` when executable name is provided to obtain the pid. Proctsta plugin will transmit IO, memory, cpu, file descriptor related measurements for every process specified. A prefix can be set to isolate individual process specific measurements.

Example:

```
[procstat]

[[procstat.specifications]]
  exe = "influxd"
  prefix = "influxd"

[[procstat.specifications]]
  pid_file = "/var/run/lxc/dnsmasq.pid"
  prefix = "dnsmasq"
```
