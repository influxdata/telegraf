# `nvidia-smi` Input Plugin

This plugin uses a query on the [`nvidia-smi`](https://developer.nvidia.com/nvidia-system-management-interface) binary to pull GPU stats including memory and GPU usage, temp and other.

### Configuration

```toml
# Pulls statistics from nvidia GPUs attached to the host
[[inputs.nvidia_smi]]
## Optional: path to nvidia-smi binary, defaults to $PATH via exec.LookPath
# bin_path = /usr/bin/nvidia-smi

## Optional: timeout for GPU polling
# timeout = 5s
```

### Windows Users
Uncomment the bin_path line above and use the following syntax

```
bin_path = "C:/Program Files/NVIDIA Corporation/NVSMI/nvidia-smi.exe"
```

### Metrics
- measurement: `nvidia_smi`
  - tags
    - `name` (type of GPU e.g. `GeForce GTX 1070 Ti`)
    - `compute_mode` (The compute mode of the GPU e.g. `Default`)
    - `index` (The port index where the GPU is connected to the motherboard e.g. `1`)
    - `pstate` (Overclocking state for the GPU e.g. `P0`)
    - `uuid` (A unique identifier for the GPU e.g. `GPU-f9ba66fc-a7f5-94c5-da19-019ef2f9c665`)
  - fields
    - `fan_speed` (integer, percentage)
    - `memory_free` (integer, MiB)
    - `memory_used` (integer, MiB)
    - `memory_total` (integer, MiB)
    - `power_draw` (float, W)
    - `temperature_gpu` (integer, degrees C)
    - `utilization_gpu` (integer, percentage)
    - `utilization_memory` (integer, percentage)

### Sample Query

The below query could be used to alert on the average temperature of the your GPUs over the last minute

```
SELECT mean("temperature_gpu") FROM "nvidia_smi" WHERE time > now() - 5m GROUP BY time(1m), "index", "name", "host"
```

### Example Output

Windows Powershell
```
PS C:\tools\telegraf> .\telegraf.exe -config .\telegraf.conf -test
> nvidia_smi,compute_mode=Default,host=8218cf,index=0,name=GeForce\ GTX\ 1070,pstate=P2,uuid=GPU-dfb42064-6911-51eb-7785-053d17566149 fan_speed=70i,memory_free=6972i,memory_total=8192i,memory_used=1220i,temperature_gpu=75i,utilization_gpu=100i,utilization_memory=86i 1531184754000000000
> nvidia_smi,compute_mode=Default,host=8218cf,index=1,name=GeForce\ GTX\ 1080\ Ti,pstate=P2,uuid=GPU-1030998b-7303-9289-1d57-7b0a2c44e29a fan_speed=53i,memory_free=10303i,memory_total=11264i,memory_used=961i,temperature_gpu=65i,utilization_gpu=91i,utilization_memory=15i 1531184754000000000
```

Linux
```
nvidia_smi,compute_mode=Default,host=8218cf,index=0,name=GeForce\ GTX\ 1070,pstate=P2,uuid=GPU-823bc202-6279-6f2c-d729-868a30f14d96 fan_speed=100i,memory_free=7563i,memory_total=8112i,memory_used=549i,temperature_gpu=53i,utilization_gpu=100i,utilization_memory=90i 1523991122000000000
nvidia_smi,compute_mode=Default,host=8218cf,index=1,name=GeForce\ GTX\ 1080,pstate=P2,uuid=GPU-f9ba66fc-a7f5-94c5-da19-019ef2f9c665 fan_speed=100i,memory_free=7557i,memory_total=8114i,memory_used=557i,temperature_gpu=50i,utilization_gpu=100i,utilization_memory=85i 1523991122000000000
nvidia_smi,compute_mode=Default,host=8218cf,index=2,name=GeForce\ GTX\ 1080,pstate=P2,uuid=GPU-d4cfc28d-0481-8d07-b81a-ddfc63d74adf fan_speed=100i,memory_free=7557i,memory_total=8114i,memory_used=557i,temperature_gpu=58i,utilization_gpu=100i,utilization_memory=86i 1523991122000000000
```
