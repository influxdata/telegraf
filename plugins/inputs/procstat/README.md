# Telegraf plugin: procstat

#### Description

The procstat plugin can be used to monitor system resource usage by an
individual process using their /proc data.

Processes can be specified either by pid file, by executable name, by command
line pattern matching, by username, by systemd unit name, or by cgroup name/path
(in this order or priority). Procstat plugin will use `pgrep` when executable
name is provided to obtain the pid. Procstat plugin will transmit IO, memory,
cpu, file descriptor related measurements for every process specified. A prefix
can be set to isolate individual process specific measurements.

The plugin will tag processes according to how they are specified in the configuration. If a pid file is used, a "pidfile" tag will be generated.
On the other hand, if an executable is used an "exe" tag will be generated. Possible tag names:

* pidfile
* exe
* pattern
* user
* systemd_unit
* cgroup

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
> procstat,pidfile=/var/run/lxc/dnsmasq.pid,process_name=dnsmasq rlimit_file_locks_soft=2147483647i,rlimit_signals_pending_hard=1758i,voluntary_context_switches=478i,read_bytes=307200i,cpu_time_user=0.01,cpu_time_guest=0,memory_swap=0i,memory_locked=0i,rlimit_num_fds_hard=4096i,rlimit_nice_priority_hard=0i,num_fds=11i,involuntary_context_switches=20i,read_count=23i,memory_rss=1388544i,rlimit_memory_rss_soft=2147483647i,rlimit_memory_rss_hard=2147483647i,nice_priority=20i,rlimit_cpu_time_hard=2147483647i,cpu_time=0i,write_bytes=0i,cpu_time_idle=0,cpu_time_nice=0,memory_data=229376i,memory_stack=135168i,rlimit_cpu_time_soft=2147483647i,rlimit_memory_data_hard=2147483647i,rlimit_memory_locked_hard=65536i,rlimit_signals_pending_soft=1758i,write_count=11i,cpu_time_iowait=0,cpu_time_steal=0,cpu_time_stolen=0,rlimit_memory_stack_soft=8388608i,cpu_time_system=0.02,cpu_time_guest_nice=0,rlimit_memory_locked_soft=65536i,rlimit_memory_vms_soft=2147483647i,rlimit_file_locks_hard=2147483647i,rlimit_realtime_priority_hard=0i,pid=828i,num_threads=1i,cpu_time_soft_irq=0,rlimit_memory_vms_hard=2147483647i,rlimit_realtime_priority_soft=0i,memory_vms=15884288i,rlimit_memory_stack_hard=2147483647i,cpu_time_irq=0,rlimit_memory_data_soft=2147483647i,rlimit_num_fds_soft=1024i,signals_pending=0i,rlimit_nice_priority_soft=0i,realtime_priority=0i
> procstat,exe=influxd,process_name=influxd rlimit_num_fds_hard=16384i,rlimit_signals_pending_hard=1758i,realtime_priority=0i,rlimit_memory_vms_hard=2147483647i,rlimit_signals_pending_soft=1758i,cpu_time_stolen=0,rlimit_memory_stack_hard=2147483647i,rlimit_realtime_priority_hard=0i,cpu_time=0i,pid=500i,voluntary_context_switches=975i,cpu_time_idle=0,memory_rss=3072000i,memory_locked=0i,rlimit_nice_priority_soft=0i,signals_pending=0i,nice_priority=20i,read_bytes=823296i,cpu_time_soft_irq=0,rlimit_memory_data_hard=2147483647i,rlimit_memory_locked_soft=65536i,write_count=8i,cpu_time_irq=0,memory_vms=33501184i,rlimit_memory_stack_soft=8388608i,cpu_time_iowait=0,rlimit_memory_vms_soft=2147483647i,rlimit_nice_priority_hard=0i,num_fds=29i,memory_data=229376i,rlimit_cpu_time_soft=2147483647i,rlimit_file_locks_soft=2147483647i,num_threads=1i,write_bytes=0i,cpu_time_steal=0,rlimit_memory_rss_hard=2147483647i,cpu_time_guest=0,cpu_time_guest_nice=0,cpu_usage=0,rlimit_memory_locked_hard=65536i,rlimit_file_locks_hard=2147483647i,involuntary_context_switches=38i,read_count=16851i,memory_swap=0i,rlimit_memory_data_soft=2147483647i,cpu_time_user=0.11,rlimit_cpu_time_hard=2147483647i,rlimit_num_fds_soft=16384i,rlimit_realtime_priority_soft=0i,cpu_time_system=0.27,cpu_time_nice=0,memory_stack=135168i,rlimit_memory_rss_soft=2147483647i
```

# Measurements
Note: prefix can be set by the user, per process.


Threads related measurement names:
- procstat_[prefix_]num_threads value=5

File descriptor related measurement names (*telegraf* needs to run as **root**):
- procstat_[prefix_]num_fds value=4

Priority related measurement names:
- procstat_[prefix_]realtime_priority value=0
- procstat_[prefix_]nice_priority value=20

Signals related measurement names:
- procstat_[prefix_]signals_pending value=0

Context switch related measurement names:
- procstat_[prefix_]voluntary_context_switches value=250
- procstat_[prefix_]involuntary_context_switches value=0

I/O related measurement names (*telegraf* needs to run as **root**):
- procstat_[prefix_]read_count value=396
- procstat_[prefix_]write_count value=1
- procstat_[prefix_]read_bytes value=1019904
- procstat_[prefix_]write_bytes value=1

CPU related measurement names:
- procstat_[prefix_]cpu_time value=0.01
- procstat_[prefix_]cpu_time_user value=0
- procstat_[prefix_]cpu_time_system value=0.01
- procstat_[prefix_]cpu_time_idle value=0
- procstat_[prefix_]cpu_time_nice value=0
- procstat_[prefix_]cpu_time_iowait value=0
- procstat_[prefix_]cpu_time_irq value=0
- procstat_[prefix_]cpu_time_soft_irq value=0
- procstat_[prefix_]cpu_time_steal value=0
- procstat_[prefix_]cpu_time_stolen value=0
- procstat_[prefix_]cpu_time_guest value=0
- procstat_[prefix_]cpu_time_guest_nice value=0

Memory related measurement names:
- procstat_[prefix_]memory_rss value=1777664
- procstat_[prefix_]memory_vms value=24227840
- procstat_[prefix_]memory_swap value=282624
- procstat_[prefix_]memory_data value=229376
- procstat_[prefix_]memory_stack value=135168
- procstat_[prefix_]memory_locked value=0

Resource limits:
- procstat_[prefix_]rlimit_cpu_time_hard value=2147483647
- procstat_[prefix_]rlimit_cpu_time_soft value=2147483647
- procstat_[prefix_]rlimit_file_locks_hard value=2147483647
- procstat_[prefix_]rlimit_file_locks_soft value=2147483647
- procstat_[prefix_]rlimit_memory_data_hard value=2147483647
- procstat_[prefix_]rlimit_memory_data_soft value=2147483647
- procstat_[prefix_]rlimit_memory_locked_hard value=65536
- procstat_[prefix_]rlimit_memory_locked_soft value=65536
- procstat_[prefix_]rlimit_memory_rss_hard value=2147483647
- procstat_[prefix_]rlimit_memory_rss_soft value=2147483647
- procstat_[prefix_]rlimit_memory_stack_hard value=2147483647
- procstat_[prefix_]rlimit_memory_stack_soft value=8388608
- procstat_[prefix_]rlimit_memory_vms_hard value=2147483647
- procstat_[prefix_]rlimit_memory_vms_soft value=2147483647
- procstat_[prefix_]rlimit_nice_priority_hard value=0
- procstat_[prefix_]rlimit_nice_priority_soft value=0
- procstat_[prefix_]rlimit_num_fds_hard value=16384
- procstat_[prefix_]rlimit_num_fds_soft value=16384
- procstat_[prefix_]rlimit_realtime_priority_hard value=0
- procstat_[prefix_]rlimit_realtime_priority_soft value=0
- procstat_[prefix_]rlimit_signals_pending_hard value=1758
- procstat_[prefix_]rlimit_signals_pending_soft value=1758

*NOTE: Due to a limitation in an underlying library Telegraf uses, any resource limit > 2147483647 will be misreported as 2147483647.*
