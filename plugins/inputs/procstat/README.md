# Procstat Input Plugin

The procstat plugin can be used to monitor the system resource usage of one or more processes.
The procstat_lookup metric displays the query information, 
specifically the number of PIDs returned on a search

Processes can be selected for monitoring using one of several methods:
- pidfile
- exe
- pattern
- user
- systemd_unit
- cgroup

### Configuration:

```toml
# Monitor process cpu and memory usage
[[inputs.procstat]]
  ## PID file to monitor process
  pid_file = "/var/run/nginx.pid"
  ## executable name (ie, pgrep <exe>)
  # exe = "nginx"
  ## pattern as argument for pgrep (ie, pgrep -f <pattern>)
  # pattern = "nginx"
  ## user as argument for pgrep (ie, pgrep -u <user>)
  # user = "nginx"
  ## Systemd unit name
  # systemd_unit = "nginx.service"
  ## CGroup name or path
  # cgroup = "systemd/system.slice/nginx.service"

  ## override for process_name
  ## This is optional; default is sourced from /proc/<pid>/status
  # process_name = "bar"

  ## Field name prefix
  # prefix = ""

  ## Add PID as a tag instead of a field; useful to differentiate between
  ## processes whose tags are otherwise the same.  Can create a large number
  ## of series, use judiciously.
  # pid_tag = false

  ## Method to use when finding process IDs.  Can be one of 'pgrep', or
  ## 'native'.  The pgrep finder calls the pgrep executable in the PATH while
  ## the native finder performs the search directly in a manor dependent on the
  ## platform.  Default is 'pgrep'
  # pid_finder = "pgrep"
```

#### Windows support

Preliminary support for Windows has been added, however you may prefer using
the `win_perf_counters` input plugin as a more mature alternative.

