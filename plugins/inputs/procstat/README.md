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
- win_service

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

  ## Windows service name
  # win_service = ""

  ## override for process_name
  ## This is optional; default is sourced from /proc/<pid>/status
  # process_name = "bar"

  ## Field name prefix
  # prefix = ""

  ## When true add the full cmdline as a tag.
  # cmdline_tag = false

  ## Mode to use when calculating CPU usage. Can be one of 'solaris' or 'irix'.
  # mode = "irix"

  ## Add the PID as a tag instead of as a field.  When collecting multiple
  ## processes with otherwise matching tags this setting should be enabled to
  ## ensure each process has a unique identity.
  ##
  ## Enabling this option may result in a large number of series, especially
  ## when processes have a short lifetime.
  # pid_tag = false

  ## Method to use when finding process IDs.  Can be one of 'pgrep', or
  ## 'native'.  The pgrep finder calls the pgrep executable in the PATH while
  ## the native finder performs the search directly in a manor dependent on the
  ## platform.  Default is 'pgrep'
  # pid_finder = "pgrep"

  ## Select wich extra metrics should be added:
  ##  - "threads": to enable collection of number of file descriptors
  ##  - "fds": to enable collection of context switches
  ##  - "ctx_switches": to enable collection of page faults
  ##  - "page_faults": to enable collection of IO
  ##  - "io": to enable collection of proc creation time
  ##  - "create_time": to enable collection of CPU time used
  ##  - "cpu": to enable collection of percentage of CPU used
  ##  - "cpu_percent": to enable collection of memory used
  ##  - "mem": to enable collection of memory percentage used
  ##  - "mem_percent": to enable collection of procs' limits
  ##  - "limits": to enable collection of procs' limits
  ##  - "tcp_stats": tcp_* and upd_socket metrics
  ##  - "connections_endpoints": new metric procstat_tcp with connections and listeners endpoints
  ## Default value:
  # metrics_include = [
  #  "threads",
  #  "fds",
  #  "ctx_switches",
  #  "page_faults",
  #  "io",
  #  "create_time",
  #  "cpu",
  #  "cpu_percent",
  #  "mem",
  #  "mem_percent",
  #  "limits",
  # ]
