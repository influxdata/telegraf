# Intel RDT Input Plugin

This plugin collects information provided by monitoring features of the
[Intel Resource Director Technology][rdt], a hardware framework to monitor and
control the utilization of shared resources (e.g. last level cache,
memory bandwidth).

Intel’s Resource Director Technology (RDT) framework consists of:

- Cache Monitoring Technology (CMT)
- Memory Bandwidth Monitoring (MBM)
- Cache Allocation Technology (CAT)
- Code and Data Prioritization (CDP)

As multithreaded and multicore platform architectures emerge, the last level
cache and memory bandwidth are key resources to manage for running workloads in
single-threaded, multithreaded, or complex virtual machine environments. Intel
introduces CMT, MBM, CAT and CDP to manage these workloads across shared
resources.

⭐ Telegraf v1.16.0
🏷️ hardware, system
💻 linux, freebsd, macos

[rdt]: https://www.intel.com/content/www/us/en/architecture-and-technology/resource-director-technology.html

## Requirements

The plugin requires the `pqos` cli tool in version 4.0+ to be installed and
configured to work in `OS Interface` mode. The tool is part of the
[Intel(R) RDT Software Package][cmt_cat].

> [!IMPORTANT]
> The `pqos` binary needs to run as root. If telegraf is not running as root
> you need to enable sudo for `pqos` and set the `use_sudo` option to `true`.

To setup `pqos` correctly check the [installation guide][install]. For help on
how to configure the tool visit the [wiki][wiki] and read the
[resource control documentation][resctl]

[cmt_cat]: https://github.com/intel/intel-cmt-cat
[install]: https://github.com/intel/intel-cmt-cat/blob/master/INSTALL
[wiki]: https://github.com/intel/intel-cmt-cat/wiki
[resctl]: https://github.com/intel/intel-cmt-cat/wiki/resctrl

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read Intel RDT metrics
# This plugin ONLY supports non-Windows
[[inputs.intel_rdt]]
  ## Optionally set sampling interval to Nx100ms.
  ## This value is propagated to pqos tool. Interval format is defined by pqos itself.
  ## If not provided or provided 0, will be set to 10 = 10x100ms = 1s.
  # sampling_interval = "10"

  ## Optionally specify the path to pqos executable.
  ## If not provided, auto discovery will be performed.
  # pqos_path = "/usr/local/bin/pqos"

  ## Optionally specify if IPC and LLC_Misses metrics shouldn't be propagated.
  ## If not provided, default value is false.
  # shortened_metrics = false

  ## Specify the list of groups of CPU core(s) to be provided as pqos input.
  ## Mandatory if processes aren't set and forbidden if processes are specified.
  ## e.g. ["0-3", "4,5,6"] or ["1-3,4"]
  # cores = ["0-3"]

  ## Specify the list of processes for which Metrics will be collected.
  ## Mandatory if cores aren't set and forbidden if cores are specified.
  ## e.g. ["qemu", "pmd"]
  # processes = ["process"]

  ## Specify if the pqos process should be called with sudo.
  ## Mandatory if the telegraf process does not run as root.
  # use_sudo = false
```

## Troubleshooting

Pointing to non-existing cores will lead to throwing an error by _pqos_ and the
plugin will not work properly. Be sure to check provided core number exists
within desired system.

Be aware, reading Intel RDT metrics by _pqos_ cannot be done simultaneously on
the same resource.  Do not use any other _pqos_ instance that is monitoring the
same cores or PIDs within the working system.  It is not possible to monitor
same cores or PIDs on different groups.

PIDs associated for the given process could be manually checked by `pidof`
command. E.g:

```sh
pidof PROCESS
```

where `PROCESS` is process name.

## Metrics

| Name          | Full name                                     | Description |
|---------------|-----------------------------------------------|-------------|
| MBL           | Memory Bandwidth on Local NUMA Node  |     Memory bandwidth utilization by the relevant CPU core/process on the local NUMA memory channel        |
| MBR           | Memory Bandwidth on Remote NUMA Node |     Memory bandwidth utilization by the relevant CPU core/process on the remote NUMA memory channel        |
| MBT           | Total Memory Bandwidth               |     Total memory bandwidth utilized by a CPU core/process on local and remote NUMA memory channels        |
| LLC           | L3 Cache Occupancy                   |     Total Last Level Cache occupancy by a CPU core/process         |
| LLC_Misses*    | L3 Cache Misses                      |    Total Last Level Cache misses by a CPU core/process       |
| IPC*           | Instructions Per Cycle               |     Total instructions per cycle executed by a CPU core/process        |

*optional

## Example Output

```text
rdt_metric,cores=12\,19,host=r2-compute-20,name=IPC,process=top value=0 1598962030000000000
rdt_metric,cores=12\,19,host=r2-compute-20,name=LLC_Misses,process=top value=0 1598962030000000000
rdt_metric,cores=12\,19,host=r2-compute-20,name=LLC,process=top value=0 1598962030000000000
rdt_metric,cores=12\,19,host=r2-compute-20,name=MBL,process=top value=0 1598962030000000000
rdt_metric,cores=12\,19,host=r2-compute-20,name=MBR,process=top value=0 1598962030000000000
rdt_metric,cores=12\,19,host=r2-compute-20,name=MBT,process=top value=0 1598962030000000000
```
