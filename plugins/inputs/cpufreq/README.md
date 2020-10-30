# CPU Freqency Input Plugin

The `cpufreq` plugin gather metrics on the system CPU's frequency.
This plugin using linux kernel filesystems.

#### Configuration
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

### Metrics

- cpufreq
  - tags:
    - cpu (CPU ID or `cpu-total`)
  - fields:
    - min_freq (float)
    - max_freq (float)
    - cur_freq (float)

### Example Output

```
> cpufreq,cpu=0,host=Z370M-DS3H cur_freq=1382522000i,max_freq=4900000000i,min_freq=800000000i 1604049750000000000
> cpufreq,cpu=1,host=Z370M-DS3H cur_freq=1094884000i,max_freq=4900000000i,min_freq=800000000i 1604049750000000000
> cpufreq,cpu=2,host=Z370M-DS3H cur_freq=1010482000i,max_freq=4900000000i,min_freq=800000000i 1604049750000000000
> cpufreq,cpu=3,host=Z370M-DS3H cur_freq=2089249000i,max_freq=4900000000i,min_freq=800000000i 1604049750000000000
> cpufreq,cpu=4,host=Z370M-DS3H cur_freq=1272475000i,max_freq=4900000000i,min_freq=800000000i 1604049750000000000
> cpufreq,cpu=5,host=Z370M-DS3H cur_freq=1374903000i,max_freq=4900000000i,min_freq=800000000i 1604049750000000000
> cpufreq,cpu=6,host=Z370M-DS3H cur_freq=1355753000i,max_freq=4900000000i,min_freq=800000000i 1604049750000000000
> cpufreq,cpu=7,host=Z370M-DS3H cur_freq=1153656000i,max_freq=4900000000i,min_freq=800000000i 1604049750000000000
```
