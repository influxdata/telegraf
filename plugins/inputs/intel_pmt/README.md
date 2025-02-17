# Intel® Platform Monitoring Technology Input Plugin

This plugin collects metrics via the Linux kernel driver for
Intel® Platform Monitoring Technology (Intel® PMT), an architecture capable of
enumerating and accessing hardware monitoring capabilities on supported devices.

⭐ Telegraf v1.28.0
🏷️ hardware, system
💻 linux

## Requirements

- supported device
- Linux kernel >= 5.11
- `intel_pmt_telemetry` module loaded (on kernels 5.11-5.14)
- `intel_pmt` module loaded (on kernels 5.14+)

This plugin supports devices exposing PMT, e.g.

- 4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids / SPR)
- 6th Generation Intel® Xeon® Scalable Processors (Granite Rapids / GNR)

Support has been added to the mainline Linux kernel under the platform driver
(`drivers/platform/x86/intel/pmt`) which exposes the Intel PMT telemetry space
as a sysfs entry at `/sys/class/intel_pmt/`. Each discovered telemetry
aggregator is exposed as a directory (with a `telem` prefix) containing a `guid`
identifying the unique PMT space. This file is associated with a set of XML
specification files which can be found in the [Intel-PMT Repository][repo].
The XML specification must be specified as an absolute path to the `pmt.xml`
file using the `spec` setting .

This plugin discovers and parses the telemetry data exposed by the kernel driver
using the specification inside the XML files. Furthermore, the plugin then reads
low level samples/counters and evaluates high level samples/counters according
to transformation formulas, and then reports the collected values.

> [!IMPORTANT]
> PMT space is located in `/sys/class/intel_pmt` with `telem` files requiring
> **root privileges** to be read. If Telegraf is not running as root you should
> add the following capability to the Telegraf executable:
>
> ```sh
> sudo setcap cap_dac_read_search+ep /usr/bin/telegraf
> ```

[repo]: https://github.com/intel/Intel-PMT

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Intel Platform Monitoring Technology plugin exposes Intel PMT metrics available through the Intel PMT kernel space.
# This plugin ONLY supports Linux.
[[inputs.intel_pmt]]
  ## Filepath to PMT XML within local copies of XML files from PMT repository.
  ## The filepath should be absolute.
  spec = "/home/telegraf/Intel-PMT/xml/pmt.xml"
  
  ## Enable metrics by their datatype.
  ## See the Enabling Metrics section in README for more details.
  ## If empty, all metrics are enabled.
  ## When used, the alternative option samples_enabled should NOT be used.
  # datatypes_enabled = []
  
  ## Enable metrics by their name.
  ## See the Enabling Metrics section in README for more details.
  ## If empty, all metrics are enabled.
  ## When used, the alternative option datatypes_enabled should NOT be used.
  # samples_enabled = []
