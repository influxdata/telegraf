# CPUFreq Input Plugin For Telegraf Agent

The CPUFreq plugin collects the current CPU's frequency. This plugin work only with linux.

## Configuration

```toml
[[inputs.cpufreq]]
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # path_sysfs = "/sys"
  ## Gather CPU throttles per socker
  ## Defaults:
  # throttles_per_socket = false
  ## Gather CPU throttles per physical core
  ## Defaults:
  # throttles_per_core = false
```

## Example output

```
> cpufreq,cpu=0,host=server01 cur_freq=3756293000,max_freq=3900000000,min_freq=800000000 1527789803000000000
> cpufreq,cpu=1,host=server01 cur_freq=3735119000,max_freq=3900000000,min_freq=800000000 1527789803000000000
> cpufreq,cpu=2,host=server01 cur_freq=3786381000,max_freq=3900000000,min_freq=800000000 1527789803000000000
> cpufreq,cpu=3,host=server01 cur_freq=3823190000,max_freq=3900000000,min_freq=800000000 1527789803000000000
> cpufreq,cpu=4,host=server01 cur_freq=3780804000,max_freq=3900000000,min_freq=800000000 1527789803000000000
> cpufreq,cpu=5,host=server01 cur_freq=3801758000,max_freq=3900000000,min_freq=800000000 1527789803000000000
> cpufreq,cpu=6,host=server01 cur_freq=3839194000,max_freq=3900000000,min_freq=800000000 1527789803000000000
> cpufreq,cpu=7,host=server01 cur_freq=3877989000,max_freq=3900000000,min_freq=800000000 1527789803000000000
> cpufreq_cpu_throttles,cpu=0,host=server01 count=0 1527789803000000000
> cpufreq_core_throttles,core=0,cpu=0,host=server01 count=0 1527789803000000000
> cpufreq_core_throttles,core=1,cpu=0,host=server01 count=0 1527789803000000000
> cpufreq_core_throttles,core=2,cpu=0,host=server01 count=0 1527789803000000000
> cpufreq_core_throttles,core=3,cpu=0,host=server01 count=0 1527789803000000000
```
