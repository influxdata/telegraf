# CPUFreq Input Plugin For Telegraf Agent

The CPUFreq plugin collects the current CPU's frequency. This plugin work only with linux.

## Configuration

```toml
[[inputs.cpufreq]]
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # host_sys = "/sys"
  ## Gather CPU throttles per core
  ## Defaults:
  # gather_throttles = false
```

## Example output

```
> cpufreq,cpu=0,host=server01 cur_freq=3756293000i,max_freq=3900000000i,min_freq=800000000i,throttle_count=0i 1527789803000000000
> cpufreq,cpu=1,host=server01 cur_freq=3735119000i,max_freq=3900000000i,min_freq=800000000i,throttle_count=0i 1527789803000000000
> cpufreq,cpu=2,host=server01 cur_freq=3786381000i,max_freq=3900000000i,min_freq=800000000i,throttle_count=0i 1527789803000000000
> cpufreq,cpu=3,host=server01 cur_freq=3823190000i,max_freq=3900000000i,min_freq=800000000i,throttle_count=0i 1527789803000000000
> cpufreq,cpu=4,host=server01 cur_freq=3780804000i,max_freq=3900000000i,min_freq=800000000i,throttle_count=0i 1527789803000000000
> cpufreq,cpu=5,host=server01 cur_freq=3801758000i,max_freq=3900000000i,min_freq=800000000i,throttle_count=0i 1527789803000000000
> cpufreq,cpu=6,host=server01 cur_freq=3839194000i,max_freq=3900000000i,min_freq=800000000i,throttle_count=0i 1527789803000000000
> cpufreq,cpu=7,host=server01 cur_freq=3877989000i,max_freq=3900000000i,min_freq=800000000i,throttle_count=0i 1527789803000000000
```
