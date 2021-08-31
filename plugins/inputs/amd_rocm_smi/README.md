# ROCm System Management Interface (SMI) Input Plugin

This plugin uses a query on the [`rocm-smi`](https://github.com/RadeonOpenCompute/rocm_smi_lib/tree/master/python_smi_tools) binary to pull GPU stats including memory and GPU usage, temperatures and other.

### Configuration

```toml
# Pulls statistics from nvidia GPUs attached to the host
[[inputs.amd_rocm_smi]]
  ## Optional: path to rocm-smi binary, defaults to $PATH via exec.LookPath
  # bin_path = "/opt/rocm/bin/rocm-smi"

  ## Optional: timeout for GPU polling
  # timeout = "5s"
```

### Metrics
- measurement: `amd_rocm_smi`
  - tags
    - `name` (entry name assigned by rocm-smi executable)
    - `gpu_id` (id of the GPU according to rocm-smi)
    - `gpu_unique_id` (unique id of the GPU)

  - fields
    - `driver_version` (integer)
    - `fan_speed`(integer)
    - `memory_total`(integer B)
    - `memory_used`(integer B)
    - `memory_free`(integer B)
    - `temperature_sensor_edge` (float, Celsius)
    - `temperature_sensor_junction` (float, Celsius)
    - `temperature_sensor_memory` (float, Celsius)
    - `utilization_gpu` (integer, percentage)
    - `utilization_memory` (integer, percentage)
    - `clocks_current_sm` (integer, Mhz)
    - `clocks_current_memory` (integer, Mhz)
    - `power_draw` (float, Watt)

### Troubleshooting
Check the full output by running `rocm-smi` binary manually.

Linux:
```sh
rocm-smi rocm-smi -o -l -m -M  -g -c -t -u -i -f -p -P -s -S -v --showreplaycount --showpids --showdriverversion --showmemvendor --showfwinfo --showproductname --showserial --showuniqueid --showbus --showpendingpages --showpagesinfo --showretiredpages --showunreservablepages --showmemuse --showvoltage --showtopo --showtopoweight --showtopohops --showtopotype --showtoponuma --showmeminfo all --json
```
Please include the output of this command if opening a GitHub issue, together with ROCm version.

### Limitations and notices
Please notice that this plugin has been developed and tested on a limited number of versions and small set of GPUs. Currently the latest ROCm version tested is 4.3.0.
Notice that depending on the device and driver versions the amount of information provided by `rocm-smi` can vary so that some fields would start/stop appearing in the metrics upon updates.
The `rocm-smi` JSON output is not perfectly homogeneous and is possibly changing in the future, hence parsing and unmarshaling can start failing upon updating ROCm.

Inspired by the current state of the art of the `nvidia-smi` plugin.
