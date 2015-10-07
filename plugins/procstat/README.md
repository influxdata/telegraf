# Telegraf plugin: procstat

#### Description

The procstat plugin can be used to monitor system resource usage by an
individual process using their /proc data.

The plugin will tag processes by their PID and their process name.

Processes can be specified either by pid file or by executable name. Procstat
plugin will use `pgrep` when executable name is provided to obtain the pid.
Proctstas plugin will transmit IO, memory, cpu, file descriptor related
measurements for every process specified. A prefix can be set to isolate
individual process specific measurements.

Example:

```
    [procstat]

    [[procstat.specifications]]
    exe = "influxd"
    prefix = "influxd"

    [[procstat.specifications]]
    pid_file = "/var/run/lxc/dnsmasq.pid"
```

The above configuration would result in output like:

```
[...]
> [name="dnsmasq" pid="44979"] procstat_cpu_user value=0.14
> [name="dnsmasq" pid="44979"] procstat_cpu_system value=0.07
[...]
> [name="influxd" pid="34337"] procstat_influxd_cpu_user value=25.43
> [name="influxd" pid="34337"] procstat_influxd_cpu_system value=21.82
```

# Measurements
Note: prefix can be set by the user, per process.

File descriptor related measurement names:
- procstat_[prefix_]num_fds value=4

Context switch related measurement names:
- procstat_[prefix_]voluntary_context_switches value=250
- procstat_[prefix_]involuntary_context_switches value=0

I/O related measurement names:
- procstat_[prefix_]read_count value=396
- procstat_[prefix_]write_count value=1
- procstat_[prefix_]read_bytes value=1019904
- procstat_[prefix_]write_bytes value=1

CPU related measurement names:
- procstat_[prefix_]cpu_user value=0
- procstat_[prefix_]cpu_system value=0.01
- procstat_[prefix_]cpu_idle value=0
- procstat_[prefix_]cpu_nice value=0
- procstat_[prefix_]cpu_iowait value=0
- procstat_[prefix_]cpu_irq value=0
- procstat_[prefix_]cpu_soft_irq value=0
- procstat_[prefix_]cpu_soft_steal value=0
- procstat_[prefix_]cpu_soft_stolen value=0
- procstat_[prefix_]cpu_soft_guest value=0
- procstat_[prefix_]cpu_soft_guest_nice value=0

Memory related measurement names:
- procstat_[prefix_]memory_rss value=1777664
- procstat_[prefix_]memory_vms value=24227840
- procstat_[prefix_]memory_swap value=282624
