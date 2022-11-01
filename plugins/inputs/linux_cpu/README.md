# Linux CPU Input Plugin

The `linux_cpu` plugin gathers CPU metrics exposed on Linux-based systems.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Collects CPU metrics exposed on Linux
[[inputs.linux_cpu]]
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # host_sys = "/sys"

  ## CPU metrics collected by the plugin.
  ## Supported options:
  ## "cpufreq", "thermal"
  ## Defaults:
  # metrics = ["cpufreq"]
```

## Metrics

The following tags are emitted by the plugin under the name `linux_cpu`:

| Tag   | Description           |
|-------|-----------------------|
| `cpu` | Identifier of the CPU |

The following fields are emitted by the plugin when selecting `cpufreq`:

| Metric name (field) | Description                                                | Units |
|---------------------|------------------------------------------------------------|-------|
| `scaling_cur_freq`  | Current frequency of the CPU as determined by CPUFreq      | KHz   |
| `scaling_min_freq`  | Minimum frequency the governor can scale to                | KHz   |
| `scaling_max_freq`  | Maximum frequency the governor can scale to                | KHz   |
| `cpuinfo_cur_freq`  | Current frequency of the CPU as determined by the hardware | KHz   |
| `cpuinfo_min_freq`  | Minimum operating frequency of the CPU                     | KHz   |
| `cpuinfo_max_freq`  | Maximum operating frequency of the CPU                     | KHz   |

The following fields are emitted by the plugin when selecting `thermal`:

| Metric name (field)   | Description                                                 | Units |
|-----------------------|-------------------------------------------------------------|-------|
| `throttle_count`      | Number of thermal throttle events reported by the CPU       |       |
| `throttle_max_time`   | Maximum amount of time CPU was in throttled state           | ms    |
| `throtlle_total_time` | Cumulative time during which the CPU was in throttled state | ms    |

## Example Output

```shell
> linux_cpu,cpu=0,host=go scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i,throttle_count=0i,throttle_max_time=0i,throttle_total_time=0i,scaling_cur_freq=803157i,scaling_min_freq=400000i 1617621150000000000
> linux_cpu,cpu=1,host=go throttle_total_time=0i,scaling_cur_freq=802939i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i,throttle_count=0i,throttle_max_time=0i 1617621150000000000
> linux_cpu,cpu=10,host=go throttle_max_time=0i,throttle_total_time=0i,scaling_cur_freq=838343i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i,throttle_count=0i 1617621150000000000
> linux_cpu,cpu=11,host=go cpuinfo_max_freq=4700000i,throttle_count=0i,throttle_max_time=0i,throttle_total_time=0i,scaling_cur_freq=800054i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i 1617621150000000000
> linux_cpu,cpu=2,host=go throttle_total_time=0i,scaling_cur_freq=800404i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i,throttle_count=0i,throttle_max_time=0i 1617621150000000000
> linux_cpu,cpu=3,host=go throttle_total_time=0i,scaling_cur_freq=800126i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i,throttle_count=0i,throttle_max_time=0i 1617621150000000000
> linux_cpu,cpu=4,host=go cpuinfo_max_freq=4700000i,throttle_count=0i,throttle_max_time=0i,throttle_total_time=0i,scaling_cur_freq=800359i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i 1617621150000000000
> linux_cpu,cpu=5,host=go throttle_max_time=0i,throttle_total_time=0i,scaling_cur_freq=800093i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i,throttle_count=0i 1617621150000000000
> linux_cpu,cpu=6,host=go cpuinfo_max_freq=4700000i,throttle_count=0i,throttle_max_time=0i,throttle_total_time=0i,scaling_cur_freq=741646i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i 1617621150000000000
> linux_cpu,cpu=7,host=go scaling_cur_freq=700006i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i,throttle_count=0i,throttle_max_time=0i,throttle_total_time=0i 1617621150000000000
> linux_cpu,cpu=8,host=go throttle_max_time=0i,throttle_total_time=0i,scaling_cur_freq=700046i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i,throttle_count=0i 1617621150000000000
> linux_cpu,cpu=9,host=go throttle_count=0i,throttle_max_time=0i,throttle_total_time=0i,scaling_cur_freq=700075i,scaling_min_freq=400000i,scaling_max_freq=4700000i,cpuinfo_min_freq=400000i,cpuinfo_max_freq=4700000i 1617621150000000000
```