```

### Enabling metrics

By default, the plugin collects all available metrics.

To limit the metrics collected by the plugin,
two options are available for selecting metrics:

- enable by datatype (groups of metrics),
- enable by name.

It's important to note that only one enabling option
should be chosen at a time.

See the table below for available datatypes and related metrics:

| Datatype                | Metric name             | Description                                                                                                                 |
|-------------------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| `txtal_strap`           | `XTAL_FREQ`             | Clock rate of the crystal oscillator on this silicon                                                                        |
| `tdram_energy`          | `DRAM_ENERGY_LOW`       | DRAM energy consumed by all DIMMS in all Channels (uJ)                                                                      |
|                         | `DRAM_ENERGY_HIGH`      | DRAM energy consumed by all DIMMS in all Channels (uJ)                                                                      |
| `tbandwidth_32b`        | `C2U_BW`                | Core to Uncore Bandwidth (per core and per uncore)                                                                          |
|                         | `U2C_BW`                | Uncore to Core Bandwidth (per core and per uncore)                                                                          |
|                         | `PC2_LOW`               | Time spent in the Package C-State 2 (PC2)                                                                                   |
|                         | `PC2_HIGH`              | Time spent in the Package C-State 2 (PC2)                                                                                   |
|                         | `PC6_LOW`               | Time spent in the Package C-State 6 (PC6)                                                                                   |
|                         | `PC6_HIGH`              | Time spent in the Package C-State 6 (PC6)                                                                                   |
|                         | `MEM_RD_BW`             | Memory Read Bandwidth (per channel)                                                                                         |
|                         | `MEM_WR_BW`             | Memory Write Bandwidth (per channel)                                                                                        |
|                         | `DDRT_READ_BW`          | DDRT Read Bandwidth (per channel)                                                                                           |
|                         | `DDRT_WR_BW`            | DDRT Write Bandwidth (per channel)                                                                                          |
|                         | `THRT_COUNT`            | Number of clock ticks when throttling occurred on IMC channel (per channel)                                                 |
|                         | `PMSUM`                 | Energy accumulated by IMC channel (per channel)                                                                             |
|                         | `CMD_CNT_CH0`           | Command count for IMC channel subchannel 0 (per channel)                                                                    |
|                         | `CMD_CNT_CH1`           | Command count for IMC channel subchannel 1 (per channel)                                                                    |
| `tU32.0`                | `PEM_ANY`               | Duration for which a core frequency excursion occurred due to a listed or unlisted reason                                   |
|                         | `PEM_THERMAL`           | Duration for which a core frequency excursion occurred due to EMTTM                                                         |
|                         | `PEM_EXT_PROCHOT`       | Duration for which a core frequency excursion occurred due to an external PROCHOT assertion                                 |
|                         | `PEM_PBM`               | Duration for which a core frequency excursion occurred due to PBM                                                           |
|                         | `PEM_PL1`               | Duration for which a core frequency excursion occurred due to PL1                                                           |
|                         | `PEM_RESERVED`          | PEM Reserved Counter                                                                                                        |
|                         | `PEM_PL2`               | Duration for which a core frequency excursion occurred due to PL2                                                           |
|                         | `PEM_PMAX`              | Duration for which a core frequency excursion occurred due to PMAX                                                          |
| `tbandwidth_28b`        | `C0Residency`           | Core C0 Residency (per core)                                                                                                |
|                         | `C1Residency`           | Core C1 Residency (per core)                                                                                                |
| `tratio`                | `FET`                   | Current Frequency Excursion Threshold. Ratio of the core frequency.                                                         |
| `tbandwidth_24b`        | `UFS_MAX_RING_TRAFFIC`  | IO Bandwidth for DMI or PCIE port (per port)                                                                                |
| `ttemperature`          | `TEMP`                  | Current temperature of a core (per core)                                                                                    |
| `tU8.0`                 | `VERSION`               | For SPR, it's 0. New feature versions will uprev this.                                                                      |
| `tebb_energy`           | `FIVR_HBM_ENERGY`       | FIVR HBM Energy in uJ (per HBM)                                                                                             |
| `tBOOL`                 | `OOB_PEM_ENABLE`        | 0x0 (Default)=Inband interface for PEM is enabled. 0x1=OOB interface for PEM is enabled.                                    |
|                         | `ENABLE_PEM`            | 0 (Default): Disable PEM. 1: Enable PEM                                                                                     |
|                         | `ANY`                   | Set if a core frequency excursion occurs due to a listed or unlisted reason                                                 |
|                         | `THERMAL`               | Set if a core frequency excursion occurs due to any thermal event in core/uncore                                            |
|                         | `EXT_PROCHOT`           | Set if a core frequency excursion occurs due to external PROCHOT assertion                                                  |
|                         | `PBM`                   | Set if a core frequency excursion occurs due to a power limit (socket RAPL and/or platform RAPL)                            |
|                         | `PL1`                   | Set if a core frequency excursion occurs due to PL1 input from any interfaces                                               |
|                         | `PL2`                   | Set if a core frequency excursion occurs due to PL2 input from any interfaces                                               |
|                         | `PMAX`                  | Set if a core frequency excursion occurs due to PMAX                                                                        |
| `ttsc`                  | `ART`                   | TSC Delta HBM (per HBM)                                                                                                     |
| `tproduct_id`           | `PRODUCT_ID`            | Product ID                                                                                                                  |
| `tstring`               | `LOCAL_REVISION`        | Local Revision ID for this product                                                                                          |
|                         | `RECORD_TYPE`           | Record Type                                                                                                                 |
| `tcore_state`           | `EN`                    | Core x is enabled (per core)                                                                                                |
| `thist_counter`         | `FREQ_HIST_R0`          | Frequency histogram range 0 (core in C6) counter (per core)                                                                 |
|                         | `FREQ_HIST_R1`          | Frequency histogram range 1 (16.67-800 MHz) counter (per core)                                                              |
|                         | `FREQ_HIST_R2`          | Frequency histogram range 2 (801-1200 MHz) counter (per core)                                                               |
|                         | `FREQ_HIST_R3`          | Frequency histogram range 3 (1201-1600 MHz) counter (per core)                                                              |
|                         | `FREQ_HIST_R4`          | Frequency histogram range 4 (1601-2000 MHz) counter (per core)                                                              |
|                         | `FREQ_HIST_R5`          | Frequency histogram range 5 (2001-2400 MHz) counter (per core)                                                              |
|                         | `FREQ_HIST_R6`          | Frequency histogram range 6 (2401-2800 MHz) counter (per core)                                                              |
|                         | `FREQ_HIST_R7`          | Frequency histogram range 7 (2801-3200 MHz) counter (per core)                                                              |
|                         | `FREQ_HIST_R8`          | Frequency histogram range 8 (3201-3600 MHz) counter (per core)                                                              |
|                         | `FREQ_HIST_R9`          | Frequency histogram range 9 (3601-4000 MHz) counter (per core)                                                              |
|                         | `FREQ_HIST_R10`         | Frequency histogram range 10 (4001-4400 MHz) counter (per core)                                                             |
|                         | `FREQ_HIST_R11`         | Frequency histogram range 11 (greater then 4400 MHz) (per core)                                                             |
|                         | `VOLT_HIST_R0`          | Voltage histogram range 0 (less then 602 mV) counter (per core)                                                             |
|                         | `VOLT_HIST_R1`          | Voltage histogram range 1 (602.5-657 mV) counter (per core)                                                                 |
|                         | `VOLT_HIST_R2`          | Voltage histogram range 2 (657.5-712 mV) counter (per core)                                                                 |
|                         | `VOLT_HIST_R3`          | Voltage histogram range 3 (712.5-767 mV) counter (per core)                                                                 |
|                         | `VOLT_HIST_R4`          | Voltage histogram range 4 (767.5-822 mV) counter (per core)                                                                 |
|                         | `VOLT_HIST_R5`          | Voltage histogram range 5 (822.5-877 mV) counter (per core)                                                                 |
|                         | `VOLT_HIST_R6`          | Voltage histogram range 6 (877.5-932 mV) counter (per core)                                                                 |
|                         | `VOLT_HIST_R7`          | Voltage histogram range 7 (932.5-987 mV) counter (per core)                                                                 |
|                         | `VOLT_HIST_R8`          | Voltage histogram range 8 (987.5-1042 mV) counter (per core)                                                                |
|                         | `VOLT_HIST_R9`          | Voltage histogram range 9 (1042.5-1097 mV) counter (per core)                                                               |
|                         | `VOLT_HIST_R10`         | Voltage histogram range 10 (1097.5-1152 mV) counter (per core)                                                              |
|                         | `VOLT_HIST_R11`         | Voltage histogram range 11 (greater then 1152 mV) counter (per core)                                                        |
|                         | `TEMP_HIST_R0`          | Temperature histogram range 0 (less then 20°C) counter                                                                      |
|                         | `TEMP_HIST_R1`          | Temperature histogram range 1 (20.5-27.5°C) counter                                                                         |
|                         | `TEMP_HIST_R2`          | Temperature histogram range 2 (28-35°C) counter                                                                             |
|                         | `TEMP_HIST_R3`          | Temperature histogram range 3 (35.5-42.5°C) counter                                                                         |
|                         | `TEMP_HIST_R4`          | Temperature histogram range 4 (43-50°C) counter                                                                             |
|                         | `TEMP_HIST_R5`          | Temperature histogram range 5 (50.5-57.5°C) counter                                                                         |
|                         | `TEMP_HIST_R6`          | Temperature histogram range 6 (58-65°C) counter                                                                             |
|                         | `TEMP_HIST_R7`          | Temperature histogram range 7 (65.5-72.5°C) counter                                                                         |
|                         | `TEMP_HIST_R8`          | Temperature histogram range 8 (73-80°C) counter                                                                             |
|                         | `TEMP_HIST_R9`          | Temperature histogram range 9 (80.5-87.5°C) counter                                                                         |
|                         | `TEMP_HIST_R10`         | Temperature histogram range 10 (88-95°C) counter                                                                            |
|                         | `TEMP_HIST_R11`         | Temperature histogram range 11 (greater then 95°C) counter                                                                  |
| `tpvp_throttle_counter` | `PVP_THROTTLE_64`       | Counter indicating the number of times the core x was throttled in the last 64 cycles window                                |
|                         | `PVP_THROTTLE_1024`     | Counter indicating the number of times the core x was throttled in the last 1024 cycles window                              |
| `tpvp_level_res`        | `PVP_LEVEL_RES_128_L0`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 0 of this type of CPU instruction |
|                         | `PVP_LEVEL_RES_128_L1`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 1 of this type of CPU instruction |
|                         | `PVP_LEVEL_RES_128_L2`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 2 of this type of CPU instruction |
|                         | `PVP_LEVEL_RES_128_L3`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 3 of this type of CPU instruction |
|                         | `PVP_LEVEL_RES_256_L0`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 0 of AVX256 CPU instructions      |
|                         | `PVP_LEVEL_RES_256_L1`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 1 of AVX256 CPU instructions      |
|                         | `PVP_LEVEL_RES_256_L2`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 2 of AVX256 CPU instructions      |
|                         | `PVP_LEVEL_RES_256_L3`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 3 of AVX256 CPU instructions      |
|                         | `PVP_LEVEL_RES_512_L0`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 0 of AVX512 CPU instructions      |
|                         | `PVP_LEVEL_RES_512_L1`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 1 of AVX512 CPU instructions      |
|                         | `PVP_LEVEL_RES_512_L2`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 2 of AVX512 CPU instructions      |
|                         | `PVP_LEVEL_RES_512_L3`  | Counter indicating the percentage of residency during the last 2 ms measurement for level 3 of AVX512 CPU instructions      |
|                         | `PVP_LEVEL_RES_TMUL_L0` | Counter indicating the percentage of residency during the last 2 ms measurement for level 0 of TMUL CPU instructions        |
|                         | `PVP_LEVEL_RES_TMUL_L1` | Counter indicating the percentage of residency during the last 2 ms measurement for level 1 of TMUL CPU instructions        |
|                         | `PVP_LEVEL_RES_TMUL_L2` | Counter indicating the percentage of residency during the last 2 ms measurement for level 2 of TMUL CPU instructions        |
|                         | `PVP_LEVEL_RES_TMUL_L3` | Counter indicating the percentage of residency during the last 2 ms measurement for level 3 of TMUL CPU instructions        |
| `ttsc_timer`            | `TSC_TIMER`             | OOBMSM TSC (Time Stamp Counter) value                                                                                       |
| `tnum_en_cha`           | `NUM_EN_CHA`            | Number of enabled CHAs                                                                                                      |
| `trmid_usage_counter`   | `RMID0_RDT_CMT`         | CHA x RMID 0 LLC cache line usage counter (per CHA)                                                                         |
|                         | `RMID1_RDT_CMT`         | CHA x RMID 1 LLC cache line usage counter (per CHA)                                                                         |
|                         | `RMID2_RDT_CMT`         | CHA x RMID 2 LLC cache line usage counter (per CHA)                                                                         |
|                         | `RMID3_RDT_CMT`         | CHA x RMID 3 LLC cache line usage counter (per CHA)                                                                         |
|                         | `RMID4_RDT_CMT`         | CHA x RMID 4 LLC cache line usage counter (per CHA)                                                                         |
|                         | `RMID5_RDT_CMT`         | CHA x RMID 5 LLC cache line usage counter (per CHA)                                                                         |
|                         | `RMID6_RDT_CMT`         | CHA x RMID 6 LLC cache line usage counter (per CHA)                                                                         |
|                         | `RMID7_RDT_CMT`         | CHA x RMID 7 LLC cache line usage counter (per CHA)                                                                         |
|                         | `RMID0_RDT_MBM_TOTAL`   | CHA x RMID 0 total memory transactions counter (per CHA)                                                                    |
|                         | `RMID1_RDT_MBM_TOTAL`   | CHA x RMID 1 total memory transactions counter (per CHA)                                                                    |
|                         | `RMID2_RDT_MBM_TOTAL`   | CHA x RMID 2 total memory transactions counter (per CHA)                                                                    |
|                         | `RMID3_RDT_MBM_TOTAL`   | CHA x RMID 3 total memory transactions counter (per CHA)                                                                    |
|                         | `RMID4_RDT_MBM_TOTAL`   | CHA x RMID 4 total memory transactions counter (per CHA)                                                                    |
|                         | `RMID5_RDT_MBM_TOTAL`   | CHA x RMID 5 total memory transactions counter (per CHA)                                                                    |
|                         | `RMID6_RDT_MBM_TOTAL`   | CHA x RMID 6 total memory transactions counter (per CHA)                                                                    |
|                         | `RMID7_RDT_MBM_TOTAL`   | CHA x RMID 7 total memory transactions counter (per CHA)                                                                    |
|                         | `RMID0_RDT_MBM_LOCAL`   | CHA x RMID 0 local memory transactions counter (per CHA)                                                                    |
|                         | `RMID1_RDT_MBM_LOCAL`   | CHA x RMID 1 local memory transactions counter (per CHA)                                                                    |
|                         | `RMID2_RDT_MBM_LOCAL`   | CHA x RMID 2 local memory transactions counter (per CHA)                                                                    |
|                         | `RMID3_RDT_MBM_LOCAL`   | CHA x RMID 3 local memory transactions counter (per CHA)                                                                    |
|                         | `RMID4_RDT_MBM_LOCAL`   | CHA x RMID 4 local memory transactions counter (per CHA)                                                                    |
|                         | `RMID5_RDT_MBM_LOCAL`   | CHA x RMID 5 local memory transactions counter (per CHA)                                                                    |
|                         | `RMID6_RDT_MBM_LOCAL`   | CHA x RMID 6 local memory transactions counter (per CHA)                                                                    |
|                         | `RMID7_RDT_MBM_LOCAL`   | CHA x RMID 7 local memory transactions counter (per CHA)                                                                    |
| `ttw_unit`              | `TW`                    | Time window. Valid TW range is 0 to 17. The unit is calculated as `2.3 * 2^TW` ms (e.g. `2.3 * 2^17` ms = ~302 seconds).    |
| `tcore_stress_level`    | `STRESS_LEVEL`          | Accumulating counter indicating relative stress level for a core (per core)                                                 |

### Example: C-State residency and temperature with a datatype metric filter

This configuration allows getting only a subset of metrics
with the use of a datatype filter:

```toml
[[inputs.intel_pmt]]
  spec = "/home/telegraf/Intel-PMT/xml/pmt.xml"
  datatypes_enabled = ["tbandwidth_28b","ttemperature"]