When using the `pid_finder = "native"` in Windows, the pattern lookup method is
implemented as a WMI query.  The pattern allows fuzzy matching using only
[WMI query patterns](https://msdn.microsoft.com/en-us/library/aa392263(v=vs.85).aspx):
```toml
[[inputs.procstat]]
  pattern = "%influx%"
  pid_finder = "native"
```

### Metrics:

- procstat
  - tags:
    - pid (when `pid_tag` is true)
    - process_name
    - pidfile (when defined)
    - exe (when defined)
    - pattern (when defined)
    - user (when selected)
    - systemd_unit (when defined)
    - cgroup (when defined)
  - fields:
    - cpu_time (int)
    - cpu_time_guest (float)
    - cpu_time_guest_nice (float)
    - cpu_time_idle (float)
    - cpu_time_iowait (float)
    - cpu_time_irq (float)
    - cpu_time_nice (float)
    - cpu_time_soft_irq (float)
    - cpu_time_steal (float)
    - cpu_time_stolen (float)
    - cpu_time_system (float)
    - cpu_time_user (float)
    - cpu_usage (float)
    - involuntary_context_switches (int)
    - memory_data (int)
    - memory_locked (int)
    - memory_rss (int)
    - memory_stack (int)
    - memory_swap (int)
    - memory_vms (int)
    - nice_priority (int)
    - num_fds (int, *telegraf* may need to be ran as **root**)
    - num_threads (int)
    - pid (int)
    - read_bytes (int, *telegraf* may need to be ran as **root**)
    - read_count (int, *telegraf* may need to be ran as **root**)
    - realtime_priority (int)
    - rlimit_cpu_time_hard (int)
    - rlimit_cpu_time_soft (int)
    - rlimit_file_locks_hard (int)
    - rlimit_file_locks_soft (int)
    - rlimit_memory_data_hard (int)
    - rlimit_memory_data_soft (int)
    - rlimit_memory_locked_hard (int)
    - rlimit_memory_locked_soft (int)
    - rlimit_memory_rss_hard (int)
    - rlimit_memory_rss_soft (int)
    - rlimit_memory_stack_hard (int)
    - rlimit_memory_stack_soft (int)
    - rlimit_memory_vms_hard (int)
    - rlimit_memory_vms_soft (int)
    - rlimit_nice_priority_hard (int)
    - rlimit_nice_priority_soft (int)
    - rlimit_num_fds_hard (int)
    - rlimit_num_fds_soft (int)
    - rlimit_realtime_priority_hard (int)
    - rlimit_realtime_priority_soft (int)
    - rlimit_signals_pending_hard (int)
    - rlimit_signals_pending_soft (int)
    - signals_pending (int)
    - voluntary_context_switches (int)
    - write_bytes (int, *telegraf* may need to be ran as **root**)
    - write_count (int, *telegraf* may need to be ran as **root**)
- procstat_lookup
  - tags:
    - exe (string)
    - pid_finder (string)
    - pid_file (string)
    - pattern (string)
    - prefix (string)
    - user (string)
    - systemd_unit (string)
    - cgroup (string)
  - fields:
    - pid_count (int)
*NOTE: Resource limit > 2147483647 will be reported as 2147483647.*

### Example Output:

```
procstat,pidfile=/var/run/lxc/dnsmasq.pid,process_name=dnsmasq rlimit_file_locks_soft=2147483647i,rlimit_signals_pending_hard=1758i,voluntary_context_switches=478i,read_bytes=307200i,cpu_time_user=0.01,cpu_time_guest=0,memory_swap=0i,memory_locked=0i,rlimit_num_fds_hard=4096i,rlimit_nice_priority_hard=0i,num_fds=11i,involuntary_context_switches=20i,read_count=23i,memory_rss=1388544i,rlimit_memory_rss_soft=2147483647i,rlimit_memory_rss_hard=2147483647i,nice_priority=20i,rlimit_cpu_time_hard=2147483647i,cpu_time=0i,write_bytes=0i,cpu_time_idle=0,cpu_time_nice=0,memory_data=229376i,memory_stack=135168i,rlimit_cpu_time_soft=2147483647i,rlimit_memory_data_hard=2147483647i,rlimit_memory_locked_hard=65536i,rlimit_signals_pending_soft=1758i,write_count=11i,cpu_time_iowait=0,cpu_time_steal=0,cpu_time_stolen=0,rlimit_memory_stack_soft=8388608i,cpu_time_system=0.02,cpu_time_guest_nice=0,rlimit_memory_locked_soft=65536i,rlimit_memory_vms_soft=2147483647i,rlimit_file_locks_hard=2147483647i,rlimit_realtime_priority_hard=0i,pid=828i,num_threads=1i,cpu_time_soft_irq=0,rlimit_memory_vms_hard=2147483647i,rlimit_realtime_priority_soft=0i,memory_vms=15884288i,rlimit_memory_stack_hard=2147483647i,cpu_time_irq=0,rlimit_memory_data_soft=2147483647i,rlimit_num_fds_soft=1024i,signals_pending=0i,rlimit_nice_priority_soft=0i,realtime_priority=0i
procstat,exe=influxd,process_name=influxd rlimit_num_fds_hard=16384i,rlimit_signals_pending_hard=1758i,realtime_priority=0i,rlimit_memory_vms_hard=2147483647i,rlimit_signals_pending_soft=1758i,cpu_time_stolen=0,rlimit_memory_stack_hard=2147483647i,rlimit_realtime_priority_hard=0i,cpu_time=0i,pid=500i,voluntary_context_switches=975i,cpu_time_idle=0,memory_rss=3072000i,memory_locked=0i,rlimit_nice_priority_soft=0i,signals_pending=0i,nice_priority=20i,read_bytes=823296i,cpu_time_soft_irq=0,rlimit_memory_data_hard=2147483647i,rlimit_memory_locked_soft=65536i,write_count=8i,cpu_time_irq=0,memory_vms=33501184i,rlimit_memory_stack_soft=8388608i,cpu_time_iowait=0,rlimit_memory_vms_soft=2147483647i,rlimit_nice_priority_hard=0i,num_fds=29i,memory_data=229376i,rlimit_cpu_time_soft=2147483647i,rlimit_file_locks_soft=2147483647i,num_threads=1i,write_bytes=0i,cpu_time_steal=0,rlimit_memory_rss_hard=2147483647i,cpu_time_guest=0,cpu_time_guest_nice=0,cpu_usage=0,rlimit_memory_locked_hard=65536i,rlimit_file_locks_hard=2147483647i,involuntary_context_switches=38i,read_count=16851i,memory_swap=0i,rlimit_memory_data_soft=2147483647i,cpu_time_user=0.11,rlimit_cpu_time_hard=2147483647i,rlimit_num_fds_soft=16384i,rlimit_realtime_priority_soft=0i,cpu_time_system=0.27,cpu_time_nice=0,memory_stack=135168i,rlimit_memory_rss_soft=2147483647i
```