```

#### Windows support

Preliminary support for Windows has been added, however you may prefer using
the `win_perf_counters` input plugin as a more mature alternative.

### Metrics:

- procstat
  - tags:
    - pid (when `pid_tag` is true)
    - cmdline (when 'cmdline_tag' is true)
    - process_name
    - pidfile (when defined)
    - exe (when defined)
    - pattern (when defined)
    - user (when selected)
    - systemd_unit (when defined)
    - cgroup (when defined)
    - win_service (when defined)
  - fields:
    - child_major_faults (int)
    - child_minor_faults (int)
    - created_at (int) [epoch in nanoseconds]
    - cpu_time (int)
    - cpu_time_guest (float)
    - cpu_time_guest_nice (float)
    - cpu_time_idle (float)
    - cpu_time_iowait (float)
    - cpu_time_irq (float)
    - cpu_time_nice (float)
    - cpu_time_soft_irq (float)
    - cpu_time_steal (float)
    - cpu_time_system (float)
    - cpu_time_user (float)
    - cpu_usage (float)
    - involuntary_context_switches (int)
    - major_faults (int)
    - memory_data (int)
    - memory_locked (int)
    - memory_rss (int)
    - memory_stack (int)
    - memory_swap (int)
    - memory_usage (float)
    - memory_vms (int)
    - minor_faults (int)
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
    - exe
    - pid_finder
    - pid_file
    - pattern
    - prefix
    - user
    - systemd_unit
    - cgroup
    - win_service
    - result
  - fields:
    - pid_count (int)
    - running (int)
    - result_code (int, success = 0, lookup_error = 1)

If ``connections_stats`` enabled, added fields:
- procstat
  - fields:
    - tcp_close (int)
    - tcp_close_wait (int)
    - tcp_closing (int)
    - tcp_established (int)
    - tcp_fin_wait1 (int)
    - tcp_fin_wait2 (int)
    - tcp_last_ack (int)
    - tcp_listen (int)
    - tcp_none (int)
    - tcp_syn_recv (int)
    - tcp_syn_sent (int)

If ``connections_endpoints`` enabled, added fields:
- procstat_tcp
  - tags:
    - pid (when `pid_tag` is true)
    - cmdline (when 'cmdline_tag' is true)
    - process_name
    - pidfile (when defined)
    - exe (when defined)
    - pattern (when defined)
    - user (when selected)
    - systemd_unit (when defined)
    - cgroup (when defined)
  - fields:
    - conn (string)
    - listen (string)

To gather connection info, if Telegraf is not run as root, it needs the following capabilities
```
sudo setcap "CAP_DAC_READ_SEARCH,CAP_SYS_PTRACE+ep" telegraf
```

*NOTE: Resource limit > 2147483647 will be reported as 2147483647.*

### Example Output:

```
procstat_lookup,host=prash-laptop,pattern=influxd,pid_finder=pgrep,result=success pid_count=1i,running=1i,result_code=0i 1582089700000000000
procstat,host=prash-laptop,pattern=influxd,process_name=influxd,user=root involuntary_context_switches=151496i,child_minor_faults=1061i,child_major_faults=8i,cpu_time_user=2564.81,cpu_time_idle=0,cpu_time_irq=0,cpu_time_guest=0,pid=32025i,major_faults=8609i,created_at=1580107536000000000i,voluntary_context_switches=1058996i,cpu_time_system=616.98,cpu_time_steal=0,cpu_time_guest_nice=0,memory_swap=0i,memory_locked=0i,memory_usage=1.7797634601593018,num_threads=18i,cpu_time_nice=0,cpu_time_iowait=0,cpu_time_soft_irq=0,memory_rss=148643840i,memory_vms=1435688960i,memory_data=0i,memory_stack=0i,minor_faults=1856550i 1582089700000000000
procstat,host=laptop,pattern=httpd,process_name=httpd,user=root child_major_faults=0i,child_minor_faults=70i,cpu_time=0i,cpu_time_guest=0,cpu_time_guest_nice=0,cpu_time_idle=0,cpu_time_iowait=0,cpu_time_irq=0,cpu_time_nice=0,cpu_time_soft_irq=0,cpu_time_steal=0,cpu_time_system=0.01,cpu_time_user=0.02,cpu_usage=0,created_at=1611738400000000000i,involuntary_context_switches=15i,listen=1i,major_faults=0i,memory_data=999424i,memory_locked=0i,memory_rss=4677632i,memory_stack=135168i,memory_swap=0i,memory_usage=0.013990458101034164,memory_vms=6078464i,minor_faults=1636i,nice_priority=20i,num_fds=8i,num_threads=1i,pid=1738811i,read_bytes=0i,read_count=4397i,realtime_priority=0i,rlimit_cpu_time_hard=2147483647i,rlimit_cpu_time_soft=2147483647i,rlimit_file_locks_hard=2147483647i,rlimit_file_locks_soft=2147483647i,rlimit_memory_data_hard=2147483647i,rlimit_memory_data_soft=2147483647i,rlimit_memory_locked_hard=65536i,rlimit_memory_locked_soft=65536i,rlimit_memory_rss_hard=2147483647i,rlimit_memory_rss_soft=2147483647i,rlimit_memory_stack_hard=2147483647i,rlimit_memory_stack_soft=8388608i,rlimit_memory_vms_hard=2147483647i,rlimit_memory_vms_soft=2147483647i,rlimit_nice_priority_hard=0i,rlimit_nice_priority_soft=0i,rlimit_num_fds_hard=1048576i,rlimit_num_fds_soft=1048576i,rlimit_realtime_priority_hard=0i,rlimit_realtime_priority_soft=0i,rlimit_signals_pending_hard=127473i,rlimit_signals_pending_soft=127473i,signals_pending=0i,tcp_close=0i,tcp_close_wait=0i,tcp_closing=0i,tcp_established=0i,tcp_fin_wait1=0i,tcp_fin_wait2=0i,tcp_last_ack=0i,tcp_listen=1i,tcp_syn_recv=0i,tcp_syn_sent=0i,voluntary_context_switches=169i,write_bytes=53248i,write_count=10i 1611738522000000000
procstat_tcp,host=laptop,pattern=httpd,process_name=httpd,user=root conn="",listen="192.168.1.35:80,192.168.1.48:80,[da01:beef:234:3830:aeda:f001:a00c:0091]:80,[aa01:beef:234:3830:e8e:0000:000a:6b0f]:80" 1611738522000000000
```
