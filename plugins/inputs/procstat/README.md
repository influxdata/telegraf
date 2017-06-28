# Telegraf plugin: procstat

#### Description

The procstat plugin can be used to monitor system resource usage by an
individual process using their /proc data.

Processes can be specified either by pid file, by executable name, by command
line pattern matching, or by username (in this order or priority. Procstat
plugin will use `pgrep` when executable name is provided to obtain the pid.
Procstat plugin will transmit IO, memory, cpu, file descriptor related
measurements for every process specified. A prefix can be set to isolate
individual process specific measurements.

The plugin will tag processes according to how they are specified in the configuration. If a pid file is used, a "pidfile" tag will be generated.
On the other hand, if an executable is used an "exe" tag will be generated. Possible tag names:

* pidfile
* exe
* pattern
* user

Additionally the plugin will tag processes by their PID (pid_tag = true in the config) and their process name:

* pid
* process_name

Example:

```
[[inputs.procstat]]
  exe = "influxd"
  prefix = "influxd"

[[inputs.procstat]]
  pid_file = "/var/run/lxc/dnsmasq.pid"
```

The above configuration would result in output like:

```
> procstat,pidfile=/var/run/lxc/dnsmasq.pid,process_name=dnsmasq,pid=44979 cpu_user=0.14,cpu_system=0.07
> procstat,exe=influxd,process_name=influxd,pid=34337 influxd_cpu_user=25.43,influxd_cpu_system=21.82
```

# Measurements
Note: prefix can be set by the user, per process.


Threads related measurement names:
- procstat_[prefix_]num_threads value=5

File descriptor related measurement names (*telegraf* needs to run as **root**):
- procstat_[prefix_]num_fds value=4

Context switch related measurement names:
- procstat_[prefix_]voluntary_context_switches value=250
- procstat_[prefix_]involuntary_context_switches value=0

I/O related measurement names (*telegraf* needs to run as **root**):
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
