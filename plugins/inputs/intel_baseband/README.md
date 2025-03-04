# Intel Baseband Accelerator Input Plugin

This plugin collects metrics from both dedicated and integrated Intel devices
providing Wireless Baseband hardware acceleration. These devices play a key role
in accelerating 5G and 4G Virtualized Radio Access Networks (vRAN) workloads,
increasing the overall compute capacity of commercial, off-the-shelf platforms
by integrating e.g.

- Forward Error Correction (FEC) processing,
- 4G Turbo FEC processing,
- 5G Low Density Parity Check (LDPC)
- Fast Fourier Transform (FFT) block providing DFT/iDFT processing offload for
  the 5G Sounding Reference Signal (SRS)

⭐ Telegraf v1.27.0
🏷️ hardware, network, system
💻 linux

## Requirements

- supported Intel Baseband device installed and configured
- Linux kernel 5.7+
- [pf-bb-config](https://github.com/intel/pf-bb-config) (version >= v23.03)
  installed and running

This plugin supports the following hardware:

- Intel® vRAN Boost integrated accelerators:
  - 4th Gen Intel® Xeon® Scalable processor with Intel® vRAN Boost
    (also known as Sapphire Rapids Edge Enhanced / SPR-EE)
- External expansion cards connected to the PCI bus:
  - Intel® vRAN Dedicated Accelerator ACC100 SoC (code named Mount Bryce)

For more information regarding system configuration, please follow DPDK
installation guides:

- [Intel® vRAN Boost Poll Mode Driver (PMD)][VRB1]
- [Intel® ACC100 5G/4G FEC Poll Mode Drivers][ACC100]

[VRB1]: https://doc.dpdk.org/guides/bbdevs/vrb1.html#installation
[ACC100]: https://doc.dpdk.org/guides/bbdevs/acc100.html#installation

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Intel Baseband Accelerator Input Plugin collects metrics from both dedicated and integrated
# Intel devices that provide Wireless Baseband hardware acceleration.
# This plugin ONLY supports Linux.
[[inputs.intel_baseband]]
  ## Path to socket exposed by pf-bb-config for CLI interaction (mandatory).
  ## In version v23.03 of pf-bb-config the path is created according to the schema:
  ##   "/tmp/pf_bb_config.0000\:<b>\:<d>.<f>.sock" where 0000\:<b>\:<d>.<f> is the PCI device ID.
  socket_path = ""

  ## Path to log file exposed by pf-bb-config with telemetry to read (mandatory).
  ## In version v23.03 of pf-bb-config the path is created according to the schema:
  ##   "/var/log/pf_bb_cfg_0000\:<b>\:<d>.<f>.log" where 0000\:<b>\:<d>.<f> is the PCI device ID.
  log_file_path = ""

  ## Specifies plugin behavior regarding unreachable socket (which might not have been initialized yet).
  ## Available choices:
  ##   - error: Telegraf will return an error on startup if socket is unreachable
  ##   - ignore: Telegraf will ignore error regarding unreachable socket on both startup and gather
  # unreachable_socket_behavior = "error"

  ## Duration that defines how long the connected socket client will wait for
  ## a response before terminating connection.
  ## Since it's local socket access to a fast packet processing application, the timeout should
  ## be sufficient for most users.
  ## Setting the value to 0 disables the timeout (not recommended).
  # socket_access_timeout = "1s"

  ## Duration that defines maximum time plugin will wait for pf-bb-config to write telemetry to the log file.
  ## Timeout may differ depending on the environment.
  ## Must be equal or larger than 50ms.
  # wait_for_telemetry_timeout = "1s"
```

## Metrics

Depending on version of Intel Baseband device and version of pf-bb-config,
subset of following measurements may be exposed:

**The following tags and fields are supported by Intel Baseband plugin:**

| Tag         | Description                                                 |
|-------------|-------------------------------------------------------------|
| `metric`    | Type of metric : "code_blocks", "data_bytes", "per_engine". |
| `operation` | Type of operation: "5GUL", "5GDL", "4GUL", "4GDL", "FFT".   |
| `vf`        | Virtual Function number.                                    |
| `engine`    | Engine number.                                              |

| Metric name (field)  | Description                                                       |
|----------------------|-------------------------------------------------------------------|
| `value`              | Metric value for a given operation (non-negative integer, gauge). |

## Example Output

```text
intel_baseband,host=ubuntu,metric=code_blocks,operation=5GUL,vf=0 value=54i 1685695885000000000
intel_baseband,host=ubuntu,metric=code_blocks,operation=5GDL,vf=0 value=0i 1685695885000000000
intel_baseband,host=ubuntu,metric=code_blocks,operation=FFT,vf=0 value=0i 1685695885000000000
intel_baseband,host=ubuntu,metric=code_blocks,operation=5GUL,vf=1 value=0i 1685695885000000000
intel_baseband,host=ubuntu,metric=code_blocks,operation=5GDL,vf=1 value=32i 1685695885000000000
intel_baseband,host=ubuntu,metric=code_blocks,operation=FFT,vf=1 value=0i 1685695885000000000
intel_baseband,host=ubuntu,metric=data_bytes,operation=5GUL,vf=0 value=18560i 1685695885000000000
intel_baseband,host=ubuntu,metric=data_bytes,operation=5GDL,vf=0 value=0i 1685695885000000000
intel_baseband,host=ubuntu,metric=data_bytes,operation=FFT,vf=0 value=0i 1685695885000000000
intel_baseband,host=ubuntu,metric=data_bytes,operation=5GUL,vf=1 value=0i 1685695885000000000
intel_baseband,host=ubuntu,metric=data_bytes,operation=5GDL,vf=1 value=86368i 1685695885000000000
intel_baseband,host=ubuntu,metric=data_bytes,operation=FFT,vf=1 value=0i 1685695885000000000
intel_baseband,engine=0,host=ubuntu,metric=per_engine,operation=5GUL value=72i 1685695885000000000
intel_baseband,engine=1,host=ubuntu,metric=per_engine,operation=5GUL value=72i 1685695885000000000
intel_baseband,engine=2,host=ubuntu,metric=per_engine,operation=5GUL value=72i 1685695885000000000
intel_baseband,engine=3,host=ubuntu,metric=per_engine,operation=5GUL value=72i 1685695885000000000
intel_baseband,engine=4,host=ubuntu,metric=per_engine,operation=5GUL value=72i 1685695885000000000
intel_baseband,engine=0,host=ubuntu,metric=per_engine,operation=5GDL value=132i 1685695885000000000
intel_baseband,engine=1,host=ubuntu,metric=per_engine,operation=5GDL value=130i 1685695885000000000
intel_baseband,engine=0,host=ubuntu,metric=per_engine,operation=FFT value=0i 1685695885000000000
```
