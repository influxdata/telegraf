# Procstat Input Plugin

The procstat plugin can be used to monitor the system resource usage of one or
more processes.  The procstat_lookup metric displays the query information,
specifically the number of PIDs returned on a search

Processes can be selected for monitoring using one of several methods:

- pidfile
- exe
- pattern
- user
- systemd_unit
- cgroup
- supervisor_unit
- win_service

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
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
  ## Systemd unit name, supports globs when include_systemd_children is set to true
  # systemd_unit = "nginx.service"
  # include_systemd_children = false
  ## CGroup name or path, supports globs
  # cgroup = "systemd/system.slice/nginx.service"
  ## Supervisor service names of hypervisorctl management
  # supervisor_units = ["webserver", "proxy"]

  ## Windows service name
  # win_service = ""

  ## override for process_name
  ## This is optional; default is sourced from /proc/<pid>/status
  # process_name = "bar"

  ## Field name prefix
  # prefix = ""

  ## Mode to use when calculating CPU usage. Can be one of 'solaris' or 'irix'.
  # mode = "irix"

  ## Add the given information tag instead of a field
  ## This allows to create unique metrics/series when collecting processes with
  ## otherwise identical tags. However, please be careful as this can easily
  ## result in a large number of series, especially with short-lived processes,
  ## creating high cardinality at the output.
  ## Available options are:
  ##   cmdline -- full commandline
  ##   pid     -- ID of the process
  ##   ppid    -- ID of the process' parent
  ##   status  -- state of the process
  ##   user    -- username owning the process
  # tag_with = []


  ## Method to use when finding process IDs.  Can be one of 'pgrep', or
  ## 'native'.  The pgrep finder calls the pgrep executable in the PATH while
  ## the native finder performs the search directly in a manor dependent on the
  ## platform.  Default is 'pgrep'
  # pid_finder = "pgrep"
```

### Windows support

Preliminary support for Windows has been added, however you may prefer using
the `win_perf_counters` input plugin as a more mature alternative.

### Darwin specifics

If you use this plugin with `supervisor_units` *and* `pattern` on Darwin, you
**have to** use the `pgrep` finder as the underlying library relies on `pgrep`.

### Permissions

Some files or directories may require elevated permissions. As such a user may
need to provide telegraf with higher levels of permissions to access and produce
metrics.

## Metrics

For descriptions of these tags and fields, consider reading one of the
following:

- [Linux Kernel /proc Filesystem][kernel /proc]
- [proc manpage][manpage]

[kernel /proc]: https://www.kernel.org/doc/html/latest/filesystems/proc.html
[manpage]: https://man7.org/linux/man-pages/man5/proc.5.html

Below are an example set of tags and fields:

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
    - cgroup_full (when cgroup or systemd_unit is used with glob)
    - supervisor_unit (when defined)
    - win_service (when defined)
  - fields:
    - child_major_faults (int)
    - child_minor_faults (int)
    - created_at (int) [epoch in nanoseconds]
    - cpu_time (int)
    - cpu_time_iowait (float) (zero for all OSes except Linux)
    - cpu_time_system (float)
    - cpu_time_user (float)
    - cpu_usage (float)
    - involuntary_context_switches (int)
    - major_faults (int)
    - memory_anonymous (int)
    - memory_private_clean (int)
    - memory_private_dirty (int)
    - memory_pss (int)
    - memory_referenced (int)
    - memory_rss (int)
    - memory_shared_clean (int)
    - memory_shared_dirty (int)
    - memory_size (int)
    - memory_swap (int)
    - memory_usage (float)
    - memory_vms (int)
    - minor_faults (int)
    - nice_priority (int)
    - num_fds (int, *telegraf* may need to be ran as **root**)
    - num_threads (int)
    - pid (int)
    - ppid (int)
    - status (string)
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
    - supervisor_unit
    - win_service
    - result
  - fields:
    - pid_count (int)
    - running (int)
    - result_code (int, success = 0, lookup_error = 1)

*NOTE: Resource limit > 2147483647 will be reported as 2147483647.*

## Example Output

```text
procstat_lookup,host=prash-laptop,pattern=influxd,pid_finder=pgrep,result=success pid_count=1i,running=1i,result_code=0i 1582089700000000000
procstat,host=prash-laptop,pattern=influxd,process_name=influxd,user=root involuntary_context_switches=151496i,child_minor_faults=1061i,child_major_faults=8i,cpu_time_user=2564.81,pid=32025i,major_faults=8609i,created_at=1580107536000000000i,voluntary_context_switches=1058996i,cpu_time_system=616.98,memory_swap=0i,memory_locked=0i,memory_usage=1.7797634601593018,num_threads=18i,cpu_time_iowait=0,memory_rss=148643840i,memory_vms=1435688960i,memory_data=0i,memory_stack=0i,minor_faults=1856550i 1582089700000000000
```
