# `nvidia-smi` Input Plugin

This plugin uses a query on the [`nvidia-smi`](https://developer.nvidia.com/nvidia-system-management-interface) binary to pull GPU stats including memory and GPU usage, temp and other.

```
nvidia-smi --query-gpu=fan.speed,memory.total,memory.used,memory.free,pstate,temperature.gpu,name,uuid,compute_mode,utilization.gpu,utilization.memory,index --id={{ gpu_index }} --format=csv,noheader,nounits
```

#### Tags

- `name`: The type of GPU e.g. `GeForce GTX 170 Ti`
- `compute_mode`: The compute mode of the GPU e.g. `Default`
- `index`: The port index where the GPU is connected to the motherboard e.g. `1`
- `pstate`: Overclocking state for the GPU e.g. `P0`
- `uuid`: A unique identifier for the GPU e.g. `GPU-f9ba66fc-a7f5-94c5-da19-019ef2f9c665`

#### Fields

- `fan_speed`: Fan speed as a percentage of max speed e.g. `80`
- `memory_free`: Amount of memory the GPU has free in kb e.g. `7650`
- `memory_used`:  Amount of memory the GPU is using in kb e.g. `550`
- `memory_total`:  Amount of memory the GPU is using in kb e.g. `8110`
- `temperature_gpu`: The temperature of the GPU in Degrees C
- `utilization_gpu`: The % utilization of the GPU
- `utilization_memory`: The % utilization of the GPU memory
