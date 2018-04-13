# `nvidia-smi` Input Plugin

This plugin uses a query on the `nvidia-smi` binary to pull GPU stats including memory and GPU usage, temp and other.

```
nvidia-smi --query-gpu=fan.speed,memory.total,memory.used,memory.free,pstate,temperature.gpu,name,uuid,compute_mode,utilization.gpu,utilization.memory,index --id={{ gpu_index }} --format=csv,noheader,nounits
```
