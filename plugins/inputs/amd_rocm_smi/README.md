# AMD ROCm System Management Interface (SMI) Input Plugin

This plugin gathers statistics including memory and GPU usage, temperatures
etc from [AMD ROCm platform][amd_rocm] GPUs.

> [!IMPORTANT]
> The [`rocm-smi` binary][binary] is required and needs to be installed on the
> system.

⭐ Telegraf v1.20.0
🏷️ hardware, system
💻 all

[amd_rocm]: https://rocm.docs.amd.com/
[binary]: https://github.com/RadeonOpenCompute/rocm_smi_lib/tree/master/python_smi_tools

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Startup error behavior options

In addition to the plugin-specific and global configuration settings the plugin
supports options for specifying the behavior when experiencing startup errors
using the `startup_error_behavior` setting. Available values are:

- `error`:  Telegraf with stop and exit in case of startup errors. This is the
            default behavior.
- `ignore`: Telegraf will ignore startup errors for this plugin and disables it
            but continues processing for all other plugins.
- `retry`:  NOT AVAILABLE

## Configuration

```toml @sample.conf
# Query statistics from AMD Graphics cards using rocm-smi binary
[[inputs.amd_rocm_smi]]
  ## Optional: path to rocm-smi binary, defaults to $PATH via exec.LookPath
  # bin_path = "/opt/rocm/bin/rocm-smi"

  ## Optional: timeout for GPU polling
  # timeout = "5s"
```

## Metrics

- measurement: `amd_rocm_smi`
  - tags
    - `name` (entry name assigned by rocm-smi executable)
    - `gpu_id` (id of the GPU according to rocm-smi)
    - `gpu_unique_id` (unique id of the GPU)

  - fields
    - `driver_version` (integer)
    - `fan_speed` (integer)
    - `memory_total` (integer, B)
    - `memory_used` (integer, B)
    - `memory_free` (integer, B)
    - `temperature_sensor_edge` (float, Celsius)
    - `temperature_sensor_junction` (float, Celsius)
    - `temperature_sensor_memory` (float, Celsius)
    - `utilization_gpu` (integer, percentage)
    - `utilization_memory` (integer, percentage)
    - `clocks_current_sm` (integer, Mhz)
    - `clocks_current_memory` (integer, Mhz)
    - `clocks_current_display` (integer, Mhz)
    - `clocks_current_fabric` (integer, Mhz)
    - `clocks_current_system` (integer, Mhz)
    - `power_draw` (float, Watt)
    - `card_series` (string)
    - `card_model` (string)
    - `card_vendor` (string)

## Troubleshooting

Check the full output by running `rocm-smi` binary manually.

Linux:

```sh
rocm-smi rocm-smi -o -l -m -M  -g -c -t -u -i -f -p -P -s -S -v --showreplaycount --showpids --showdriverversion --showmemvendor --showfwinfo --showproductname --showserial --showuniqueid --showbus --showpendingpages --showpagesinfo --showretiredpages --showunreservablepages --showmemuse --showvoltage --showtopo --showtopoweight --showtopohops --showtopotype --showtoponuma --showmeminfo all --json
```

Please include the output of this command if opening a GitHub issue, together
with ROCm version.

## Example Output

```text
amd_rocm_smi,gpu_id=0x6861,gpu_unique_id=0x2150e7d042a1124,host=ali47xl,name=card0 clocks_current_memory=167i,clocks_current_sm=852i,driver_version=51114i,fan_speed=14i,memory_free=17145282560i,memory_total=17163091968i,memory_used=17809408i,power_draw=7,temperature_sensor_edge=28,temperature_sensor_junction=29,temperature_sensor_memory=92,utilization_gpu=0i 1630572551000000000
amd_rocm_smi,gpu_id=0x6861,gpu_unique_id=0x2150e7d042a1124,host=ali47xl,name=card0 clocks_current_memory=167i,clocks_current_sm=852i,driver_version=51114i,fan_speed=14i,memory_free=17145282560i,memory_total=17163091968i,memory_used=17809408i,power_draw=7,temperature_sensor_edge=29,temperature_sensor_junction=30,temperature_sensor_memory=91,utilization_gpu=0i 1630572701000000000
amd_rocm_smi,gpu_id=0x6861,gpu_unique_id=0x2150e7d042a1124,host=ali47xl,name=card0 clocks_current_memory=167i,clocks_current_sm=852i,driver_version=51114i,fan_speed=14i,memory_free=17145282560i,memory_total=17163091968i,memory_used=17809408i,power_draw=7,temperature_sensor_edge=29,temperature_sensor_junction=29,temperature_sensor_memory=92,utilization_gpu=0i 1630572749000000000
```

## Limitations and notices

Please notice that this plugin has been developed and tested on a limited number
of versions and small set of GPUs. Currently the latest ROCm version tested is
4.3.0.  Notice that depending on the device and driver versions the amount of
information provided by `rocm-smi` can vary so that some fields would start/stop
appearing in the metrics upon updates.  The `rocm-smi` JSON output is not
perfectly homogeneous and is possibly changing in the future, hence parsing and
unmarshalling can start failing upon updating ROCm.

Inspired by the current state of the art of the `nvidia-smi` plugin.