```

### Example: C-State residency and temperature with a sample metric filter

This configuration allows getting only a subset of metrics
with the use of a sample filter:

```toml
[[inputs.intel_pmt]]
  spec = "/home/telegraf/Intel-PMT/xml/pmt.xml"
  samples_enabled = ["C0Residency","C1Residency", "Cx_TEMP"]
```

## Metrics

All metrics have the following tags:

- `guid` (unique id of an Intel PMT space).
- `numa_node` (NUMA node the sample is collected from).
- `pci_bdf` (PCI Bus:Device.Function (BDF) the sample is collected from).
- `sample_name` (name of the gathered sample).
- `sample_group` (name of a group to which the sample belongs).
- `datatype_idref` (datatype to which the sample belongs).

`sample_name` prefixed in XMLs with `Cx_` where `x`
is the core number also have the following tag:

- `core` (core to which the metric relates).

`sample_name` prefixed in XMLs with `CHAx_` where `x`
is the CHA number also have the following tag:

- `cha` (Caching and Home Agent to which the metric relates).

## Example Output

Example output with `tpvp_throttle_counter` as a datatype metric filter:

```text
intel_pmt,core=0,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C0_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1886465i 1693766334000000000
intel_pmt,core=1,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C1_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=2,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C2_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=3,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C3_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=4,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C4_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1357578i 1693766334000000000
intel_pmt,core=5,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C5_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=6,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C6_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2024801i 1693766334000000000
intel_pmt,core=7,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C7_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=8,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C8_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1390741i 1693766334000000000
intel_pmt,core=9,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C9_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=10,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C10_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1536483i 1693766334000000000
intel_pmt,core=11,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C11_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=12,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C12_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=13,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C13_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=14,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C14_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1604964i 1693766334000000000
intel_pmt,core=15,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C15_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1168673i 1693766334000000000
intel_pmt,core=16,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C16_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=17,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C17_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=18,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C18_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1276588i 1693766334000000000
intel_pmt,core=19,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C19_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1139005i 1693766334000000000
intel_pmt,core=20,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C20_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=21,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C21_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=22,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C22_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=970698i 1693766334000000000
intel_pmt,core=23,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C23_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=24,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C24_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=25,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C25_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1178462i 1693766334000000000
intel_pmt,core=26,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C26_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=27,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C27_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2093384i 1693766334000000000
intel_pmt,core=28,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C28_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=29,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C29_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=30,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C30_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=31,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C31_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=32,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C32_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2825174i 1693766334000000000
intel_pmt,core=33,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C33_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2592279i 1693766334000000000
intel_pmt,core=34,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C34_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=35,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C35_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=36,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C36_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1960662i 1693766334000000000
intel_pmt,core=37,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C37_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1821914i 1693766334000000000
intel_pmt,core=38,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C38_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=39,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C39_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=40,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C40_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=41,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C41_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2654651i 1693766334000000000
intel_pmt,core=42,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C42_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2230984i 1693766334000000000
intel_pmt,core=43,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C43_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=44,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C44_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=45,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C45_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=46,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C46_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2325520i 1693766334000000000
intel_pmt,core=47,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C47_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=48,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C48_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=49,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C49_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=50,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C50_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=51,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C51_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=52,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C52_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1468880i 1693766334000000000
intel_pmt,core=53,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C53_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2151919i 1693766334000000000
intel_pmt,core=54,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C54_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=55,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C55_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=2065994i 1693766334000000000
intel_pmt,core=56,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C56_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=57,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C57_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=58,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C58_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1553691i 1693766334000000000
intel_pmt,core=59,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C59_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=1624177i 1693766334000000000
intel_pmt,core=60,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C60_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=61,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C61_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=62,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C62_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=63,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C63_PVP_THROTTLE_64,sample_name=PVP_THROTTLE_64 value=0i 1693766334000000000
intel_pmt,core=0,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C0_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=12977949i 1693766334000000000
intel_pmt,core=1,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C1_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=2,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C2_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=3,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C3_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=4,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C4_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=7180524i 1693766334000000000
intel_pmt,core=5,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C5_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=6,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C6_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=8667263i 1693766334000000000
intel_pmt,core=7,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C7_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=8,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C8_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=5945851i 1693766334000000000
intel_pmt,core=9,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C9_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=10,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C10_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=6669829i 1693766334000000000
intel_pmt,core=11,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C11_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=12,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C12_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=13,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C13_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=14,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C14_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=6579832i 1693766334000000000
intel_pmt,core=15,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C15_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=6101856i 1693766334000000000
intel_pmt,core=16,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C16_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=17,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C17_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=18,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C18_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=7796183i 1693766334000000000
intel_pmt,core=19,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C19_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=6849098i 1693766334000000000
intel_pmt,core=20,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C20_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=21,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C21_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=22,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C22_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=12378942i 1693766334000000000
intel_pmt,core=23,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C23_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=24,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C24_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=25,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C25_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=8299231i 1693766334000000000
intel_pmt,core=26,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C26_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=27,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C27_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=7986390i 1693766334000000000
intel_pmt,core=28,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C28_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=29,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C29_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=30,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C30_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=31,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C31_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=32,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C32_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=9876325i 1693766334000000000
intel_pmt,core=33,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C33_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=8547471i 1693766334000000000
intel_pmt,core=34,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C34_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=35,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C35_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=36,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C36_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=9231744i 1693766334000000000
intel_pmt,core=37,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C37_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=8133031i 1693766334000000000
intel_pmt,core=38,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C38_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=39,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C39_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=40,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C40_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=41,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C41_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=6136417i 1693766334000000000
intel_pmt,core=42,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C42_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=6091019i 1693766334000000000
intel_pmt,core=43,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C43_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=44,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C44_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=45,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C45_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=46,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C46_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=5804639i 1693766334000000000
intel_pmt,core=47,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C47_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=48,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C48_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=49,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C49_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=50,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C50_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=51,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C51_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=52,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C52_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=5738491i 1693766334000000000
intel_pmt,core=53,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C53_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=6058504i 1693766334000000000
intel_pmt,core=54,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C54_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=55,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C55_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=5987093i 1693766334000000000
intel_pmt,core=56,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C56_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=57,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C57_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=58,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C58_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=10384909i 1693766334000000000
intel_pmt,core=59,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C59_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=7305786i 1693766334000000000
intel_pmt,core=60,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C60_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=61,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C61_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=62,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C62_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
intel_pmt,core=63,datatype_idref=tpvp_throttle_counter,guid=0x87b6fef1,pmt,numa_node=0,pci_bdf=0000:e7:03.1,sample_group=C63_PVP_THROTTLE_1024,sample_name=PVP_THROTTLE_1024 value=0i 1693766334000000000
```
