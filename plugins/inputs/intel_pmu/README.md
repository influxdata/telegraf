# Intel Performance Monitoring Unit Plugin

This input plugin exposes Intel PMU (Performance Monitoring Unit) metrics available through [Linux Perf](https://perf.wiki.kernel.org/index.php/Main_Page) subsystem.

PMU metrics gives insight into performance and health of IA processor's internal components,
including core and uncore units. With the number of cores increasing and processor topology getting more complex
the insight into those metrics is vital to assure the best CPU performance and utilization.

Performance counters are CPU hardware registers that count hardware events such as instructions executed, cache-misses suffered, or branches mispredicted.
They form a basis for profiling applications to trace dynamic control flow and identify hotspots.

## Configuration

```toml
# Intel Performance Monitoring Unit plugin exposes Intel PMU metrics available through Linux Perf subsystem
[[inputs.intel_pmu]]
  ## List of filesystem locations of JSON files that contain PMU event definitions.
  event_definitions = ["/var/cache/pmu/GenuineIntel-6-55-4-core.json", "/var/cache/pmu/GenuineIntel-6-55-4-uncore.json"]
  
  ## List of core events measurement entities. There can be more than one core_events sections.
  [[inputs.intel_pmu.core_events]]
    ## List of events to be counted. Event names shall match names from event_definitions files.
    ## Single entry can contain name of the event (case insensitive) augmented with config options and perf modifiers.
    ## If absent, all core events from provided event_definitions are counted skipping unresolvable ones.
    events = ["INST_RETIRED.ANY", "CPU_CLK_UNHALTED.THREAD_ANY:config1=0x4043200000000k"]

    ## Limits the counting of events to core numbers specified.
    ## If absent, events are counted on all cores.
    ## Single "0", multiple "0,1,2" and range "0-2" notation is supported for each array element.
    ##   example: cores = ["0,2", "4", "12-16"]
    cores = ["0"]

    ## Indicator that plugin shall attempt to run core_events.events as a single perf group.
    ## If absent or set to false, each event is counted individually. Defaults to false.
    ## This limits the number of events that can be measured to a maximum of available hardware counters per core.
    ## Could vary depending on type of event, use of fixed counters.
    # perf_group = false

    ## Optionally set a custom tag value that will be added to every measurement within this events group.
    ## Can be applied to any group of events, unrelated to perf_group setting.
    # events_tag = ""

  ## List of uncore event measurement entities. There can be more than one uncore_events sections.
  [[inputs.intel_pmu.uncore_events]]
    ## List of events to be counted. Event names shall match names from event_definitions files.
    ## Single entry can contain name of the event (case insensitive) augmented with config options and perf modifiers.
    ## If absent, all uncore events from provided event_definitions are counted skipping unresolvable ones.
    events = ["UNC_CHA_CLOCKTICKS", "UNC_CHA_TOR_OCCUPANCY.IA_MISS"]

    ## Limits the counting of events to specified sockets.
    ## If absent, events are counted on all sockets.
    ## Single "0", multiple "0,1" and range "0-1" notation is supported for each array element.
    ##   example: sockets = ["0-2"]
    sockets = ["0"]

    ## Indicator that plugin shall provide an aggregated value for multiple units of same type distributed in an uncore.
    ## If absent or set to false, events for each unit are exposed as separate metric. Defaults to false.
    # aggregate_uncore_units = false

    ## Optionally set a custom tag value that will be added to every measurement within this events group.
    # events_tag = ""
```

### Modifiers

Perf modifiers adjust event-specific perf attribute to fulfill particular requirements.
Details about perf attribute structure could be found in [perf_event_open](https://man7.org/linux/man-pages/man2/perf_event_open.2.html) syscall manual.

General schema of configuration's `events` list element:

```regexp
EVENT_NAME(:(config|config1|config2)=(0x[0-9a-f]{1-16})(p|k|u|h|H|I|G|D))*
```

where:

| Modifier | Underlying attribute            | Description                 |
|----------|---------------------------------|-----------------------------|
| config   | perf_event_attr.config          | type-specific configuration |
| config1  | perf_event_attr.config1         | extension of config         |
| config2  | perf_event_attr.config2         | extension of config1        |
| p        | perf_event_attr.precise_ip      | skid constraint             |
| k        | perf_event_attr.exclude_user    | don't count user            |
| u        | perf_event_attr.exclude_kernel  | don't count kernel          |
| h / H    | perf_event_attr.exclude_guest   | don't count in guest        |
| I        | perf_event_attr.exclude_idle    | don't count when idle       |
| G        | perf_event_attr.exclude_hv      | don't count hypervisor      |
| D        | perf_event_attr.pinned          | must always be on PMU       |

## Requirements

The plugin is using [iaevents](https://github.com/intel/iaevents) library which is a golang package that makes accessing the Linux kernel's perf interface easier.

Intel PMU plugin, is only intended for use on **linux 64-bit** systems.

Event definition JSON files for specific architectures can be found at [01.org](https://download.01.org/perfmon/).
A script to download the event definitions that are appropriate for your system (event_download.py) is available at [pmu-tools](https://github.com/andikleen/pmu-tools).
Please keep these files in a safe place on your system.

## Measuring

Plugin allows measuring both core and uncore events. During plugin initialization the event names provided by user are compared
with event definitions included in JSON files and translated to perf attributes. Next, those events are activated to start counting.
During every telegraf interval, the plugin reads proper measurement for each previously activated event.

Each single core event may be counted severally on every available CPU's core. In contrast, uncore events could be placed in
many PMUs within specified CPU package. The plugin allows choosing core ids (core events) or socket ids (uncore events) on which the counting should be executed.
Uncore events are separately activated on all socket's PMUs, and can be exposed as separate
measurement or to be summed up as one measurement.

Obtained measurements are stored as three values: **Raw**, **Enabled** and **Running**. Raw is a total count of event. Enabled and running are total time the event was enabled and running.
Normally these are the same. If more events are started than available counter slots on the PMU, then multiplexing
occurs and events only run part of the time. Therefore, the plugin provides a 4-th value called **scaled** which is calculated using following formula:
`raw * enabled / running`.

Events are measured for all running processes.

### Core event groups

Perf allows assembling events as a group. A perf event group is scheduled onto the CPU as a unit: it will be put onto the CPU only if all of the events in the group can be put onto the CPU.
This means that the values of the member events can be meaningfully compared — added, divided (to get ratios), and so on — with each other,
since they have counted events for the same set of executed instructions [(source)](https://man7.org/linux/man-pages/man2/perf_event_open.2.html).

> **NOTE:**
> Be aware that the plugin will throw an error when trying to create core event group of size that exceeds available core PMU counters.
> The error message from perf syscall will be shown as "invalid argument". If you want to check how many PMUs are supported by your Intel CPU, you can use the [cpuid](https://linux.die.net/man/1/cpuid) command.

### Note about file descriptors

The plugin opens a number of file descriptors dependent on number of monitored CPUs and number of monitored
counters. It can easily exceed the default per process limit of allowed file descriptors. Depending on
configuration, it might be required to increase the limit of opened file descriptors allowed.
This can be done for example by using `ulimit -n command`.

## Metrics

On each Telegraf interval, Intel PMU plugin transmits following data:

### Metric Fields

| Field   | Type   | Description                                                                                                                                   |
|---------|--------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| enabled | uint64 | time counter, contains time the associated perf event was enabled                                                                             |
| running | uint64 | time counter, contains time the event was actually counted                                                                                    |
| raw     | uint64 | value counter, contains event count value during the time the event was actually counted                                                      |
| scaled  | uint64 | value counter, contains approximated value of counter if the event was continuously counted, using scaled = raw * (enabled / running) formula |

### Metric Tags - common

| Tag   | Description                  |
|-------|------------------------------|
| host  | hostname as read by Telegraf |
| event | name of the event            |

### Metric Tags - core events

| Tag        | Description                                                                                        |
|------------|----------------------------------------------------------------------------------------------------|
| cpu        | CPU id as identified by linux OS (either logical cpu id when HT on or physical cpu id when HT off) |
| events_tag | (optional) tag as defined in "intel_pmu.core_events" configuration element                           |

### Metric Tags - uncore events

| Tag       | Description                                                                                                                                                                                |
|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| socket    | socket number as identified by linux OS (physical_package_id)                                                                                                                              |
| unit_type | type of event-capable PMU that the event was counted for, provides category of PMU that the event was counted for, e.g. cbox for uncore_cbox_1, r2pcie for uncore_r2pcie etc.              |
| unit      | name of event-capable PMU that the event was counted for, as listed in /sys/bus/event_source/devices/ e.g. uncore_cbox_1, uncore_imc_1 etc.  Present for non-aggregated uncore events only |
| events_tag| (optional) tag as defined in "intel_pmu.uncore_events" configuration element                           |

## Example outputs

Event group:

```text
pmu_metric,cpu=0,event=CPU_CLK_THREAD_UNHALTED.REF_XCLK,events_tag=unhalted,host=xyz enabled=2871237051i,running=2871237051i,raw=1171711i,scaled=1171711i 1621254096000000000
pmu_metric,cpu=0,event=CPU_CLK_UNHALTED.THREAD_P_ANY,events_tag=unhalted,host=xyz enabled=2871240713i,running=2871240713i,raw=72340716i,scaled=72340716i 1621254096000000000
pmu_metric,cpu=1,event=CPU_CLK_THREAD_UNHALTED.REF_XCLK,events_tag=unhalted,host=xyz enabled=2871118275i,running=2871118275i,raw=1646752i,scaled=1646752i 1621254096000000000
pmu_metric,cpu=1,event=CPU_CLK_UNHALTED.THREAD_P_ANY,events_tag=unhalted,host=xyz raw=108802421i,scaled=108802421i,enabled=2871120107i,running=2871120107i 1621254096000000000
pmu_metric,cpu=2,event=CPU_CLK_THREAD_UNHALTED.REF_XCLK,events_tag=unhalted,host=xyz enabled=2871143950i,running=2871143950i,raw=1316834i,scaled=1316834i 1621254096000000000
pmu_metric,cpu=2,event=CPU_CLK_UNHALTED.THREAD_P_ANY,events_tag=unhalted,host=xyz enabled=2871074681i,running=2871074681i,raw=68728436i,scaled=68728436i 1621254096000000000
```

Uncore event not aggregated:

```text
pmu_metric,event=UNC_CBO_XSNP_RESPONSE.MISS_XCORE,host=xyz,socket=0,unit=uncore_cbox_0,unit_type=cbox enabled=2870630747i,running=2870630747i,raw=183996i,scaled=183996i 1621254096000000000
pmu_metric,event=UNC_CBO_XSNP_RESPONSE.MISS_XCORE,host=xyz,socket=0,unit=uncore_cbox_1,unit_type=cbox enabled=2870608194i,running=2870608194i,raw=185703i,scaled=185703i 1621254096000000000
pmu_metric,event=UNC_CBO_XSNP_RESPONSE.MISS_XCORE,host=xyz,socket=0,unit=uncore_cbox_2,unit_type=cbox enabled=2870600211i,running=2870600211i,raw=187331i,scaled=187331i 1621254096000000000
pmu_metric,event=UNC_CBO_XSNP_RESPONSE.MISS_XCORE,host=xyz,socket=0,unit=uncore_cbox_3,unit_type=cbox enabled=2870593914i,running=2870593914i,raw=184228i,scaled=184228i 1621254096000000000
pmu_metric,event=UNC_CBO_XSNP_RESPONSE.MISS_XCORE,host=xyz,socket=0,unit=uncore_cbox_4,unit_type=cbox scaled=195355i,enabled=2870558952i,running=2870558952i,raw=195355i 1621254096000000000
pmu_metric,event=UNC_CBO_XSNP_RESPONSE.MISS_XCORE,host=xyz,socket=0,unit=uncore_cbox_5,unit_type=cbox enabled=2870554131i,running=2870554131i,raw=197756i,scaled=197756i 1621254096000000000
```

Uncore event aggregated:

```text
pmu_metric,event=UNC_CBO_XSNP_RESPONSE.MISS_XCORE,host=xyz,socket=0,unit_type=cbox enabled=13199712335i,running=13199712335i,raw=467485i,scaled=467485i 1621254412000000000 
```

Time multiplexing:

```text
pmu_metric,cpu=0,event=CPU_CLK_THREAD_UNHALTED.REF_XCLK,host=xyz raw=2947727i,scaled=4428970i,enabled=2201071844i,running=1464935978i 1621254412000000000
pmu_metric,cpu=0,event=CPU_CLK_UNHALTED.THREAD_P_ANY,host=xyz running=1465155618i,raw=302553190i,scaled=454511623i,enabled=2201035323i 1621254412000000000
pmu_metric,cpu=0,event=CPU_CLK_UNHALTED.REF_XCLK,host=xyz enabled=2200994057i,running=1466812391i,raw=3177535i,scaled=4767982i 1621254412000000000
pmu_metric,cpu=0,event=CPU_CLK_UNHALTED.REF_XCLK_ANY,host=xyz enabled=2200963921i,running=1470523496i,raw=3359272i,scaled=5027894i 1621254412000000000
pmu_metric,cpu=0,event=L1D_PEND_MISS.PENDING_CYCLES_ANY,host=xyz enabled=2200933946i,running=1470322480i,raw=23631950i,scaled=35374798i 1621254412000000000
pmu_metric,cpu=0,event=L1D_PEND_MISS.PENDING_CYCLES,host=xyz raw=18767833i,scaled=28169827i,enabled=2200888514i,running=1466317384i 1621254412000000000
```
