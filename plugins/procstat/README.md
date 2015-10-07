# Telegraf plugin: procstat

#### Description

The procstat plugin can be used to monitor system resource usage by an
individual process using their /proc data.

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

# Measurements
Note: prefix will set by the user, per process.

File descriptor related measurement names:
- procstat_prefix_num_fds value=4

Context switch related measurement names:
- procstat_prefix_voluntary_context_switches value=250
- procstat_prefix_involuntary_context_switches value=0

I/O related measurement names:
- procstat_prefix_read_count value=396
- procstat_prefix_write_count value=1
- procstat_prefix_read_bytes value=1019904
- procstat_prefix_write_bytes value=1

CPU related measurement names:
- procstat_prefix_cpu_user value=0
- procstat_prefix_cpu_system value=0.01
- procstat_prefix_cpu_idle value=0
- procstat_prefix_cpu_nice value=0
- procstat_prefix_cpu_iowait value=0
- procstat_prefix_cpu_irq value=0
- procstat_prefix_cpu_soft_irq value=0
- procstat_prefix_cpu_soft_steal value=0
- procstat_prefix_cpu_soft_stolen value=0
- procstat_prefix_cpu_soft_guest value=0
- procstat_prefix_cpu_soft_guest_nice value=0

Memory related measurement names:
- procstat_prefix_memory_rss value=1777664
- procstat_prefix_memory_vms value=24227840
- procstat_prefix_memory_swap value=282624
